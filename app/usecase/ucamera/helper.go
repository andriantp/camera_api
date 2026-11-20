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
	"github.com/andriantp/camera"
	"net/http"
	"time"

	"gorm.io/gorm"
)

// bypassesCameraAccess reports whether a role sees every camera. super_admin
// and admin are not restricted by camera_access; operator and user only see
// cameras granted to them.
func bypassesCameraAccess(role string) bool {
	switch role {
	case "super_admin", "admin":
		return true
	default:
		return false
	}
}

// findCamera loads a camera the actor may access, with its protocol and
// credential: enough for every camera action. Use findCameraDetail when the
// direction and pan-tilts are needed.
func (u *usecase) findCamera(ctx context.Context, actor request.UserActor, operation string, cameraUUID string) (*model.Camera, error) {
	return u.lookupCamera(ctx, actor, operation, cameraUUID, false)
}

// findCameraDetail is findCamera plus the direction and its pan-tilts.
func (u *usecase) findCameraDetail(ctx context.Context, actor request.UserActor, operation string, cameraUUID string) (*model.Camera, error) {
	return u.lookupCamera(ctx, actor, operation, cameraUUID, true)
}

func (u *usecase) lookupCamera(ctx context.Context, actor request.UserActor, operation string, cameraUUID string, detail bool) (*model.Camera, error) {
	var camera *model.Camera
	var err error

	switch {
	case bypassesCameraAccess(actor.Role) && detail:
		camera, err = u.repoCamera.FindCameraDetailByUUID(ctx, cameraUUID)
	case bypassesCameraAccess(actor.Role):
		camera, err = u.repoCamera.FindCameraByUUID(ctx, cameraUUID)
	case detail:
		camera, err = u.repoCamera.FindCameraAccessUserDetail(ctx, cameraUUID, actor.UserID)
	default:
		camera, err = u.repoCamera.FindCameraAccessUser(ctx, cameraUUID, actor.UserID)
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

// maxMoveDuration caps how long a relative PTZ move may run. Since
// andriantp/camera v1.0.7, RelativeMove blocks until the duration elapses,
// so an unbounded value would hold the HTTP request (and the camera) for
// that long.
const maxMoveDuration = 10 * time.Second

// parseMoveDuration parses a move duration such as "500ms" or "3s" and
// enforces 0 < duration <= maxMoveDuration.
func parseMoveDuration(value string) (time.Duration, error) {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid duration",
		}
	}

	if err := checkMoveDuration(duration); err != nil {
		return 0, err
	}

	return duration, nil
}

// checkMoveDuration validates a duration that is already parsed, e.g. one
// loaded from a stored pan-tilt.
func checkMoveDuration(duration time.Duration) error {
	if duration <= 0 || duration > maxMoveDuration {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    fmt.Sprintf("Duration must be greater than 0 and at most %s", maxMoveDuration),
		}
	}

	return nil
}

// errNoChange aborts a transaction when the requested state already holds
// (e.g. access already granted). Callers report it as a successful no-op.
var errNoChange = errors.New("no change")

// transactionError maps an error returned from a transaction to an API error:
// APIErrors raised inside the transaction (e.g. 404 from findCamera) pass
// through unchanged, gorm.ErrRecordNotFound becomes 404 with notFoundMessage,
// and anything else is logged and becomes 500.
func transactionError(operation string, err error, notFoundMessage string) error {
	var apiErr response.APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return response.APIError{
			StatusCode: http.StatusNotFound,
			Message:    notFoundMessage,
		}
	}

	logger.Error(operation+"-Transaction", err.Error())

	return response.APIError{
		StatusCode: http.StatusInternalServerError,
		Message:    "Internal server error",
	}
}

// cameraTarget builds the adapter target for a camera row. Every field is
// passed; each adapter call reads only what it needs, and unset optional
// values (profile, channel, search/track ID) fall back to adapter defaults.
func cameraTarget(row *model.Camera) adaptercamera.CameraTarget {
	return adaptercamera.CameraTarget{
		Credential: camera.Credential{
			Host:     row.Host,
			Username: row.Username,
			Password: row.Password,
		},
		Protocol: row.Protocol.Name,
		Profile:  row.Profile,
		Channel:  row.Channel,
		SearchID: row.SearchID,
		TrackID:  row.TrackID,
	}
}
