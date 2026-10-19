package ucamera

import (
	adaptercamera "camera_api/pkg/adapter/camera"
	"camera_api/pkg/api/request"
	"camera_api/pkg/utils/logger"
	"context"

	cameraModel "github.com/andriantp/camera/model"

	"github.com/andriantp/camera"
)

type Device interface {
	GetCapabilities(ctx context.Context, actor request.UserActor, cameraUUID string) (*cameraModel.Capabilities, error)
	GetPTZCapability(ctx context.Context, actor request.UserActor, cameraUUID string) (*adaptercamera.PTZCapability, error)
	GetStorage(ctx context.Context, actor request.UserActor, cameraUUID string) (*adaptercamera.Storage, error)
}

func (u *usecase) GetCapabilities(ctx context.Context, actor request.UserActor, cameraUUID string) (*cameraModel.Capabilities, error) {
	row, err := u.findCamera(ctx, actor, "GetCapability", cameraUUID)
	if err != nil {
		return nil, err
	}

	capabilities, err := u.camera.GetCapability(
		ctx,
		adaptercamera.CameraTarget{
			Credential: camera.Credential{
				Host:     row.Host,
				Username: row.Username,
				Password: row.Password,
			},
			Protocol: row.Protocol.Name,
			Profile:  row.Profile,
			Channel:  row.Channel,
		},
	)
	if err != nil {
		logger.Error("GetCapability", err.Error())
		return nil, err
	}

	return capabilities, nil
}

func (u *usecase) GetPTZCapability(ctx context.Context, actor request.UserActor, cameraUUID string) (*adaptercamera.PTZCapability, error) {
	row, err := u.findCamera(ctx, actor, "GetPTZCapability", cameraUUID)
	if err != nil {
		return nil, err
	}

	capabilities, err := u.camera.GetPTZCapability(
		ctx,
		adaptercamera.CameraTarget{
			Credential: camera.Credential{
				Host:     row.Host,
				Username: row.Username,
				Password: row.Password,
			},
			Protocol: row.Protocol.Name,
			Profile:  row.Profile,
			Channel:  row.Channel,
		},
	)
	if err != nil {
		logger.Error("GetPTZCapability", err.Error())
		return nil, err
	}

	return capabilities, nil
}

func (u *usecase) GetStorage(ctx context.Context, actor request.UserActor, cameraUUID string) (*adaptercamera.Storage, error) {
	row, err := u.findCamera(ctx, actor, "GetStorage", cameraUUID)
	if err != nil {
		return nil, err
	}

	storage, err := u.camera.GetStorage(ctx,
		adaptercamera.CameraTarget{
			Credential: camera.Credential{
				Host:     row.Host,
				Username: row.Username,
				Password: row.Password,
			},
			Protocol: row.Protocol.Name,
			Profile:  row.Profile,
			Channel:  row.Channel,
		},
	)
	if err != nil {
		logger.Error("GetPosition-Camera", err.Error())
		return nil, err
	}

	return storage, nil
}
