package camera

import (
	"camera_api/pkg/adapter/database/model"
	"context"

	"github.com/andriantp/camera"
)

type Configuration interface {
	UpdateCameraSN(ctx context.Context, cameraUUID string, sn string) error
	UpdateCameraStatus(ctx context.Context, cameraUUID string, active bool) error
	UpdateCameraCredential(ctx context.Context, cameraUUID string, credential camera.Credential) error
	UpdateCameraProfile(ctx context.Context, cameraUUID string, profile string) error
	UpdateCameraChannel(ctx context.Context, cameraUUID string, channel int) error
	UpdateCameraDirection(ctx context.Context, cameraUUID, directionUUID string) error
	UpdateCameraContent(ctx context.Context, cameraUUID string, searchID int, trackID int) error
}

func (r *repo) UpdateCameraContent(ctx context.Context, cameraUUID string, searchID int, trackID int) error {
	return r.provider.Db(ctx).
		Model(&model.Camera{}).
		Where("uuid = ? AND deleted_at IS NULL", cameraUUID).
		Updates(map[string]interface{}{
			"search_id": searchID,
			"track_id":  trackID,
		}).
		Error
}
