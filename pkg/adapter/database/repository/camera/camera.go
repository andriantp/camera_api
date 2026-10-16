package camera

import (
	"context"

	"camera_api/pkg/adapter/database/model"

	"github.com/andriantp/camera"
	"gorm.io/gorm"
)

type Camera interface {
	FindCamerasByProtocol(ctx context.Context, protocolUUID string) ([]model.CameraSummary, error)
	FindCamerasByProtocolByUser(ctx context.Context, protocolUUID string, userUUID string) ([]model.CameraSummary, error)

	FindCameras(ctx context.Context) ([]model.Camera, error)
	FindCamerasByUser(ctx context.Context, userUUID string) ([]model.Camera, error)
	FindCameraAccess(ctx context.Context, cameraUUID string, userUUID string) (*model.CameraAccess, error)
	CreateCameraAccess(ctx context.Context, access *model.CameraAccess) error
	FindCameraByUUID(ctx context.Context, cameraUUID string) (*model.Camera, error)
	FindCameraAccessUser(ctx context.Context, cameraUUID string, userUUID string) (*model.Camera, error)

	CreateCamera(ctx context.Context, camera *model.Camera) error
}

func (r *repo) FindCameras(ctx context.Context) ([]model.Camera, error) {
	var cameras []model.Camera

	err := r.provider.Db(ctx).
		Where("deleted_at IS NULL").
		Order("id ASC").
		Preload("Protocol").
		Find(&cameras).
		Error

	if err != nil {
		return nil, err
	}

	return cameras, nil
}

func (r *repo) FindCamerasByUser(ctx context.Context, userUUID string) ([]model.Camera, error) {
	var cameras []model.Camera

	err := r.provider.Db(ctx).
		Joins("JOIN camera_access ca ON ca.camera_id = cameras.uuid").
		Where("ca.user_id = ?", userUUID).
		Preload("Protocol").
		Find(&cameras).Error

	if err != nil {
		return nil, err
	}

	return cameras, nil
}

func (r *repo) FindCameraAccess(ctx context.Context, cameraUUID string, userUUID string) (*model.CameraAccess, error) {
	var access model.CameraAccess

	err := r.provider.Db(ctx).
		Where("camera_id = ? AND user_id = ?", cameraUUID, userUUID).
		First(&access).Error

	if err != nil {
		return nil, err
	}

	return &access, nil
}

func (r *repo) CreateCameraAccess(ctx context.Context, access *model.CameraAccess) error {
	return r.provider.Db(ctx).Create(access).Error
}

func (r *repo) FindCameraByUUID(ctx context.Context, cameraUUID string) (*model.Camera, error) {
	var camera model.Camera

	err := r.provider.Db(ctx).
		Where("uuid = ? AND deleted_at IS NULL", cameraUUID).
		Preload("Protocol").
		Preload("Direction").
		Preload("Direction.Left").
		Preload("Direction.Right").
		Preload("Direction.Up").
		Preload("Direction.Down").
		First(&camera).
		Error

	if err != nil {
		return nil, err
	}

	return &camera, nil
}

func (r *repo) FindCameraAccessUser(ctx context.Context, cameraUUID string, userUUID string) (*model.Camera, error) {
	var camera model.Camera

	err := r.provider.Db(ctx).
		Joins("JOIN camera_access ca ON ca.camera_id = cameras.uuid").
		Where("cameras.uuid = ?", cameraUUID).
		Where("ca.user_id = ?", userUUID).
		Where("ca.active = ?", true).
		Where("cameras.deleted_at IS NULL").
		Preload("Protocol").
		Preload("Direction").
		Preload("Direction.Left").
		Preload("Direction.Right").
		Preload("Direction.Up").
		Preload("Direction.Down").
		First(&camera).
		Error

	if err != nil {
		return nil, err
	}

	return &camera, nil
}

func (r *repo) FindCamerasByProtocol(ctx context.Context, protocolUUID string) ([]model.CameraSummary, error) {
	var cameras []model.CameraSummary

	err := r.provider.Db(ctx).
		Table("cameras").
		Select(`
			cameras.uuid,
			cameras.sn,
			protocols.name AS protocol,
			cameras.active
		`).
		Joins("JOIN protocols ON protocols.uuid = cameras.protocol_id").
		Where("cameras.deleted_at IS NULL").
		Where("protocols.deleted_at IS NULL").
		Where("protocols.uuid = ?", protocolUUID).
		Order("cameras.id ASC").
		Scan(&cameras).
		Error

	if err != nil {
		return nil, err
	}

	return cameras, nil
}

func (r *repo) FindCamerasByProtocolByUser(ctx context.Context, protocolUUID string, userUUID string) ([]model.CameraSummary, error) {
	var cameras []model.CameraSummary

	err := r.provider.Db(ctx).
		Table("cameras").
		Select(`
			cameras.uuid,
			cameras.sn,
			protocols.name AS protocol,
			cameras.active
		`).
		Joins("JOIN protocols ON protocols.uuid = cameras.protocol_id").
		Joins("JOIN camera_access ca ON ca.camera_id = cameras.uuid").
		Where("cameras.deleted_at IS NULL").
		Where("protocols.deleted_at IS NULL").
		Where("protocols.uuid = ?", protocolUUID).
		Where("ca.user_id = ?", userUUID).
		Order("cameras.id ASC").
		Scan(&cameras).
		Error

	if err != nil {
		return nil, err
	}

	return cameras, nil
}

func (r *repo) CreateCamera(ctx context.Context, camera *model.Camera) error {
	return r.provider.Db(ctx).
		Create(camera).
		Error
}

func (r *repo) UpdateCameraSN(ctx context.Context, cameraUUID string, sn string) error {
	result := r.provider.Db(ctx).
		Model(&model.Camera{}).
		Where("uuid = ? AND deleted_at IS NULL", cameraUUID).
		Update("sn", sn)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *repo) UpdateCameraStatus(ctx context.Context, cameraUUID string, active bool) error {
	result := r.provider.Db(ctx).
		Model(&model.Camera{}).
		Where("uuid = ? AND deleted_at IS NULL", cameraUUID).
		Update("active", active)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *repo) UpdateCameraCredential(ctx context.Context, cameraUUID string, credential camera.Credential) error {
	result := r.provider.Db(ctx).
		Model(&model.Camera{}).
		Where("uuid = ? AND deleted_at IS NULL", cameraUUID).
		Updates(map[string]interface{}{
			"host":     credential.Host,
			"username": credential.Username,
			"password": credential.Password,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *repo) UpdateCameraProfile(ctx context.Context, cameraUUID string, profile string) error {
	result := r.provider.Db(ctx).
		Model(&model.Camera{}).
		Where("uuid = ? AND deleted_at IS NULL", cameraUUID).
		Update("profile", profile)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *repo) UpdateCameraChannel(ctx context.Context, cameraUUID string, channel int) error {
	result := r.provider.Db(ctx).
		Model(&model.Camera{}).
		Where("uuid = ? AND deleted_at IS NULL", cameraUUID).
		Update("channel", channel)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *repo) UpdateCameraDirection(ctx context.Context, cameraUUID, directionUUID string) error {
	result := r.provider.Db(ctx).
		Model(&model.Camera{}).
		Where("uuid = ? AND deleted_at IS NULL", cameraUUID).
		Update("direction_id", directionUUID)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
