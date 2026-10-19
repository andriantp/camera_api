package uuser

import (
	"camera_api/pkg/adapter/database/model"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"
	"camera_api/pkg/utils/logger"
	"camera_api/pkg/utils/password"
	"camera_api/pkg/utils/token"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"gorm.io/gorm"
)

type Auth interface {
	Login(ctx context.Context, body request.Login) (*string, *string, error)
	Logout(ctx context.Context, actor request.UserActor) error
	Refresh(ctx context.Context, refreshToken string, actor request.UserActor) (*string, *string, error)
}

func (u *usecase) GetExpire() Expire {
	return Expire{
		Token:   u.exp.Token,
		Refresh: u.exp.Refresh,
	}
}

func (u *usecase) Login(ctx context.Context, body request.Login) (*string, *string, error) {
	user, err := u.repoUser.FindUserByUserName(ctx, body.UserName)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, response.APIError{
			StatusCode: http.StatusNotFound,
			Message:    "User Not Found",
		}
	} else if err != nil {
		logger.Error("Login-FindUserByUserName", err.Error())
		return nil, nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	if !user.Active {
		return nil, nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "User Not Active",
		}
	}

	_, err = password.CheckPassword(body.Password, user.Password)
	if err != nil {
		user.LoginFailed++
		deactived := user.LoginFailed > 3

		err = u.repoUser.DoTransaction(ctx, func(ctxWithTx context.Context, _ *gorm.DB) error {
			if err := u.repoUser.UpdateLoginFailed(ctxWithTx, user.UserUUID, user.LoginFailed); err != nil {
				return err
			}

			if !deactived {
				return nil
			}

			if err := u.repoUser.UpdateStatus(ctxWithTx, user.UserUUID, false); err != nil {
				return err
			}

			msg := fmt.Sprintf(
				"Deactivated user after %d failed login attempts from %s",
				user.LoginFailed,
				body.UserAgent,
			)

			activity := model.UserActivity{
				ActorUserID: user.UserUUID,
				SessionID:   nil,
				IPAddress:   body.IPAddress,
				UserAgent:   body.UserAgent,
				Activity:    "LOGIN_FAILED",
				Message:     &msg,
			}

			return u.repoUser.CreateActivity(ctxWithTx, &activity)
		})

		if err != nil {
			logger.Error("Login-UpdateLoginFailed", err.Error())

			return nil, nil, response.APIError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Internal server error",
			}
		}

		if deactived {
			return nil, nil, response.APIError{
				StatusCode: http.StatusUnauthorized,
				Message:    "Please contact admin for activation",
			}
		}

		return nil, nil, response.APIError{
			StatusCode: http.StatusUnauthorized,
			Message:    "Invalid username or password",
		}
	}

	claims := token.ClaimsFromUserSession(user)

	tokenKey, err := u.token.GenerateToken(claims, u.exp.Token, body.SessionID)
	if err != nil {
		logger.Error("Login-GenerateToken", err.Error())
		return nil, nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed GenerateToken",
		}
	}

	tokenRefresh, err := u.token.GenerateRefresh(claims, u.exp.Refresh, body.SessionID)
	if err != nil {
		logger.Error("Login-GenerateRefresh", err.Error())
		return nil, nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed GenerateRefresh",
		}
	}

	now := time.Now().UTC()
	err = u.repoUser.DoTransaction(ctx, func(ctxWithTx context.Context, _ *gorm.DB) error {
		// reset login_failed
		if user.LoginFailed > 0 {
			if err := u.repoUser.UpdateLoginFailed(ctxWithTx, user.UserUUID, 0); err != nil {
				return err
			}
		}

		// update last_login_at
		if err := u.repoUser.UpdateLastLogin(ctxWithTx, user.UserUUID, now); err != nil {
			return err
		}

		// create session
		session := model.Session{
			UUID:       body.SessionID,
			UserID:     user.UserUUID,
			LoginAt:    now,
			LastSeenAt: now,
			IPAddress:  body.IPAddress,
			UserAgent:  body.UserAgent,
		}
		if err := u.repoUser.CreateSession(ctxWithTx, &session); err != nil {
			return err
		}

		// create activity
		msg := fmt.Sprintf("Login from %s", session.UserAgent)
		activity := model.UserActivity{
			ActorUserID: user.UserUUID,
			SessionID:   &body.SessionID,
			IPAddress:   session.IPAddress,
			UserAgent:   session.UserAgent,
			Activity:    "LOGIN",
			Message:     &msg,
		}

		return u.repoUser.CreateActivity(ctxWithTx, &activity)
	})

	if err != nil {
		logger.Error("Login-Transaction", err.Error())
		return nil, nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	return &tokenKey, &tokenRefresh, nil
}

func (u *usecase) Refresh(ctx context.Context, refreshToken string, actor request.UserActor) (*string, *string, error) {
	body, err := u.token.VerifyRefresh(refreshToken)
	if err != nil {
		return nil, nil, response.APIError{
			StatusCode: http.StatusUnauthorized,
			Message:    "Invalid refresh token",
		}
	}

	user, err := u.repoUser.FindUserByUserUUID(ctx, body.User)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, response.APIError{
			StatusCode: http.StatusNotFound,
			Message:    "User Not Found",
		}
	}
	if err != nil {
		logger.Error("Refresh-FindUserByUserUUID", err.Error())

		return nil, nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	if !user.Active {
		return nil, nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "User Not Active",
		}
	}

	claims := token.ClaimsFromUserToken(user)

	tokenKey, err := u.token.GenerateToken(claims, u.exp.Token, body.Session)
	if err != nil {
		logger.Error("Refresh-GenerateToken", err.Error())

		return nil, nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed GenerateToken",
		}
	}

	tokenRefresh, err := u.token.GenerateRefresh(claims, u.exp.Refresh, body.Session)
	if err != nil {
		logger.Error("Refresh-GenerateRefresh", err.Error())

		return nil, nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed GenerateRefresh",
		}
	}

	now := time.Now().UTC()
	err = u.repoUser.DoTransaction(ctx, func(ctxWithTx context.Context, _ *gorm.DB) error {
		if err := u.repoUser.UpdateSessionLastSeen(ctxWithTx, body.Session, now, actor.IPAddress, actor.UserAgent); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		logger.Error("Refresh-Transaction", err.Error())

		return nil, nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	return &tokenKey, &tokenRefresh, nil
}

func (u *usecase) Logout(ctx context.Context, actor request.UserActor) error {
	now := time.Now().UTC()
	err := u.repoUser.DoTransaction(ctx, func(ctxWithTx context.Context, _ *gorm.DB) error {
		if err := u.repoUser.UpdateSessionLogout(ctxWithTx, actor.SessionID, now, actor.IPAddress, actor.UserAgent); err != nil {
			return err
		}

		msg := fmt.Sprintf("Logout from %s", actor.UserAgent)
		activity := model.UserActivity{
			ActorUserID: actor.UserID,
			SessionID:   &actor.SessionID,
			IPAddress:   actor.IPAddress,
			UserAgent:   actor.UserAgent,
			Activity:    "LOGOUT",
			CreatedAt:   now,
			Message:     &msg,
		}

		return u.repoUser.CreateActivity(ctxWithTx, &activity)
	})

	if err != nil {
		logger.Error("Logout-Transaction", err.Error())

		return response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	return nil
}
