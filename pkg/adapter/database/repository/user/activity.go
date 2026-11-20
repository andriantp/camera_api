package user

import (
	"context"
	"time"

	"camera_api/pkg/adapter/database/model"

	"gorm.io/gorm"
)

type ActivityTarget string

const (
	ActivityTargetAll    ActivityTarget = "all"
	ActivityTargetUser   ActivityTarget = "user"
	ActivityTargetCamera ActivityTarget = "camera"
)

type ActivityFilter struct {
	ActorUserID string
	Target      ActivityTarget
	From        time.Time
	To          time.Time
	Limit       int
	Offset      int
	Order       string
}

type Activity interface {
	FindUserActivities(ctx context.Context, filter ActivityFilter) ([]model.UserActivitySummary, error)
	CreateActivity(ctx context.Context, activity *model.UserActivity) error
}

func (r *repo) CreateActivity(ctx context.Context, activity *model.UserActivity) error {
	return r.provider.Db(ctx).Create(activity).Error
}

func (r *repo) FindUserActivities(ctx context.Context, filter ActivityFilter) ([]model.UserActivitySummary, error) {
	var activities []model.UserActivitySummary

	err := activityQuery(r.provider.Db(ctx), filter).
		Scan(&activities).
		Error

	if err != nil {
		return nil, err
	}

	return activities, nil
}

// activityQuery builds the activity listing query without executing it, so
// the generated SQL can be tested without a database.
func activityQuery(db *gorm.DB, filter ActivityFilter) *gorm.DB {
	query := db.
		Table("user_activities").
		Joins("LEFT JOIN users actor ON actor.uuid = user_activities.actor_user_id").
		Joins("LEFT JOIN users target ON target.uuid = user_activities.target_user_id").
		Joins("LEFT JOIN cameras ON cameras.uuid = user_activities.camera_id").
		Where("user_activities.actor_user_id = ?", filter.ActorUserID)

	// Zero From/To means no bound on that side.
	if !filter.From.IsZero() {
		query = query.Where("user_activities.created_at >= ?", filter.From)
	}

	if !filter.To.IsZero() {
		query = query.Where("user_activities.created_at <= ?", filter.To)
	}

	switch filter.Target {
	case ActivityTargetUser:
		query = query.Where("user_activities.target_user_id IS NOT NULL")
	case ActivityTargetCamera:
		query = query.Where("user_activities.camera_id IS NOT NULL")
	}

	order := "DESC"
	if filter.Order == "asc" {
		order = "ASC"
	}

	return query.
		Select(`
			user_activities.uuid,
			user_activities.created_at AS at,
			actor.username AS actor,
			user_activities.ip_address,
			user_activities.user_agent,
			user_activities.activity,
			target.username AS target,
			cameras.sn AS camera,
			user_activities.message,
			COUNT(*) OVER() AS total
		`).
		Order("user_activities.created_at " + order).
		Limit(filter.Limit).
		Offset(filter.Offset)
}
