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
	RevokeCameraAccess(ctx context.Context, actor request.UserActor, cameraUUID, userUUID string) error
	GetCamera(ctx context.Context, actor request.UserActor, cameraUUID string) (*response.CameraDetail, error)
	GetCameraWithCredential(ctx context.Context, cameraUUID string) (*response.CameraDetail, error)
	CreateCamera(ctx context.Context, actor request.UserActor, req request.CreateCamera) (*string, error)
}

func (u *usecase) GetCamerasByProtocol(ctx context.Context, actor request.UserActor, protocolUUID string) (*response.CameraList, error) {
	var cameras []model.CameraSummary
	var err error

	if bypassesCameraAccess(actor.Role) {
		cameras, err = u.repoCamera.FindCamerasByProtocol(ctx, protocolUUID)
	} else {
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

	if bypassesCameraAccess(actor.Role) {
		cameras, err = u.repoCamera.FindCameras(ctx)
	} else {
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
	err := u.repoCamera.DoTransaction(ctx, func(ctxWithTx context.Context, _ *gorm.DB) error {
		now := time.Now().UTC()

		cam, err := u.repoCamera.FindCameraByUUID(ctxWithTx, req.CameraUUID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.APIError{
				StatusCode: http.StatusNotFound,
				Message:    "Camera not found",
			}
		}
		if err != nil {
			return err
		}

		target, err := u.repoUser.FindUserByUserUUID(ctxWithTx, req.UserUUID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.APIError{
				StatusCode: http.StatusNotFound,
				Message:    "User not found",
			}
		}
		if err != nil {
			return err
		}

		// super_admin and admin already see every camera; a grant for them
		// would have no effect.
		if bypassesCameraAccess(target.RoleName) {
			return response.APIError{
				StatusCode: http.StatusBadRequest,
				Message:    "Camera access can only be assigned to operator or user",
			}
		}

		action := "Assigned"

		access, err := u.repoCamera.FindCameraAccess(ctxWithTx, req.CameraUUID, req.UserUUID)
		switch {
		case err == nil && access.Active:
			return response.APIError{
				StatusCode: http.StatusConflict,
				Message:    "Camera already assigned to user",
			}

		case err == nil:
			// A revoked grant exists: re-enable it instead of adding a row.
			if err := u.repoCamera.ActivateCameraAccess(ctxWithTx, access.ID); err != nil {
				return err
			}

			action = "Reactivated"

		case errors.Is(err, gorm.ErrRecordNotFound):
			err := u.repoCamera.CreateCameraAccess(
				ctxWithTx,
				&model.CameraAccess{
					CameraID:  req.CameraUUID,
					UserID:    req.UserUUID,
					CreatedBy: actor.UserID,
					Active:    true,
				},
			)
			// A concurrent request created the same grant first.
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return response.APIError{
					StatusCode: http.StatusConflict,
					Message:    "Camera already assigned to user",
				}
			}
			if err != nil {
				return err
			}

		default:
			return err
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

		msg := fmt.Sprintf(
			"%s access to camera %s (%s) for user %s",
			action,
			cam.SN,
			req.CameraUUID,
			target.UserName,
		)

		activity := model.UserActivity{
			ActorUserID:  actor.UserID,
			TargetUserID: &req.UserUUID,
			SessionID:    &actor.SessionID,
			CameraID:     &req.CameraUUID,
			IPAddress:    actor.IPAddress,
			UserAgent:    actor.UserAgent,
			Activity:     "ASSIGNED_CAMERA",
			Message:      &msg,
			CreatedAt:    now,
		}

		return u.repoUser.CreateActivity(ctxWithTx, &activity)
	})
	if err != nil {
		return transactionError("AssignCamera", err, "Camera not found")
	}

	return nil
}

// RevokeCameraAccess deactivates a user's grant to a camera. The camera
// disappears from the user's lists and can no longer be opened; assigning it
// again re-enables the same grant.
func (u *usecase) RevokeCameraAccess(ctx context.Context, actor request.UserActor, cameraUUID, userUUID string) error {
	err := u.repoCamera.DoTransaction(ctx, func(ctxWithTx context.Context, _ *gorm.DB) error {
		now := time.Now().UTC()

		cam, err := u.repoCamera.FindCameraByUUID(ctxWithTx, cameraUUID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.APIError{
				StatusCode: http.StatusNotFound,
				Message:    "Camera not found",
			}
		}
		if err != nil {
			return err
		}

		access, err := u.repoCamera.FindCameraAccess(ctxWithTx, cameraUUID, userUUID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.APIError{
				StatusCode: http.StatusNotFound,
				Message:    "Camera access not found",
			}
		}
		if err != nil {
			return err
		}

		if !access.Active {
			return response.APIError{
				StatusCode: http.StatusConflict,
				Message:    "Camera access already revoked",
			}
		}

		if err := u.repoCamera.DeactivateCameraAccess(ctxWithTx, access.ID); err != nil {
			return err
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

		// The grant may outlive its user (soft-deleted); fall back to the UUID.
		username := userUUID
		if target, err := u.repoUser.FindUserByUserUUID(ctxWithTx, userUUID); err == nil {
			username = target.UserName
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		msg := fmt.Sprintf(
			"Revoked access to camera %s (%s) for user %s",
			cam.SN,
			cameraUUID,
			username,
		)

		activity := model.UserActivity{
			ActorUserID:  actor.UserID,
			TargetUserID: &userUUID,
			SessionID:    &actor.SessionID,
			CameraID:     &cameraUUID,
			IPAddress:    actor.IPAddress,
			UserAgent:    actor.UserAgent,
			Activity:     "REVOKED_CAMERA",
			Message:      &msg,
			CreatedAt:    now,
		}

		return u.repoUser.CreateActivity(ctxWithTx, &activity)
	})
	if err != nil {
		return transactionError("RevokeCameraAccess", err, "Camera not found")
	}

	return nil
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

		// Roles restricted by camera_access need a grant to see the camera
		// they just created.
		if !bypassesCameraAccess(actor.Role) {
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
