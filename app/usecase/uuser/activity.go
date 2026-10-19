package uuser

import (
	"camera_api/pkg/adapter/database/repository/user"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"
	"camera_api/pkg/utils/helper"
	"camera_api/pkg/utils/logger"
	"context"
	"net/http"
	"time"

	"gorm.io/gorm"
)

type Activity interface {
	GetMyActivities(ctx context.Context, actor request.UserActor, filter request.Pagination) (*response.ActivityUser, error)
	GetUserActivities(ctx context.Context, actor request.UserActor, userUUID string, filter request.Pagination) (*response.ActivityUser, error)
}

func (u *usecase) GetMyActivities(ctx context.Context, actor request.UserActor, filter request.Pagination) (*response.ActivityUser, error) {
	start, end, loc, err := helper.ConvertTime(filter.From, filter.To)
	if err != nil {
		logger.Error("GetMyActivities-ConvertTime", err.Error())
		return nil, err
	}

	activities, err := u.repoUser.FindUserActivities(ctx, user.ActivityFilter{
		ActorUserID: actor.UserID,
		Target:      user.ActivityTargetAll,
		From:        *start,
		To:          *end,
		Limit:       filter.Limit,
		Offset:      filter.Offset,
		Order:       filter.Order,
	})
	if err != nil {
		logger.Error("GetMyActivities-FindUserActivities", err.Error())

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	total := 0
	if len(activities) > 0 {
		total = activities[0].Total
	}

	result := make([]response.Activity, 0, len(activities))

	for _, activity := range activities {
		result = append(result, response.Activity{
			UUID: activity.UUID,
			At:   activity.At.In(loc),
			// Actor:    activity.Actor,
			Activity: activity.Activity,
			Target:   activity.Target,
			Camera:   activity.Camera,
			Message:  activity.Message,
		})
	}

	return &response.ActivityUser{
		Total:    total,
		Returned: len(result),
		Activities: response.ActivityUserRange{
			Range: response.Range{
				From: filter.From,
				To:   filter.To,
			},
			Items: result,
		},
	}, nil
}

func (u *usecase) GetUserActivities(ctx context.Context, actor request.UserActor, userUUID string, filter request.Pagination) (*response.ActivityUser, error) {
	// Make sure actor can access target user.
	targetUser, err := u.repoUser.FindUserByUserUUID(ctx, userUUID)
	if err != nil {
		logger.Error("FindUserByUserUUID-FindUser", err.Error())

		return nil, response.APIError{
			StatusCode: http.StatusNotFound,
			Message:    "User not found",
		}
	}

	targetRole, err := u.repoUser.FindRoleByUUID(ctx, targetUser.RoleUUID)
	if err != nil {
		logger.Error("GetUserActivities-FindRole", err.Error())

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	actorRole, err := u.repoUser.FindRoleByName(ctx, actor.Role)
	if err != nil {
		logger.Error("GetUserActivities-FindActorRole", err.Error())

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	if targetRole.ID < actorRole.ID {
		return nil, response.APIError{
			StatusCode: http.StatusForbidden,
			Message:    "Forbidden",
		}
	}

	start, end, loc, err := helper.ConvertTime(filter.From, filter.To)
	if err != nil {
		logger.Error("GetMyActivities-ConvertTime", err.Error())
		return nil, err
	}

	activities, err := u.repoUser.FindUserActivities(ctx, user.ActivityFilter{
		ActorUserID: userUUID,
		Target:      user.ActivityTarget(filter.Target),
		From:        *start,
		To:          *end,
		Limit:       filter.Limit,
		Offset:      filter.Offset,
		Order:       filter.Order,
	})
	if err != nil {
		logger.Error("GetUserActivities-FindUserActivities", err.Error())

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	total := 0
	if len(activities) > 0 {
		total = activities[0].Total
	}

	result := make([]response.Activity, 0, len(activities))

	for _, activity := range activities {
		result = append(result, response.Activity{
			UUID:      activity.UUID,
			At:        activity.At.In(loc),
			Actor:     activity.Actor,
			IPAddress: activity.IPAddress,
			UserAgent: activity.UserAgent,
			Activity:  activity.Activity,
			Target:    activity.Target,
			Camera:    activity.Camera,
			Message:   activity.Message,
		})
	}

	err = u.repoUser.DoTransaction(ctx, func(ctxWithTx context.Context, _ *gorm.DB) error {
		now := time.Now().UTC()
		if err := u.repoUser.UpdateSessionLastSeen(ctxWithTx, actor.SessionID, now, actor.IPAddress, actor.UserAgent); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		logger.Error("GetUserActivities-Transaction", err.Error())
	}

	return &response.ActivityUser{
		Total:    total,
		Returned: len(result),
		Activities: response.ActivityUserRange{
			Range: response.Range{
				From: filter.From,
				To:   filter.To,
			},
			Items: result,
		},
	}, nil
}
