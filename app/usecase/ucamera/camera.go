package ucamera

import (
	"camera_api/pkg/adapter/database/model"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"
	"camera_api/pkg/utils/logger"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/andriantp/camera"
	"gorm.io/gorm"
)

type Camera interface {
	GetCamerasByProtocol(ctx context.Context, actor request.UserActor, protocolUUID string) (*response.CameraList, error)
	GetCameras(ctx context.Context, actor request.UserActor) (*response.CameraList, error)
	AssignCamera(ctx context.Context, actor request.UserActor, req request.AssignCameraRequest) error
	GetCamera(ctx context.Context, actor request.UserActor, cameraUUID string) (*response.CameraDetail, error)
	GetCameraWithCredential(ctx context.Context, cameraUUID string) (*response.CameraDetail, error)
	CreateCamera(ctx context.Context, actor request.UserActor, req request.CreateCamera) (*string, error)
}

func (u *usecase) GetCamerasByProtocol(ctx context.Context, actor request.UserActor, protocolUUID string) (*response.CameraList, error) {
	var cameras []model.CameraSummary
	var err error

	switch actor.Role {
	case "super_admin":
		cameras, err = u.repoCamera.FindCamerasByProtocol(ctx, protocolUUID)

	default:
		cameras, err = u.repoCamera.FindCamerasByProtocolByUser(ctx, protocolUUID, actor.UserID)
	}

	if err != nil {
		logger.Error("GetCamerasByProtocol-Repository", err.Error())

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	result := make([]response.CameraSummary, 0, len(cameras))

	for _, camera := range cameras {
		result = append(result, response.CameraSummary{
			UUID:     camera.UUID,
			SN:       camera.SN,
			Protocol: camera.Protocol,
			Active:   camera.Active,
		})
	}

	return &response.CameraList{
		Total:   int64(len(result)),
		Cameras: result,
	}, nil
}

func (u *usecase) GetCameras(ctx context.Context, actor request.UserActor) (*response.CameraList, error) {
	var cameras []model.Camera
	var err error

	switch actor.Role {
	case "super_admin":
		cameras, err = u.repoCamera.FindCameras(ctx)

	default:
		cameras, err = u.repoCamera.FindCamerasByUser(ctx, actor.UserID)
	}

	if err != nil {
		logger.Error("GetCameras-Repository", err.Error())

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	result := make([]response.CameraSummary, 0, len(cameras))

	for _, camera := range cameras {
		result = append(result, response.CameraSummary{
			UUID:     camera.UUID.String(),
			SN:       camera.SN,
			Protocol: camera.Protocol.Name,
			Active:   camera.Active,
		})
	}

	return &response.CameraList{
		Total:   int64(len(result)),
		Cameras: result,
	}, nil
}

func (u *usecase) AssignCamera(ctx context.Context, actor request.UserActor, req request.AssignCameraRequest) error {
	return u.repoCamera.DoTransaction(ctx, func(ctxWithTx context.Context, _ *gorm.DB) error {
		now := time.Now().UTC()

		access, err := u.repoCamera.FindCameraAccess(
			ctxWithTx,
			req.CameraUUID,
			req.UserUUID,
		)

		if err == nil && access != nil {
			return response.APIError{
				StatusCode: http.StatusConflict,
				Message:    "Camera already assigned to user",
			}
		}

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error("AssignCamera-FindCameraAccess", err.Error())

			return response.APIError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Internal server error",
			}
		}

		if err := u.repoCamera.CreateCameraAccess(
			ctxWithTx,
			&model.CameraAccess{
				CameraID:  req.CameraUUID,
				UserID:    req.UserUUID,
				CreatedBy: actor.UserID,
				Active:    true,
			},
		); err != nil {
			logger.Error("AssignCamera-CreateCameraAccess", err.Error())

			return response.APIError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Internal server error",
			}
		}

		if err := u.repoUser.UpdateSessionLastSeen(
			ctxWithTx,
			actor.SessionID,
			now,
			actor.IPAddress,
			actor.UserAgent,
		); err != nil {
			return err
		}

		msg := fmt.Sprintf("Assigned camera %s", req.CameraUUID)

		activity := model.UserActivity{
			ActorUserID: actor.UserID,
			SessionID:   &actor.SessionID,
			CameraID:    &req.CameraUUID,
			IPAddress:   actor.IPAddress,
			UserAgent:   actor.UserAgent,
			Activity:    "ASSIGNED_CAMERA",
			Message:     &msg,
			CreatedAt:   now,
		}

		return u.repoUser.CreateActivity(ctxWithTx, &activity)
	})
}

