package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"camera_api/app/middleware"
	"camera_api/pkg/api/response"

	"github.com/labstack/echo/v4"
)

func runAuth(t *testing.T, sessions *fakeSessions) (echo.Context, bool, error) {
	t.Helper()

	e := echo.New()
	m := newTestMiddlewareWithSessions(sessions)

	nextCalled := false
	handler := m.AuthMiddleware(func(c echo.Context) error {
		nextCalled = true
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Client-Type", "api")
	req.Header.Set("Authorization", "Bearer "+newTestToken(t))

	c := e.NewContext(req, httptest.NewRecorder())

	return c, nextCalled, handler(c)
}

func TestAuthMiddlewareActiveSession(t *testing.T) {
	sessions := &fakeSessions{role: "super_admin"}

	c, nextCalled, err := runAuth(t, sessions)
	if err != nil {
		t.Fatalf("AuthMiddleware: %v", err)
	}

	if !nextCalled {
		t.Fatal("expected next handler to be called")
	}

	if sessions.calls != 1 {
		t.Fatalf("ValidateSession called %d times, want 1", sessions.calls)
	}

	claims, err := middleware.GetClaims(c)
	if err != nil {
		t.Fatalf("GetClaims: %v", err)
	}

	// Token was issued with role "admin"; the current DB role must win.
	if claims.Role != "super_admin" {
		t.Fatalf("role = %q, want %q (from database)", claims.Role, "super_admin")
	}
}

func TestAuthMiddlewareRevokedSession(t *testing.T) {
	sessions := &fakeSessions{
		err: response.APIError{StatusCode: http.StatusUnauthorized, Message: "Unauthorized"},
	}

	_, nextCalled, err := runAuth(t, sessions)
	if nextCalled {
		t.Fatal("next handler must not be called for a revoked session")
	}

	var apiErr response.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("err = %v, want 401 APIError", err)
	}
}

func TestAuthMiddlewareSkipsSessionCheckForInvalidToken(t *testing.T) {
	e := echo.New()
	sessions := &fakeSessions{role: "admin"}
	m := newTestMiddlewareWithSessions(sessions)

	handler := m.AuthMiddleware(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Client-Type", "api")
	req.Header.Set("Authorization", "Bearer not-a-jwt")

	if err := handler(e.NewContext(req, httptest.NewRecorder())); err == nil {
		t.Fatal("expected unauthorized error")
	}

	if sessions.calls != 0 {
		t.Fatalf("ValidateSession called %d times for an invalid token, want 0", sessions.calls)
	}
}
