package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"camera_api/app/middleware"
	"camera_api/pkg/utils/token"

	"github.com/labstack/echo/v4"
)

func TestRoleMiddlewareAllowed(t *testing.T) {
	e := echo.New()
	m := middleware.NewMiddleware(token.Token{})

	nextCalled := false

	next := func(c echo.Context) error {
		nextCalled = true
		return c.NoContent(http.StatusOK)
	}

	handler := m.RoleMiddleware("admin")(next)

	req := httptest.NewRequest(
		http.MethodGet,
		"/users",
		nil,
	)

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	c.Set(
		string(middleware.TokenClaims),
		&token.CustomClaims{
			Role: "admin",
		},
	)

	err := handler(c)
	if err != nil {
		t.Fatalf("RoleMiddleware: %v", err)
	}

	if !nextCalled {
		t.Fatal("expected next handler to be called")
	}

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"unexpected status: got %d, want %d",
			rec.Code,
			http.StatusOK,
		)
	}
}

func TestRoleMiddlewareForbidden(t *testing.T) {
	e := echo.New()
	m := middleware.NewMiddleware(token.Token{})

	nextCalled := false

	next := func(c echo.Context) error {
		nextCalled = true
		return c.NoContent(http.StatusOK)
	}

	handler := m.RoleMiddleware("super_admin")(next)

	req := httptest.NewRequest(
		http.MethodGet,
		"/users",
		nil,
	)

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	c.Set(
		string(middleware.TokenClaims),
		&token.CustomClaims{
			Role: "user",
		},
	)

	err := handler(c)
	if err == nil {
		t.Fatal("expected forbidden error")
	}

	if nextCalled {
		t.Fatal("next handler should not be called")
	}
}
