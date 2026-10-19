package ucamera

import (
	adaptercamera "camera_api/pkg/adapter/camera"
	"camera_api/pkg/adapter/database/model"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"
	"camera_api/pkg/utils/logger"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/andriantp/camera"
	cameraModel "github.com/andriantp/camera/model"
)

type Command interface {
	GetPosition(ctx context.Context, actor request.UserActor, cameraUUID string) (any, error)

	CalibrationMove(ctx context.Context, actor request.UserActor, cameraUUID string, req request.RelativeMove) (*cameraModel.Status, error)
	MoveTo(ctx context.Context, actor request.UserActor, cameraUUID string, direction string) (*cameraModel.Status, error)

	GetPresets(ctx context.Context, actor request.UserActor, cameraUUID string) ([]cameraModel.Preset, error)
	MoveToPreset(ctx context.Context, actor request.UserActor, cameraUUID string, presetID int) error

	MoveToHome(ctx context.Context, actor request.UserActor, cameraUUID string) error

	GetZoom(ctx context.Context, actor request.UserActor, cameraUUID string) (*adaptercamera.Zoom, error)
	SetZoom(ctx context.Context, actor request.UserActor, cameraUUID string, value float64) (*float64, error)
}

func (u *usecase) GetPosition(ctx context.Context, actor request.UserActor, cameraUUID string) (any, error) {
	row, err := u.findCamera(ctx, actor, "GetCapability", cameraUUID)
	if err != nil {
		return nil, err
	}

	position, err := u.camera.GetPosition(ctx,
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

	return position, nil
}

func (u *usecase) CalibrationMove(ctx context.Context, actor request.UserActor, cameraUUID string, req request.RelativeMove) (*cameraModel.Status, error) {
	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		return nil, response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid duration",
		}
	}

	row, err := u.findCamera(ctx, actor, "CalibrationMove", cameraUUID)
	if err != nil {
		return nil, err
	}

	status, err := u.camera.CalibrationMove(
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
		cameraModel.PanTilt{
			X: *req.Pan,
			Y: *req.Tilt,
		},
		duration,
	)
	if err != nil {
		logger.Error("RelativeMove-Camera", err.Error())
		return nil, err
	}

	now := time.Now().UTC()
	msg := fmt.Sprintf(
		"CalibrationMove move camera %s: pan=%v tilt=%v duration=%s",
		cameraUUID,
		*req.Pan,
		*req.Tilt,
		duration,
	)
	if err := u.createActivity(
		ctx,
		now,
		actor,
		cameraUUID,
		"CAMERA_CALIBRATION_MOVE",
		msg,
	); err != nil {
		return nil, err
	}

	return status, nil
}

func (u *usecase) MoveTo(ctx context.Context, actor request.UserActor, cameraUUID string, direction string) (*cameraModel.Status, error) {
	row, err := u.findCamera(ctx, actor, "MoveTo", cameraUUID)
	if err != nil {
		return nil, err
	}

	if row.Direction == nil {
		return nil, response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "Direction Not Set",
		}
	}

	var panTilt *model.PanTilt
	switch direction {
	case "LEFT":
		panTilt = row.Direction.Left
	case "RIGHT":
		panTilt = row.Direction.Right
	case "UP":
		panTilt = row.Direction.Up
	case "DOWN":
		panTilt = row.Direction.Down
	default:
		return nil, response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid direction",
		}
	}

	if panTilt == nil {
		return nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "PanTilt is not configured",
		}
	}

	status, err := u.camera.Move(ctx,
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
		*panTilt,
	)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	msg := fmt.Sprintf("Camera %s Move To %s", cameraUUID, direction)
	if err := u.createActivity(
		ctx,
		now,
		actor,
		cameraUUID,
		"CAMERA_MOVEMENT",
		msg,
	); err != nil {
		return nil, err
	}

	return status, nil
}

