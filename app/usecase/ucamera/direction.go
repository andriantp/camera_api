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

	"gorm.io/gorm"
)

type Direction interface {
	GetPanTilts(ctx context.Context) (*response.PanTiltList, error)
	CreatePanTilt(ctx context.Context, actor request.UserActor, req request.CreatePanTilt) (*response.PanTiltSummary, error)
	GetPanTiltByUUID(ctx context.Context, panTiltUUID string) (*response.PanTiltDetail, error)
	UpdatePanTilt(ctx context.Context, actor request.UserActor, panTiltUUID string, req request.UpdatePanTilt) error

	GetDirections(ctx context.Context) (*response.DirectionList, error)
	CreateDirection(ctx context.Context, actor request.UserActor, req request.CreateDirection) (*response.Direction, error)
	UpdateDirection(ctx context.Context, actor request.UserActor, directionUUID string, req request.UpdateDirection) error
	GetDirectionByUUID(ctx context.Context, directionUUID string) (*response.DirectionDetail, error)
}

func (u *usecase) GetPanTilts(ctx context.Context) (*response.PanTiltList, error) {
	panTilts, err := u.repoCamera.GetPanTilts(ctx)
	if err != nil {
		logger.Error("GetPanTilts-GetPanTilts", err.Error())

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	result := make([]response.PanTiltSummary, 0, len(panTilts))

	for _, panTilt := range panTilts {
		result = append(result, response.PanTiltSummary{
			UUID:     panTilt.UUID.String(),
			Name:     panTilt.Name,
			Pan:      panTilt.Pan,
			Tilt:     panTilt.Tilt,
			Duration: panTilt.Duration.String(),
		})
	}

	return &response.PanTiltList{
		Total:    int64(len(result)),
		PanTilts: result,
	}, nil
}

func (u *usecase) CreatePanTilt(ctx context.Context, actor request.UserActor, req request.CreatePanTilt) (*response.PanTiltSummary, error) {
	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		return nil, response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid duration",
		}
	}

	panTilt := &model.PanTilt{
		CreatedBy: actor.UserID,
		Name:      strings.ToUpper(req.Name),
		Pan:       *req.Pan,
		Tilt:      *req.Tilt,
		Duration:  duration,
	}

	err = u.repoUser.DoTransaction(ctx, func(ctxWithTx context.Context, _ *gorm.DB) error {
		now := time.Now().UTC()

		if err := u.repoCamera.CreatePanTilt(ctxWithTx, panTilt); err != nil {
			return err
		}

		if err := u.repoUser.UpdateSessionLastSeen(ctxWithTx, actor.SessionID, now, actor.IPAddress, actor.UserAgent); err != nil {
			return err
		}

		msg := fmt.Sprintf("Created pan tilt %s", panTilt.Name)
		activity := model.UserActivity{
			ActorUserID: actor.UserID,
			SessionID:   &actor.SessionID,
			IPAddress:   actor.IPAddress,
			UserAgent:   actor.UserAgent,
			Activity:    "CREATE_PAN_TILT",
			Message:     &msg,
			CreatedAt:   now,
		}

		return u.repoUser.CreateActivity(ctxWithTx, &activity)
	})

	if err != nil {
		logger.Error("CreatePanTilt-Transaction", err.Error())

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	return &response.PanTiltSummary{
		UUID:     panTilt.UUID.String(),
		Name:     panTilt.Name,
		Pan:      panTilt.Pan,
		Tilt:     panTilt.Tilt,
		Duration: panTilt.Duration.String(),
	}, nil
}

func (u *usecase) GetPanTiltByUUID(ctx context.Context, panTiltUUID string) (*response.PanTiltDetail, error) {
	panTilt, err := u.repoCamera.GetPanTiltByUUID(ctx, panTiltUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.APIError{
				StatusCode: http.StatusNotFound,
				Message:    "Pan tilt not found",
			}
		}

		logger.Error("GetPanTiltByUUID-GetPanTiltByUUID", err.Error())

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	return &response.PanTiltDetail{
		UUID:      panTilt.UUID.String(),
		Name:      panTilt.Name,
		Pan:       panTilt.Pan,
		Tilt:      panTilt.Tilt,
		Duration:  panTilt.Duration.String(),
		CreatedBy: panTilt.CreatedUser.Username,
		CreatedAt: panTilt.CreatedAt,
		UpdatedAt: panTilt.UpdatedAt,
	}, nil
}

