package middleware_test

import (
	"context"

	"camera_api/app/middleware"
	"camera_api/pkg/utils/token"
	"testing"
	"time"
)

// fakeSessions is a SessionValidator stub for tests.
type fakeSessions struct {
	role  string
	err   error
	calls int
}

func (f *fakeSessions) ValidateSession(_ context.Context, _, _ string) (string, error) {
	f.calls++
	if f.err != nil {
		return "", f.err
	}

	return f.role, nil
}

func newTestMiddleware() *middleware.Middleware {
	return newTestMiddlewareWithSessions(&fakeSessions{role: "admin"})
}

func newTestMiddlewareWithSessions(sessions middleware.SessionValidator) *middleware.Middleware {
	jwt := token.Token{
		Secret: []byte("secret-test-camera-api"),
		Issuer: "camera",
	}

	return middleware.NewMiddleware(jwt, sessions)
}

func newTestToken(t *testing.T) string {
	t.Helper()

	jwt := token.Token{
		Secret: []byte("secret-test-camera-api"),
		Issuer: "camera",
	}

	user := &token.Claims{
		UserUUID: "user-123",
		RoleName: "admin",
	}

	tokenString, err := jwt.GenerateToken(
		user,
		time.Hour,
		"session-123",
	)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	return tokenString
}
