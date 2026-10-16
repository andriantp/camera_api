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