func (u *usecase) GetPresets(ctx context.Context, actor request.UserActor, cameraUUID string) ([]cameraModel.Preset, error) {
	row, err := u.findCamera(ctx, actor, "GetPresets", cameraUUID)
	if err != nil {
		return nil, err
	}

	if row.Protocol.Name != model.ProtocolISAPI {
		return nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unsupported camera protocol",
		}
	}

	if row.Channel == nil {
		return nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Channel is not configured",
		}
	}

	presets, err := u.camera.GetPresets(ctx,
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
		logger.Error("GetPresets-GetPresets", err.Error())
		return nil, err
	}

	return presets, nil
}

func (u *usecase) MoveToPreset(ctx context.Context, actor request.UserActor, cameraUUID string, presetID int) error {
	row, err := u.findCamera(ctx, actor, "MoveToPreset", cameraUUID)
	if err != nil {
		return err
	}

	if row.Protocol.Name != model.ProtocolISAPI {
		return response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unsupported camera protocol",
		}
	}

	if row.Channel == nil {
		return response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Channel is not configured",
		}
	}

	if err := u.camera.GotoPreset(ctx,
		adaptercamera.CameraTarget{
			Credential: camera.Credential{
				Host:     row.Host,
				Username: row.Username,
				Password: row.Password,
			},
			Protocol: row.Protocol.Name,
			Profile:  row.Profile,
			Channel:  row.Channel,
		}, presetID,
	); err != nil {
		logger.Error("MoveToPreset-GotoPreset", err.Error())
		return err
	}

	now := time.Now().UTC()
	msg := fmt.Sprintf("Camera %s moved to preset %d", cameraUUID, presetID)
	if err := u.createActivity(
		ctx,
		now,
		actor,
		cameraUUID,
		"CAMERA_GOTO_PRESET",
		msg,
	); err != nil {
		return err
	}

	return nil
}

func (u *usecase) MoveToHome(ctx context.Context, actor request.UserActor, cameraUUID string) error {
	row, err := u.findCamera(ctx, actor, "MoveToHome", cameraUUID)
	if err != nil {
		return err
	}

	if row.Protocol.Name != model.ProtocolISAPI {
		return response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Unsupported camera protocol",
		}
	}

	if row.Channel == nil {
		return response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Channel is not configured",
		}
	}

	if err := u.camera.GotoHome(ctx,
		adaptercamera.CameraTarget{
			Credential: camera.Credential{
				Host:     row.Host,
				Username: row.Username,
				Password: row.Password,
			},
			Protocol: row.Protocol.Name,
			Profile:  row.Profile,
			Channel:  row.Channel,
		}); err != nil {
		logger.Error("MoveToHome-GotoHome", err.Error())
		return err
	}

	now := time.Now().UTC()
	msg := fmt.Sprintf("Camera %s moved to home", cameraUUID)
	if err := u.createActivity(
		ctx,
		now,
		actor,
		cameraUUID,
		"CAMERA_MOVE_HOME",
		msg,
	); err != nil {
		return err
	}

	return nil
}

func (u *usecase) GetZoom(ctx context.Context, actor request.UserActor, cameraUUID string) (*adaptercamera.Zoom, error) {
	row, err := u.findCamera(ctx, actor, "GetZoom", cameraUUID)
	if err != nil {
		return nil, err
	}

	zoom, err := u.camera.GetZoom(ctx,
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
		logger.Error("GetZoom-Camera", err.Error())
		return nil, err
	}

	return zoom, nil
}

func (u *usecase) SetZoom(ctx context.Context, actor request.UserActor, cameraUUID string, value float64) (*float64, error) {
	row, err := u.findCamera(ctx, actor, "GetZoom", cameraUUID)
	if err != nil {
		return nil, err
	}

	zoom, err := u.camera.SetZoom(ctx,
		adaptercamera.CameraTarget{
			Credential: camera.Credential{
				Host:     row.Host,
				Username: row.Username,
				Password: row.Password,
			},
			Protocol: row.Protocol.Name,
			Profile:  row.Profile,
			Channel:  row.Channel,
		}, value,
	)
	if err != nil {
		logger.Error("GetZoom-Camera", err.Error())
		return nil, err
	}

	return zoom, nil
}
