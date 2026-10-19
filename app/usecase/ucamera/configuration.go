package ucamera

import (
	adaptercamera "camera_api/pkg/adapter/camera"
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

type Configuration interface {
	UpdateCameraSN(ctx context.Context, actor request.UserActor, cameraUUID string, req request.UpdateCameraSN) error
	UpdateCameraStatus(ctx context.Context, actor request.UserActor, cameraUUID string, req request.UpdateCameraStatus) error
	UpdateCameraCredential(ctx context.Context, actor request.UserActor, cameraUUID string, req request.UpdateCameraCredential) error
	UpdateCameraDirection(ctx context.Context, actor request.UserActor, cameraUUID string, req request.UpdateCameraDirection) error

	GetConfiguration(ctx context.Context, actor request.UserActor, cameraUUID string) (any, error)
	UpdateConfiguration(ctx context.Context, actor request.UserActor, cameraUUID string, req request.UpdateCameraConfiguration) error
}

func (u *usecase) UpdateCameraSN(ctx context.Context, actor request.UserActor, cameraUUID string, req request.UpdateCameraSN) error {
	err := u.repoUser.DoTransaction(ctx, func(ctxWithTx context.Context, _ *gorm.DB) error {
		now := time.Now().UTC()

		_, err := u.findCamera(ctxWithTx, actor, "UpdateCameraSN", cameraUUID)
		if err != nil {
			return err
		}

		if err := u.repoCamera.UpdateCameraSN(ctxWithTx, cameraUUID, strings.ToUpper(req.SN)); err != nil {
			return err
		}

		if err := u.repoUser.UpdateSessionLastSeen(ctxWithTx, actor.SessionID, now, actor.IPAddress, actor.UserAgent); err != nil {
			return err
		}

		msg := fmt.Sprintf("Updated camera %s serial number to %s", cameraUUID, req.SN)

		activity := model.UserActivity{
			ActorUserID: actor.UserID,
			CameraID:    &cameraUUID,
			SessionID:   &actor.SessionID,
			IPAddress:   actor.IPAddress,
			UserAgent:   actor.UserAgent,
			Activity:    "UPDATE_CAMERA_SN",
			Message:     &msg,
			CreatedAt:   now,
		}

		return u.repoUser.CreateActivity(ctxWithTx, &activity)
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.APIError{
				StatusCode: http.StatusNotFound,
				Message:    "Camera not found",
			}
		}

		logger.Error("UpdateCameraSN-Transaction", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	return nil
}

func (u *usecase) UpdateCameraStatus(ctx context.Context, actor request.UserActor, cameraUUID string, req request.UpdateCameraStatus) error {
	err := u.repoUser.DoTransaction(ctx, func(ctxWithTx context.Context, _ *gorm.DB) error {
		now := time.Now().UTC()

		_, err := u.findCamera(ctxWithTx, actor, "UpdateCameraStatus", cameraUUID)
		if err != nil {
			return err
		}

		if err := u.repoCamera.UpdateCameraStatus(ctxWithTx, cameraUUID, *req.Active); err != nil {
			return err
		}

		if err := u.repoUser.UpdateSessionLastSeen(ctxWithTx, actor.SessionID, now, actor.IPAddress, actor.UserAgent); err != nil {
			return err
		}

		action := "Deactivated"
		activity := "DEACTIVATE_CAMERA"

		if *req.Active {
			action = "Activated"
			activity = "ACTIVATE_CAMERA"
		}

		msg := fmt.Sprintf("%s camera %s", action, cameraUUID)
		activityLog := model.UserActivity{
			ActorUserID: actor.UserID,
			CameraID:    &cameraUUID,
			SessionID:   &actor.SessionID,
			IPAddress:   actor.IPAddress,
			UserAgent:   actor.UserAgent,
			Activity:    activity,
			Message:     &msg,
			CreatedAt:   now,
		}

		return u.repoUser.CreateActivity(
			ctxWithTx,
			&activityLog,
		)
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.APIError{
				StatusCode: http.StatusNotFound,
				Message:    "Camera not found",
			}
		}

		logger.Error("UpdateCameraStatus-Transaction", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	return nil
}

func (u *usecase) UpdateCameraCredential(ctx context.Context, actor request.UserActor, cameraUUID string, req request.UpdateCameraCredential) error {
	err := u.repoUser.DoTransaction(ctx, func(ctxWithTx context.Context, _ *gorm.DB) error {
		now := time.Now().UTC()

		_, err := u.findCamera(ctxWithTx, actor, "UpdateCameraCredential", cameraUUID)
		if err != nil {
			return err
		}

		row, err := u.repoCamera.FindCameraByUUID(ctxWithTx, cameraUUID)
		if err != nil {
			return err
		}
		credential := camera.Credential{
			Host:     row.Host,
			Username: row.Username,
			Password: row.Password,
		}

		if req.Host != nil {
			secret, err := u.camera.Encrypt(*req.Host)
			if err != nil {
				logger.Error("UpdateCameraCredential-Encrypt-Host", err.Error())
				return err
			}
			credential.Host = secret
		}

		if req.Username != nil {
			secret, err := u.camera.Encrypt(*req.Username)
			if err != nil {
				logger.Error("UpdateCameraCredential-Encrypt-Username", err.Error())
				return err
			}
			credential.Username = secret
		}

		if req.Password != nil {
			secret, err := u.camera.Encrypt(*req.Password)
			if err != nil {
				logger.Error("UpdateCameraCredential-Encrypt-Password", err.Error())
				return err
			}
			credential.Password = secret
		}

		if err := u.repoCamera.UpdateCameraCredential(ctxWithTx, cameraUUID, credential); err != nil {
			return err
		}

		if err := u.repoUser.UpdateSessionLastSeen(ctxWithTx, actor.SessionID, now, actor.IPAddress, actor.UserAgent); err != nil {
			return err
		}

		msg := fmt.Sprintf("Updated credential for camera %s", cameraUUID)

		activity := model.UserActivity{
			ActorUserID: actor.UserID,
			CameraID:    &cameraUUID,
			SessionID:   &actor.SessionID,
			IPAddress:   actor.IPAddress,
			UserAgent:   actor.UserAgent,
			Activity:    "UPDATE_CAMERA_CREDENTIAL",
			Message:     &msg,
			CreatedAt:   now,
		}

		return u.repoUser.CreateActivity(
			ctxWithTx,
			&activity,
		)
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.APIError{
				StatusCode: http.StatusNotFound,
				Message:    "Camera not found",
			}
		}

		logger.Error("UpdateCameraCredential-Transaction", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	return nil
}

func (u *usecase) UpdateCameraDirection(ctx context.Context, actor request.UserActor, cameraUUID string, req request.UpdateCameraDirection) error {
	err := u.repoUser.DoTransaction(ctx, func(ctxWithTx context.Context, _ *gorm.DB) error {
		now := time.Now().UTC()

		_, err := u.findCamera(ctxWithTx, actor, "UpdateChannel", cameraUUID)
		if err != nil {
			return err
		}

		direction, err := u.repoCamera.GetDirectionByUUID(ctx, req.Direction)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return response.APIError{
					StatusCode: http.StatusNotFound,
					Message:    "Direction not found",
				}
			}

			logger.Error("UpdateCameraDirection-FindDirection", err.Error())

			return response.APIError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Internal server error",
			}
		}

		if err := u.repoCamera.UpdateCameraDirection(ctxWithTx, cameraUUID, direction.UUID.String()); err != nil {
			return err
		}

		if err := u.repoUser.UpdateSessionLastSeen(ctxWithTx, actor.SessionID, now, actor.IPAddress, actor.UserAgent); err != nil {
			return err
		}

		msg := fmt.Sprintf("Assigned direction %s to camera %s", direction.Name, cameraUUID)

		activity := model.UserActivity{
			ActorUserID: actor.UserID,
			CameraID:    &cameraUUID,
			SessionID:   &actor.SessionID,
			IPAddress:   actor.IPAddress,
			UserAgent:   actor.UserAgent,
			Activity:    "UPDATE_CAMERA_DIRECTION",
			Message:     &msg,
			CreatedAt:   now,
		}

		return u.repoUser.CreateActivity(ctxWithTx, &activity)
	})

	if err != nil {
		logger.Error("UpdateCameraDirection-Transaction", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	return nil
}

func (u *usecase) GetConfiguration(ctx context.Context, actor request.UserActor, cameraUUID string) (any, error) {
	row, err := u.findCamera(ctx, actor, "UpdateChannel", cameraUUID)
	if err != nil {
		return nil, err
	}

	config, err := u.camera.GetConfiguration(ctx,
		adaptercamera.CameraTarget{
			Credential: camera.Credential{
				Host:     row.Host,
				Username: row.Username,
				Password: row.Password,
			},
			Protocol: row.Protocol.Name,
		},
	)
	if err != nil {
		logger.Error("GetConfiguration-GetConfiguration", err.Error())
		return nil, err
	}

	return config, nil
}

func (u *usecase) UpdateConfiguration(ctx context.Context, actor request.UserActor, cameraUUID string, req request.UpdateCameraConfiguration) error {
	err := u.repoUser.DoTransaction(ctx, func(ctxWithTx context.Context, _ *gorm.DB) error {
		now := time.Now().UTC()

		camera, err := u.findCamera(ctxWithTx, actor, "UpdateConfiguration", cameraUUID)
		if err != nil {
			return err
		}

		var msg string
		switch req.Target {
		case "profile":
			if camera.Protocol.Name != model.ProtocolONVIF {
				return response.APIError{
					StatusCode: http.StatusConflict,
					Message:    "Unsupported camera protocol",
				}
			}
			if req.Profile == nil {
				return response.APIError{
					StatusCode: http.StatusConflict,
					Message:    "Profile is required",
				}
			}

			if err := u.repoCamera.UpdateCameraProfile(ctxWithTx, cameraUUID, *req.Profile); err != nil {
				return err
			}
			msg = fmt.Sprintf(
				"Updated profile for camera %s to %s",
				cameraUUID,
				req.Profile,
			)

		case "channel":
			if camera.Protocol.Name != model.ProtocolISAPI {
				return response.APIError{
					StatusCode: http.StatusConflict,
					Message:    "Unsupported camera protocol",
				}
			}
			if req.Channel == nil {
				return response.APIError{
					StatusCode: http.StatusConflict,
					Message:    "Channel is required",
				}
			}
			if *req.Channel < 1 {
				return response.APIError{
					StatusCode: http.StatusConflict,
					Message:    "Channel must be greater than or equal to 1",
				}
			}

			if err := u.repoCamera.UpdateCameraChannel(ctxWithTx, cameraUUID, *req.Channel); err != nil {
				return err
			}
			msg = fmt.Sprintf(
				"Updated channel for camera %s to %d",
				cameraUUID,
				*req.Channel,
			)

		case "content":
			if camera.Protocol.Name != model.ProtocolISAPI {
				return response.APIError{
					StatusCode: http.StatusConflict,
					Message:    "Unsupported camera protocol",
				}
			}

			if req.Content == nil {
				return response.APIError{
					StatusCode: http.StatusConflict,
					Message:    "Content is required",
				}
			}

			if req.Content.SearchID == nil {
				return response.APIError{
					StatusCode: http.StatusConflict,
					Message:    "Search ID is required",
				}
			}

			if req.Content.TrackID == nil {
				return response.APIError{
					StatusCode: http.StatusConflict,
					Message:    "Track ID is required",
				}
			}

			if err := u.repoCamera.UpdateCameraContent(
				ctxWithTx,
				cameraUUID,
				*req.Content.SearchID,
				*req.Content.TrackID,
			); err != nil {
				return err
			}

			msg = fmt.Sprintf(
				"Updated content for camera %s to searchID=%d, trackID=%d",
				cameraUUID,
				*req.Content.SearchID,
				*req.Content.TrackID,
			)

		default:
			return response.APIError{
				StatusCode: http.StatusConflict,
				Message:    "Unsupported Target",
			}
		}

		if err := u.repoUser.UpdateSessionLastSeen(ctxWithTx, actor.SessionID, now, actor.IPAddress, actor.UserAgent); err != nil {
			return err
		}

		activity := model.UserActivity{
			ActorUserID: actor.UserID,
			CameraID:    &cameraUUID,
			SessionID:   &actor.SessionID,
			IPAddress:   actor.IPAddress,
			UserAgent:   actor.UserAgent,
			Activity:    "UPDATE_CAMERA_CONFIGURATION",
			Message:     &msg,
			CreatedAt:   now,
		}

		return u.repoUser.CreateActivity(ctxWithTx, &activity)
	})

	if err != nil {
		var apiErr response.APIError

		if errors.As(err, &apiErr) {
			return apiErr
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.APIError{
				StatusCode: http.StatusNotFound,
				Message:    "Camera not found",
			}
		}

		logger.Error("UpdateConfiguration-Transaction", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	return nil
}
