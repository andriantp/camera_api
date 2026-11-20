package middleware

import (
	"context"

	"camera_api/pkg/utils/token"
)

// SessionValidator checks that a session is still active (not logged out,
// user active and not deleted) and returns the user's current role name.
// It must return a response.APIError (401) when the session is not usable.
type SessionValidator interface {
	ValidateSession(ctx context.Context, sessionUUID, userUUID string) (string, error)
}

type Middleware struct {
	token    token.Token
	sessions SessionValidator
}

func NewMiddleware(jwt token.Token, sessions SessionValidator) *Middleware {
	return &Middleware{
		token:    jwt,
		sessions: sessions,
	}
}
