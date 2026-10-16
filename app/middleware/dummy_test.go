package middleware_test

import (
	"camera_api/app/middleware"
	"camera_api/pkg/utils/token"
	"testing"
	"time"
)

func newTestMiddleware() *middleware.Middleware {
	jwt := token.Token{
		Secret: []byte("secret-test-camera-api"),
		Issuer: "camera",
	}

	return middleware.NewMiddleware(jwt)
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
