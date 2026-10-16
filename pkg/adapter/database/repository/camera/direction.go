package camera

import (
	"camera_api/pkg/adapter/database/model"
	"context"

	"gorm.io/gorm"
)

type Direction interface {
	GetPanTilts(ctx context.Context) ([]model.PanTilt, error)
	CreatePanTilt(ctx context.Context, panTilt *model.PanTilt) error
	GetPanTiltByUUID(ctx context.Context, panTiltUUID string) (*model.PanTilt, error)
	UpdatePanTilt(ctx context.Context, panTiltUUID string, updates map[string]interface{}) error

	GetDirections(ctx context.Context) ([]model.Direction, error)
	CreateDirection(ctx context.Context, direction *model.Direction) error
	UpdateDirection(ctx context.Context, directionUUID string, updates map[string]interface{}) error
	GetDirectionByUUID(ctx context.Context, directionUUID string) (*model.Direction, error)
}

func (r *repo) GetPanTilts(ctx context.Context) ([]model.PanTilt, error) {
	var panTilts []model.PanTilt

	err := r.provider.Db(ctx).
		Where("deleted_at IS NULL").
		Order("id ASC").
		Find(&panTilts).
		Error

	if err != nil {
		return nil, err
	}

	return panTilts, nil
}

func (r *repo) CreatePanTilt(ctx context.Context, panTilt *model.PanTilt) error {
	return r.provider.Db(ctx).
		Create(panTilt).
		Error
}

func (r *repo) GetPanTiltByUUID(ctx context.Context, panTiltUUID string) (*model.PanTilt, error) {
	var panTilt model.PanTilt

	err := r.provider.Db(ctx).
		Where("uuid = ? AND deleted_at IS NULL", panTiltUUID).
		Preload("CreatedUser").
		Take(&panTilt).
		Error

	if err != nil {
		return nil, err
	}

	return &panTilt, nil
}

func (r *repo) UpdatePanTilt(ctx context.Context, panTiltUUID string, updates map[string]interface{}) error {
	result := r.provider.Db(ctx).
		Model(&model.PanTilt{}).
		Where("uuid = ? AND deleted_at IS NULL", panTiltUUID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *repo) GetDirections(ctx context.Context) ([]model.Direction, error) {
	var directions []model.Direction

	err := r.provider.Db(ctx).
		Where("deleted_at IS NULL").
		Order("id ASC").
		Find(&directions).
		Error

	if err != nil {
		return nil, err
	}

	return directions, nil
}

func (r *repo) CreateDirection(ctx context.Context, direction *model.Direction) error {
	return r.provider.Db(ctx).Create(direction).Error
}

func (r *repo) UpdateDirection(ctx context.Context, directionUUID string, updates map[string]interface{}) error {
	result := r.provider.Db(ctx).
		Model(&model.Direction{}).
		Where("uuid = ? AND deleted_at IS NULL", directionUUID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *repo) GetDirectionByUUID(ctx context.Context, directionUUID string) (*model.Direction, error) {
	var direction model.Direction

	err := r.provider.Db(ctx).
		Where("uuid = ? AND deleted_at IS NULL", directionUUID).
		Preload("Left.CreatedUser").
		Preload("Right.CreatedUser").
		Preload("Up.CreatedUser").
		Preload("Down.CreatedUser").
		Preload("CreatedUser").
		Take(&direction).
		Error

	if err != nil {
		return nil, err
	}

	return &direction, nil
}