func (u *usecase) UpdatePanTilt(ctx context.Context, actor request.UserActor, panTiltUUID string, req request.UpdatePanTilt) error {
	updates := make(map[string]interface{})

	if req.Name != nil {
		name := *req.Name
		updates["name"] = strings.ToUpper(name)
	}

	if req.Pan != nil {
		updates["pan"] = *req.Pan
	}

	if req.Tilt != nil {
		updates["tilt"] = *req.Tilt
	}

	if req.Duration != nil {
		duration, err := time.ParseDuration(*req.Duration)
		if err != nil {
			return response.APIError{
				StatusCode: http.StatusBadRequest,
				Message:    "Invalid duration",
			}
		}

		updates["duration"] = duration
	}

	err := u.repoUser.DoTransaction(ctx, func(ctxWithTx context.Context, _ *gorm.DB) error {
		now := time.Now().UTC()

		if err := u.repoCamera.UpdatePanTilt(ctxWithTx, panTiltUUID, updates); err != nil {
			return err
		}

		if err := u.repoUser.UpdateSessionLastSeen(ctxWithTx, actor.SessionID, now, actor.IPAddress, actor.UserAgent); err != nil {
			return err
		}

		msg := fmt.Sprintf("Updated pan tilt %s", panTiltUUID)

		activity := model.UserActivity{
			ActorUserID: actor.UserID,
			SessionID:   &actor.SessionID,
			IPAddress:   actor.IPAddress,
			UserAgent:   actor.UserAgent,
			Activity:    "UPDATE_PAN_TILT",
			Message:     &msg,
			CreatedAt:   now,
		}

		return u.repoUser.CreateActivity(ctxWithTx, &activity)
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.APIError{
				StatusCode: http.StatusNotFound,
				Message:    "Pan tilt not found",
			}
		}

		logger.Error("UpdatePanTilt-Transaction", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	return nil
}

func (u *usecase) GetDirections(ctx context.Context) (*response.DirectionList, error) {
	directions, err := u.repoCamera.GetDirections(ctx)
	if err != nil {
		logger.Error("GetDirections-GetDirections", err.Error())

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	result := make([]response.DirectionSummary, 0, len(directions))

	for _, direction := range directions {
		result = append(result, response.DirectionSummary{
			UUID: direction.UUID.String(),
			Name: direction.Name,
		})
	}

	return &response.DirectionList{
		Total:      int64(len(result)),
		Directions: result,
	}, nil
}

func (u *usecase) CreateDirection(ctx context.Context, actor request.UserActor, req request.CreateDirection) (*response.Direction, error) {
	direction := &model.Direction{
		CreatedBy: actor.UserID,
		Name:      strings.ToUpper(req.Name),
		LeftID:    req.Left,
		RightID:   req.Right,
		UpID:      req.Up,
		DownID:    req.Down,
	}

	var result *response.Direction
	err := u.repoUser.DoTransaction(ctx, func(ctxWithTx context.Context, _ *gorm.DB) error {
		now := time.Now().UTC()

		if err := u.repoCamera.CreateDirection(ctxWithTx, direction); err != nil {
			return err
		}

		if err := u.repoUser.UpdateSessionLastSeen(ctxWithTx, actor.SessionID, now, actor.IPAddress, actor.UserAgent); err != nil {
			return err
		}

		msg := fmt.Sprintf("Created direction %s", direction.Name)

		activity := model.UserActivity{
			ActorUserID: actor.UserID,
			SessionID:   &actor.SessionID,
			IPAddress:   actor.IPAddress,
			UserAgent:   actor.UserAgent,
			Activity:    "CREATE_DIRECTION",
			Message:     &msg,
			CreatedAt:   now,
		}

		return u.repoUser.CreateActivity(ctxWithTx, &activity)
	})

	if err != nil {
		logger.Error("CreateDirection-Transaction", err.Error())

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	result = &response.Direction{
		UUID: direction.UUID.String(),
		Name: direction.Name,
	}
	result = mapDirection(direction)

	return result, nil
}

func (u *usecase) UpdateDirection(ctx context.Context, actor request.UserActor, directionUUID string, req request.UpdateDirection) error {
	updates := make(map[string]interface{})

	if req.Name != nil {
		name := *req.Name
		updates["name"] = strings.ToUpper(name)
	}

	if req.Left != nil {
		updates["left_id"] = *req.Left
	}

	if req.Right != nil {
		updates["right_id"] = *req.Right
	}

	if req.Up != nil {
		updates["up_id"] = *req.Up
	}

	if req.Down != nil {
		updates["down_id"] = *req.Down
	}

	if len(updates) == 0 {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "At least one field is required",
		}
	}

	err := u.repoUser.DoTransaction(ctx, func(ctxWithTx context.Context, _ *gorm.DB) error {
		now := time.Now().UTC()

		if err := u.repoCamera.UpdateDirection(ctxWithTx, directionUUID, updates); err != nil {
			return err
		}

		if err := u.repoUser.UpdateSessionLastSeen(ctxWithTx, actor.SessionID, now, actor.IPAddress, actor.UserAgent); err != nil {
			return err
		}

		msg := fmt.Sprintf("Updated direction %s", directionUUID)

		activity := model.UserActivity{
			ActorUserID: actor.UserID,
			SessionID:   &actor.SessionID,
			IPAddress:   actor.IPAddress,
			UserAgent:   actor.UserAgent,
			Activity:    "UPDATE_DIRECTION",
			Message:     &msg,
			CreatedAt:   now,
		}

		return u.repoUser.CreateActivity(ctxWithTx, &activity)
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.APIError{
				StatusCode: http.StatusNotFound,
				Message:    "Direction not found",
			}
		}

		logger.Error("UpdateDirection-Transaction", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	return nil
}

func (u *usecase) GetDirectionByUUID(ctx context.Context, directionUUID string) (*response.DirectionDetail, error) {
	direction, err := u.repoCamera.GetDirectionByUUID(ctx, directionUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.APIError{
				StatusCode: http.StatusNotFound,
				Message:    "Direction not found",
			}
		}

		logger.Error("GetDirectionByUUID-GetDirectionByUUID", err.Error())

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	return mapDirectionDetail(direction), nil
}

func mapDirection(d *model.Direction) *response.Direction {
	if d == nil {
		return nil
	}

	return &response.Direction{
		UUID:  d.UUID.String(),
		Name:  d.Name,
		Left:  mapPanTilt(d.Left),
		Right: mapPanTilt(d.Right),
		Up:    mapPanTilt(d.Up),
		Down:  mapPanTilt(d.Down),
	}
}

func mapDirectionDetail(direction *model.Direction) *response.DirectionDetail {
	if direction == nil {
		return nil
	}

	return &response.DirectionDetail{
		UUID:      direction.UUID.String(),
		Name:      direction.Name,
		Left:      mapPanTiltDetail(direction.Left),
		Right:     mapPanTiltDetail(direction.Right),
		Up:        mapPanTiltDetail(direction.Up),
		Down:      mapPanTiltDetail(direction.Down),
		CreatedBy: direction.CreatedUser.Username,
		CreatedAt: direction.CreatedAt,
		UpdatedAt: direction.UpdatedAt,
	}
}

func mapPanTilt(p *model.PanTilt) *response.PanTilt {
	if p == nil {
		return nil
	}

	return &response.PanTilt{
		UUID:     p.UUID.String(),
		Name:     p.Name,
		Pan:      p.Pan,
		Tilt:     p.Tilt,
		Duration: p.Duration.String(),
	}
}

func mapPanTiltDetail(p *model.PanTilt) *response.PanTiltDetail {
	if p == nil {
		return nil
	}

	return &response.PanTiltDetail{
		UUID:      p.UUID.String(),
		Name:      p.Name,
		Pan:       p.Pan,
		Tilt:      p.Tilt,
		Duration:  p.Duration.String(),
		CreatedBy: p.CreatedUser.Username,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
