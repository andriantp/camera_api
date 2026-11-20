package user

import (
	"context"
	"time"

	"camera_api/pkg/adapter/database/model"

	"gorm.io/gorm"
)

type Session interface {
	UpdateSessionLastSeen(ctx context.Context, sessionUUID string, timestamp time.Time, ipAddress, userAgent string) error
	UpdateSessionLogout(ctx context.Context, sessionUUID string, timestamp time.Time, ipAddress, userAgent string) error
	CreateSession(ctx context.Context, session *model.Session) error
	FindActiveSession(ctx context.Context, sessionUUID, userUUID string) (*model.ActiveSession, error)
	RevokeUserSessions(ctx context.Context, userUUID string, timestamp time.Time, exceptSessionUUID string) error
}

// FindActiveSession returns gorm.ErrRecordNotFound when the session does not
// exist, belongs to another user, is logged out, or its user is inactive/deleted.
func (r *repo) FindActiveSession(ctx context.Context, sessionUUID, userUUID string) (*model.ActiveSession, error) {
	var data model.ActiveSession

	result := r.provider.Db(ctx).
		Table(model.Session{}.TableName()+" s").
		Select("r.name AS role_name").
		Joins("JOIN users u ON u.uuid = s.user_id").
		Joins("JOIN roles r ON r.uuid = u.role_id").
		Where("s.uuid = ? AND s.user_id = ?", sessionUUID, userUUID).
		Where("s.logout_at IS NULL").
		Where("u.active = ? AND u.deleted_at IS NULL", true).
		Limit(1).
		Scan(&data)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return &data, nil
}

// RevokeUserSessions logs out every open session of a user. When
// exceptSessionUUID is not empty, that session is kept (e.g. the caller's own).
func (r *repo) RevokeUserSessions(ctx context.Context, userUUID string, timestamp time.Time, exceptSessionUUID string) error {
	query := r.provider.Db(ctx).
		Model(&model.Session{}).
		Where("user_id = ? AND logout_at IS NULL", userUUID)

	if exceptSessionUUID != "" {
		query = query.Where("uuid <> ?", exceptSessionUUID)
	}

	return query.Update("logout_at", timestamp).Error
}

func (r *repo) UpdateSessionLastSeen(ctx context.Context, sessionUUID string, timestamp time.Time, ipAddress, userAgent string) error {
	result := r.provider.Db(ctx).
		Model(&model.Session{}).
		Where("uuid = ?", sessionUUID).
		Updates(map[string]interface{}{
			"last_seen_at": timestamp,
			"ip_address":   ipAddress,
			"user_agent":   userAgent,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *repo) UpdateSessionLogout(ctx context.Context, sessionUUID string, timestamp time.Time, ipAddress, userAgent string) error {
	result := r.provider.Db(ctx).
		Model(&model.Session{}).
		Where("uuid = ? AND logout_at IS NULL", sessionUUID).
		Updates(map[string]interface{}{
			"logout_at":    timestamp,
			"last_seen_at": timestamp,
			"ip_address":   ipAddress,
			"user_agent":   userAgent,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *repo) CreateSession(ctx context.Context, session *model.Session) error {
	return r.provider.Db(ctx).Create(session).Error
}