func (u *usecase) GetCamera(ctx context.Context, actor request.UserActor, cameraUUID string) (*response.CameraDetail, error) {
	camera, err := u.findCamera(ctx, actor, "GetCamera", cameraUUID)
	if err != nil {
		return nil, err
	}

	result := &response.CameraDetail{
		UUID: camera.UUID.String(),
		SN:   camera.SN,
		Protocol: response.UUIDName{
			UUID: camera.Protocol.UUID.String(),
			Name: camera.Protocol.Name,
		},
		Profile: camera.Profile,
		Channel: camera.Channel,
		Active:  camera.Active,
	}

	if camera.SearchID != nil || camera.TrackID != nil {
		result.Content = &response.Content{
			SearchID: camera.SearchID,
			TrackID:  camera.TrackID,
		}
	}

	if camera.Direction != nil {
		result.Direction = &response.Direction{
			UUID: camera.Direction.UUID.String(),
			Name: camera.Direction.Name,
		}

		result.Direction.Left = mapPanTilt(camera.Direction.Left)
		result.Direction.Right = mapPanTilt(camera.Direction.Right)
		result.Direction.Up = mapPanTilt(camera.Direction.Up)
		result.Direction.Down = mapPanTilt(camera.Direction.Down)
	}

	return result, nil
}

func (u *usecase) GetCameraWithCredential(ctx context.Context, cameraUUID string) (*response.CameraDetail, error) {
	row, err := u.repoCamera.FindCameraByUUID(ctx, cameraUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.APIError{
				StatusCode: http.StatusNotFound,
				Message:    "Camera not found",
			}
		}

		logger.Error("GetCamera-FindCameraByUUID", err.Error())

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	credential, err := u.camera.DecryptCredential(camera.Credential{
		Host:     row.Host,
		Username: row.Username,
		Password: row.Password,
	})
	if err != nil {
		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Decrypt Credential error",
		}
	}

	result := &response.CameraDetail{
		UUID: row.UUID.String(),
		SN:   row.SN,
		Protocol: response.UUIDName{
			UUID: row.Protocol.UUID.String(),
			Name: row.Protocol.Name,
		},
		Profile: row.Profile,
		Channel: row.Channel,
		Active:  row.Active,
		Credential: &response.Credential{
			Host:     credential.Host,
			Username: credential.Username,
			Password: credential.Password,
		},
	}

	if row.SearchID != nil || row.TrackID != nil {
		result.Content = &response.Content{
			SearchID: row.SearchID,
			TrackID:  row.TrackID,
		}
	}

	if row.Direction != nil {
		result.Direction = &response.Direction{
			UUID: row.Direction.UUID.String(),
			Name: row.Direction.Name,
		}

		result.Direction.Left = mapPanTilt(row.Direction.Left)
		result.Direction.Right = mapPanTilt(row.Direction.Right)
		result.Direction.Up = mapPanTilt(row.Direction.Up)
		result.Direction.Down = mapPanTilt(row.Direction.Down)
	}

	return result, nil
}

func (u *usecase) CreateCamera(ctx context.Context, actor request.UserActor, req request.CreateCamera) (*string, error) {
	var cameraUUID string

	credential, err := u.camera.EncryptCredential(camera.Credential{
		Host:     req.Host,
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		logger.Error("CreateCamera-EncryptCredential", err.Error())

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Encrypt Credential error",
		}
	}

	_, err = u.repoCamera.FindProtocolByUUID(ctx, req.Protocol)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.APIError{
				StatusCode: http.StatusBadRequest,
				Message:    "Protocol not found",
			}
		}

		logger.Error("CreateCamera-FindProtocol", err.Error())

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	camera := model.Camera{
		CreatedBy: actor.UserID,

		SN:         strings.ToUpper(req.SN),
		ProtocolID: req.Protocol,

		Host:     credential.Host,
		Username: credential.Username,
		Password: credential.Password,
		Active:   true,
	}

	err = u.repoUser.DoTransaction(ctx, func(ctxWithTx context.Context, _ *gorm.DB) error {
		now := time.Now().UTC()

		if err := u.repoCamera.CreateCamera(ctxWithTx, &camera); err != nil {
			return err
		}

		if actor.Role != "super_admin" {
			if err := u.repoCamera.CreateCameraAccess(
				ctxWithTx,
				&model.CameraAccess{
					CameraID:  camera.UUID.String(),
					UserID:    actor.UserID,
					CreatedBy: actor.UserID,
					Active:    true,
				},
			); err != nil {
				return err
			}
		}

		if err := u.repoUser.UpdateSessionLastSeen(ctxWithTx, actor.SessionID, now, actor.IPAddress, actor.UserAgent); err != nil {
			return err
		}

		cameraUUID = camera.UUID.String()

		msg := fmt.Sprintf("Created camera %s", camera.SN)
		activity := model.UserActivity{
			ActorUserID: actor.UserID,
			SessionID:   &actor.SessionID,
			CameraID:    &cameraUUID,
			IPAddress:   actor.IPAddress,
			UserAgent:   actor.UserAgent,
			Activity:    "CREATE_CAMERA",
			Message:     &msg,
			CreatedAt:   now,
		}

		return u.repoUser.CreateActivity(ctxWithTx, &activity)
	})

	if err != nil {
		logger.Error("CreateCamera-Transaction", err.Error())

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	return &cameraUUID, nil
}
