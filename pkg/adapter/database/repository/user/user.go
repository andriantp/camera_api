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
	IncrementLoginFailed(ctx context.Context, userUUID string) (int, error)
	LockUser(ctx context.Context, userUUID string, until, now time.Time) (bool, error)
	ResetLoginFailures(ctx context.Context, userUUID string) error
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
			u.login_failed AS login_failed,
			u.locked_until AS locked_until
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
	var data model.UserSummary

	result := userSummaryQuery(r.provider.Db(ctx), userUUID).Scan(&data)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return &data, nil
}

// userSummaryQuery builds the single-user profile query without executing it.
func userSummaryQuery(db *gorm.DB, userUUID string) *gorm.DB {
	return db.
		Table("users u").
		Select(`
			u.uuid,
			u.username,
			r.id AS role_id,
			r.name AS role,
			u.last_login_at AS last_login,
			last_session.last_seen_at AS last_seen,
			u.active
		`).
		Joins("JOIN roles r ON r.uuid = u.role_id").
		Joins(latestSessionJoin("u.uuid")).
		Where("u.uuid = ? AND u.deleted_at IS NULL", userUUID)
}

// latestSessionJoin joins the user's most recent session as last_session.
// last_seen means the last time the user was active, whether or not they
// logged out since, and must mean the same in the list and the profile.
// NULLs are skipped (like MAX) and the plain DESC order lets
// idx_sessions_user_last_seen be read backwards.
func latestSessionJoin(userUUIDColumn string) string {
	return `
		LEFT JOIN LATERAL (
			SELECT s.last_seen_at
			FROM sessions s
			WHERE s.user_id = ` + userUUIDColumn + ` AND s.last_seen_at IS NOT NULL
			ORDER BY s.last_seen_at DESC
			LIMIT 1
		) last_session ON true`
}

func (r *repo) FindUsers(ctx context.Context, actorRoleID, limit, offset int) ([]model.UserSummary, error) {
	var users []model.UserSummary

	err := usersQuery(r.provider.Db(ctx), actorRoleID, limit, offset).
		Scan(&users).
		Error

	if err != nil {
		return nil, err
	}

	return users, nil
}

// usersQuery builds the user listing query without executing it, so the
// generated SQL can be tested without a database.
func usersQuery(db *gorm.DB, actorRoleID, limit, offset int) *gorm.DB {
	return db.
		Table("users").
		Select(`
			users.uuid,
			users.username,
			roles.id AS role_id,
			roles.name AS role,
			users.last_login_at AS last_login,
			last_session.last_seen_at AS last_seen,
			users.active,
			latest_activity.created_at AS last_activity,
			latest_activity.activity AS last_activity_type,
			latest_activity.message AS last_activity_message,
			COUNT(*) OVER() AS total
		`).
		Joins("JOIN roles ON roles.uuid = users.role_id").
		// Latest session per user via LATERAL instead of joining every session
		// and grouping.
		Joins(latestSessionJoin("users.uuid")).
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
		Order("roles.id ASC").
		Order("users.id ASC").
		Limit(limit).
		Offset(offset)
}

func (r *repo) UpdateUsername(ctx context.Context, userUUID, username string) error {
	return r.provider.Db(ctx).Model(&model.User{}).Where("uuid = ?", userUUID).Update("username", username).Error
}

func (r *repo) UpdatePassword(ctx context.Context, userUUID, password string) error {
	return r.provider.Db(ctx).Model(&model.User{}).Where("uuid = ?", userUUID).Update("password", password).Error
}

// IncrementLoginFailed adds one failed attempt in a single statement and
// returns the new count, so concurrent failed logins can't overwrite each
// other's increments.
func (r *repo) IncrementLoginFailed(ctx context.Context, userUUID string) (int, error) {
	var count int

	result := r.provider.Db(ctx).
		Raw(
			"UPDATE users SET login_failed = login_failed + 1 WHERE uuid = ? AND deleted_at IS NULL RETURNING login_failed",
			userUUID,
		).
		Scan(&count)

	if result.Error != nil {
		return 0, result.Error
	}

	if result.RowsAffected == 0 {
		return 0, gorm.ErrRecordNotFound
	}

	return count, nil
}

// LockUser locks a user until the given time and clears the failed-attempt
// counter. It reports whether this call set the lock: if the user is already
// locked (e.g. by a concurrent failed login), it returns false.
func (r *repo) LockUser(ctx context.Context, userUUID string, until, now time.Time) (bool, error) {
	result := r.provider.Db(ctx).
		Model(&model.User{}).
		Where("uuid = ? AND (locked_until IS NULL OR locked_until <= ?)", userUUID, now).
		Updates(map[string]interface{}{
			"locked_until": until,
			"login_failed": 0,
		})

	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected == 1, nil
}

// ResetLoginFailures clears the failed-attempt counter and any lock.
func (r *repo) ResetLoginFailures(ctx context.Context, userUUID string) error {
	return r.provider.Db(ctx).
		Model(&model.User{}).
		Where("uuid = ?", userUUID).
		Updates(map[string]interface{}{
			"login_failed": 0,
			"locked_until": nil,
		}).
		Error
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
