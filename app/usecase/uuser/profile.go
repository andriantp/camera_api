package uuser

import (
	"camera_api/pkg/adapter/database/model"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"
	"camera_api/pkg/utils/logger"
	"camera_api/pkg/utils/password"
	"context"
	"errors"
	"net/http"
	"time"

	"gorm.io/gorm"
)

type Profile interface {
	GetMyProfile(ctx context.Context, userUUID string) (*response.Profile, error)
	UpdateMyProfile(ctx context.Context, actor request.UserActor, body request.UpdateProfile) error
	ChangeMyPassword(ctx context.Context, actor request.UserActor, body request.ChangePassword) error
}

func (u *usecase) GetMyProfile(ctx context.Context, userUUID string) (*response.Profile, error) {
	user, err := u.repoUser.FindUserByUserUUID(ctx, userUUID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, response.APIError{
			StatusCode: http.StatusNotFound,
			Message:    "User Not Found",
		}
	}
	if err != nil {
		logger.Error("GetProfile-FindUserByUserUUID", err.Error())

		return nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	return &response.Profile{
		UUID:     user.UserUUID,
		Username: user.UserName,
		Role:     user.RoleName,
	}, nil
}

func (u *usecase) UpdateMyProfile(ctx context.Context, actor request.UserActor, body request.UpdateProfile) error {
	now := time.Now().UTC()

	err := u.repoUser.DoTransaction(ctx, func(ctxWithTx context.Context, _ *gorm.DB) error {
		if err := u.repoUser.UpdateUsername(ctxWithTx, actor.UserID, *body.Username); err != nil {
			return err
		}

		if err := u.repoUser.UpdateSessionLastSeen(ctxWithTx, actor.SessionID, now, actor.IPAddress, actor.UserAgent); err != nil {
			return err
		}

		msg := "User updated profile"
		activity := model.UserActivity{
			ActorUserID: actor.UserID,
			SessionID:   &actor.SessionID,
			IPAddress:   actor.IPAddress,
			UserAgent:   actor.UserAgent,
			Activity:    "UPDATE_PROFILE",
			Message:     &msg,
			CreatedAt:   now,
		}

		return u.repoUser.CreateActivity(ctxWithTx, &activity)
	})

	if err != nil {
		logger.Error("UpdateMyProfile-Transaction", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	return nil
}

func (u *usecase) ChangeMyPassword(ctx context.Context, actor request.UserActor, body request.ChangePassword) error {
	myPassword, err := u.repoUser.FindPassword(ctx, actor.UserID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return response.APIError{
			StatusCode: http.StatusNotFound,
			Message:    "User Not Found",
		}
	}
	if err != nil {
		logger.Error("ChangeMyPassword-FindPassword", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	_, err = password.CheckPassword(body.CurrentPassword, myPassword)
	if err != nil {
		return response.APIError{
			StatusCode: http.StatusUnauthorized,
			Message:    "Invalid current password",
		}
	}

	if body.CurrentPassword == body.NewPassword {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "New password must be different from current password",
		}
	}

	hashedPassword, err := password.CreatePassword(body.NewPassword)
	if err != nil {
		logger.Error("ChangeMyPassword-CreatePassword", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to create password",
		}
	}

	now := time.Now().UTC()
	err = u.repoUser.DoTransaction(ctx, func(ctxWithTx context.Context, _ *gorm.DB) error {
		if err := u.repoUser.UpdatePassword(ctxWithTx, actor.UserID, hashedPassword); err != nil {
			return err
		}

		if err := u.repoUser.UpdateLoginFailed(ctxWithTx, actor.UserID, 0); err != nil {
			return err
		}

		if err := u.repoUser.UpdateSessionLastSeen(ctxWithTx, actor.SessionID, now, actor.IPAddress, actor.UserAgent); err != nil {
			return err
		}

		msg := "User changed password"
		activity := model.UserActivity{
			ActorUserID: actor.UserID,
			SessionID:   &actor.SessionID,
			IPAddress:   actor.IPAddress,
			UserAgent:   actor.UserAgent,
			Activity:    "CHANGE_PASSWORD",
			Message:     &msg,
			CreatedAt:   now,
		}

		return u.repoUser.CreateActivity(ctxWithTx, &activity)
	})

	if err != nil {
		logger.Error("ChangeMyPassword-Transaction", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	return nil
}
