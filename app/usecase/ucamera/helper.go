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
	"time"

	"gorm.io/gorm"
)

func (u *usecase) findCamera(ctx context.Context, actor request.UserActor, operation string, cameraUUID string) (*model.Camera, error) {
	var camera *model.Camera
	var err error

	switch actor.Role {
	case "super_admin":
		camera, err = u.repoCamera.FindCameraByUUID(ctx, cameraUUID)

	default:
		camera, err = u.repoCamera.FindCameraAccessUser(
			ctx,
			cameraUUID,
			actor.UserID,
		)
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.APIError{
				StatusCode: http.StatusNotFound,
				Message:    "Camera not found",
			}
		}

		logger.Error(
			fmt.Sprintf("findCamera-usecase %s", operation),
			err.Error(),
		)

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	return camera, nil
}

func (u *usecase) createActivity(ctx context.Context, now time.Time, actor request.UserActor, cameraUUID string, activity string, message string) error {
	err := u.repoUser.DoTransaction(
		ctx,
		func(ctxWithTx context.Context, _ *gorm.DB) error {
			if err := u.repoUser.UpdateSessionLastSeen(
				ctxWithTx,
				actor.SessionID,
				now,
				actor.IPAddress,
				actor.UserAgent,
			); err != nil {
				return err
			}

			return u.repoUser.CreateActivity(
				ctxWithTx,
				&model.UserActivity{
					ActorUserID: actor.UserID,
					SessionID:   &actor.SessionID,
					CameraID:    &cameraUUID,
					IPAddress:   actor.IPAddress,
					UserAgent:   actor.UserAgent,
					Activity:    activity,
					Message:     &message,
					CreatedAt:   now,
				},
			)
		},
	)

	if err != nil {
		logger.Error("CreateActivity", err.Error())
		var apiErr response.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	return nil
}
