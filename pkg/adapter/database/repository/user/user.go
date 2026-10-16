package user

import (
	"context"
	"time"

	"camera_api/pkg/adapter/database/model"

	"gorm.io/gorm"
)

type User interface {
	FindUserByUserName(ctx context.Context, username string) (*model.UserSession, error)
	FindUserByUserUUID(ctx context.Context, userUUID string) (*model.UserToken, error)
	FindPassword(ctx context.Context, userUUID string) (string, error)
	FindUserByUserUUIDForUpdate(ctx context.Context, userUUID string) (*model.User, error)
	FindUserSummaryByUserUUID(ctx context.Context, userUUID string) (*model.UserSummary, error)
	FindUsers(ctx context.Context, actorRoleID, limit, offset int) ([]model.UserSummary, error)

	UpdateUsername(ctx context.Context, userUUID, username string) error
	UpdatePassword(ctx context.Context, userUUID, password string) error
	UpdateLoginFailed(ctx context.Context, userUUID string, loginFailed int) error
	UpdateLastLogin(ctx context.Context, userUUID string, timestamp time.Time) error
	UpdateStatus(ctx context.Context, userUUID string, status bool) error

	CreateUser(ctx context.Context, user *model.User) error
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, userUUID string) error
}

func (r *repo) FindUserByUserName(ctx context.Context, username string) (*model.UserSession, error) {
	var data model.UserSession
	db := r.provider.Db(ctx)

	err := db.
		Table(model.User{}.TableName()+" u").
		Select(`
			u.username AS user_name,
			u.uuid AS user_uuid,
			u.password AS password,
			r.name AS role_name,
			r.uuid AS role_uuid,
			u.active AS active,
			u.last_login_at AS last_login_at,
			u.login_failed AS login_failed
		`).
		Joins("JOIN roles r ON r.uuid = u.role_id").
		Where("u.username = ? AND u.deleted_at IS NULL", username).
		Take(&data).
		Error

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (r *repo) FindUserByUserUUID(ctx context.Context, userUUID string) (*model.UserToken, error) {
	var data model.UserToken
	db := r.provider.Db(ctx)

	err := db.
		Table(model.User{}.TableName()+" u").
		Select(`
			u.username AS user_name,
			u.uuid AS user_uuid,
			r.name AS role_name,
			r.uuid AS role_uuid,
			u.active AS active,
			u.last_login_at AS last_login_at
		`).
		Joins("JOIN roles r ON r.uuid = u.role_id").
		Where("u.uuid = ? AND u.deleted_at IS NULL", userUUID).
		Take(&data).
		Error

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (r *repo) FindPassword(ctx context.Context, userUUID string) (string, error) {
	var user model.User

	err := r.provider.Db(ctx).
		Select("password").
		Where("uuid = ?", userUUID).
		First(&user).Error

	if err != nil {
		return "", err
	}

	return user.Password, nil
}

func (r *repo) FindUserByUserUUIDForUpdate(ctx context.Context, userUUID string) (*model.User, error) {
	db := r.provider.Db(ctx)
	var user model.User

	err := db.
		Where("uuid = ? AND deleted_at IS NULL", userUUID).
		First(&user).
		Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *repo) FindUserSummaryByUserUUID(ctx context.Context, userUUID string) (*model.UserSummary, error) {
	db := r.provider.Db(ctx)
	var data model.UserSummary

	result := db.
		Table("users u").
		Select(`
			u.uuid,
			u.username,
			r.id AS role_id,
			r.name AS role,
			u.last_login_at AS last_login,
			s.last_seen_at AS last_seen,
			u.active
		`).
		Joins("JOIN roles r ON r.uuid = u.role_id").
		Joins(`
			LEFT JOIN LATERAL (
				SELECT s.last_seen_at
				FROM sessions s
				WHERE s.user_id = u.uuid
					AND s.logout_at IS NULL
				ORDER BY s.last_seen_at DESC
				LIMIT 1
			) s ON true
		`).
		Where("u.uuid = ? AND u.deleted_at IS NULL", userUUID).
		Scan(&data)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &data, nil
}

func (r *repo) FindUsers(ctx context.Context, actorRoleID, limit, offset int) ([]model.UserSummary, error) {
	db := r.provider.Db(ctx)

	var users []model.UserSummary

	err := db.
		Table("users").
		Select(`
			users.uuid,
			users.username,
			roles.id AS role_id,
			roles.name AS role,
			users.last_login_at AS last_login,
			MAX(sessions.last_seen_at) AS last_seen,
			users.active,
			latest_activity.created_at AS last_activity,
			latest_activity.activity AS last_activity_type,
			latest_activity.message AS last_activity_message,
			COUNT(*) OVER() AS total
		`).
		Joins("JOIN roles ON roles.uuid = users.role_id").
		Joins("LEFT JOIN sessions ON sessions.user_id = users.uuid").
		Joins(`
			LEFT JOIN LATERAL (
				SELECT
					ua.created_at,
					ua.activity,
					ua.message
				FROM user_activities ua
				WHERE ua.actor_user_id = users.uuid
				ORDER BY ua.created_at DESC
				LIMIT 1
			) latest_activity ON true
		`).
		Where("users.deleted_at IS NULL").
		Where("roles.id >= ?", actorRoleID).
		Group(`
			users.id,
			users.uuid,
			users.username,
			roles.id,
			roles.name,
			users.last_login_at,
			users.active,
			latest_activity.created_at,
			latest_activity.activity,
			latest_activity.message
		`).
		Order("roles.id ASC").
		Order("users.id ASC").
		Limit(limit).
		Offset(offset).
		Scan(&users).
		Error

	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *repo) UpdateUsername(ctx context.Context, userUUID, username string) error {
	return r.provider.Db(ctx).Model(&model.User{}).Where("uuid = ?", userUUID).Update("username", username).Error
}

func (r *repo) UpdatePassword(ctx context.Context, userUUID, password string) error {
	return r.provider.Db(ctx).Model(&model.User{}).Where("uuid = ?", userUUID).Update("password", password).Error
}

func (r *repo) UpdateLoginFailed(ctx context.Context, userUUID string, loginFailed int) error {
	return r.provider.Db(ctx).Model(&model.User{}).Where("uuid = ?", userUUID).Update("login_failed", loginFailed).Error
}

func (r *repo) UpdateLastLogin(ctx context.Context, userUUID string, timestamp time.Time) error {
	return r.provider.Db(ctx).Model(&model.User{}).Where("uuid = ?", userUUID).Update("last_login_at", timestamp).Error
}

func (r *repo) UpdateStatus(ctx context.Context, userUUID string, status bool) error {
	return r.provider.Db(ctx).Model(&model.User{}).Where("uuid = ?", userUUID).Update("active", status).Error
}

func (r *repo) CreateUser(ctx context.Context, user *model.User) error {
	return r.provider.Db(ctx).Create(user).Error
}

func (r *repo) UpdateUser(ctx context.Context, user *model.User) error {
	return r.provider.Db(ctx).
		Select("*").
		Where("id = ?", user.ID).
		Updates(user).
		Error
}

func (r *repo) DeleteUser(ctx context.Context, userUUID string) error {
	result := r.provider.Db(ctx).
		Where("uuid = ? AND deleted_at IS NULL", userUUID).
		Delete(&model.User{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
