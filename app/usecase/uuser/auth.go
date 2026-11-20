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
	"sync"
	"time"

	"gorm.io/gorm"
)

type Auth interface {
	Login(ctx context.Context, body request.Login) (*string, *string, error)
	Logout(ctx context.Context, actor request.UserActor) error
	Refresh(ctx context.Context, refreshToken string, actor request.UserActor) (*string, *string, error)
	ValidateSession(ctx context.Context, sessionUUID, userUUID string) (string, error)
}

// ValidateSession checks that a session is still usable and returns the
// user's current role name. It is called by AuthMiddleware on every request.
func (u *usecase) ValidateSession(ctx context.Context, sessionUUID, userUUID string) (string, error) {
	session, err := u.repoUser.FindActiveSession(ctx, sessionUUID, userUUID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", response.APIError{
			StatusCode: http.StatusUnauthorized,
			Message:    "Unauthorized",
		}
	}
	if err != nil {
		logger.Error("ValidateSession-FindActiveSession", err.Error())

		return "", response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	return session.RoleName, nil
}

func (u *usecase) GetExpire() Expire {
	return Expire{
		Token:   u.exp.Token,
		Refresh: u.exp.Refresh,
	}
}

// maxLoginFailed is how many consecutive failed logins are allowed; the next
// one deactivates the account.
const maxLoginFailed = 3

// errInvalidCredentials is returned for every failed login (unknown user,
// wrong password, lockout) so responses don't reveal which usernames exist.
var errInvalidCredentials = response.APIError{
	StatusCode: http.StatusUnauthorized,
	Message:    "Invalid username or password",
}

var (
	dummyHashOnce sync.Once
	dummyHash     string
)

// checkDummyPassword runs a bcrypt comparison against a throwaway hash so a
// login for an unknown username takes as long as one for an existing user.
// The hash is created with password.CreatePassword, so it follows the same
// bcrypt cost as real passwords.
func checkDummyPassword(plain string) {
	dummyHashOnce.Do(func() {
		hash, err := password.CreatePassword("camera-api-dummy-password")
		if err != nil {
			logger.Error("Login-DummyHash", err.Error())
			return
		}

		dummyHash = hash
	})

	if dummyHash != "" {
		_, _ = password.CheckPassword(plain, dummyHash)
	}
}

func (u *usecase) Login(ctx context.Context, body request.Login) (*string, *string, error) {
	user, err := u.repoUser.FindUserByUserName(ctx, body.UserName)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		checkDummyPassword(body.Password)

		return nil, nil, errInvalidCredentials
	} else if err != nil {
		logger.Error("Login-FindUserByUserName", err.Error())
		return nil, nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}

	_, err = password.CheckPassword(body.Password, user.Password)

	// Only reveal that an account is inactive to someone who knows its
	// password. Failed attempts on an inactive account are not counted, so
	// they don't re-run the lockout or flood the activity log.
	if !user.Active {
		if err != nil {
			return nil, nil, errInvalidCredentials
		}

		return nil, nil, response.APIError{
			StatusCode: http.StatusForbidden,
			Message:    "User Not Active",
		}
	}

	if err != nil {
		err = u.repoUser.DoTransaction(ctx, func(ctxWithTx context.Context, _ *gorm.DB) error {
			failed, err := u.repoUser.IncrementLoginFailed(ctxWithTx, user.UserUUID)
			if err != nil {
				return err
			}

			if failed <= maxLoginFailed {
				return nil
			}

			// Concurrent failed logins may all pass the limit; only the one
			// that actually deactivates the user revokes sessions and logs it.
			deactivated, err := u.repoUser.DeactivateIfActive(ctxWithTx, user.UserUUID)
			if err != nil || !deactivated {
				return err
			}

			if err := u.repoUser.RevokeUserSessions(ctxWithTx, user.UserUUID, time.Now().UTC(), ""); err != nil {
				return err
			}

			msg := fmt.Sprintf(
				"Deactivated user after %d failed login attempts from %s",
				failed,
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

		// The lockout uses the same response as any failed login; otherwise
		// a few wrong guesses would confirm that the username exists. The
		// user sees "User Not Active" on their next login with the right
		// password.
		return nil, nil, errInvalidCredentials
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

	// Reject refresh tokens of sessions that were logged out or revoked.
	_, err = u.repoUser.FindActiveSession(ctx, body.Session, body.User)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, response.APIError{
			StatusCode: http.StatusUnauthorized,
			Message:    "Invalid refresh token",
		}
	}
	if err != nil {
		logger.Error("Refresh-FindActiveSession", err.Error())

		return nil, nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
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
