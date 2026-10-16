package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"camera_api/app/middleware"

	"github.com/labstack/echo/v4"
)

func TestAuthMiddlewareAPI(t *testing.T) {
	e := echo.New()
	m := newTestMiddleware()

	accessToken := newTestToken(t)

	nextCalled := false

	next := func(c echo.Context) error {
		nextCalled = true
		return c.NoContent(http.StatusOK)
	}

	handler := m.AuthMiddleware(next)

	req := httptest.NewRequest(
		http.MethodGet,
		"/users",
		nil,
	)

	req.Header.Set("X-Client-Type", "api")
	req.Header.Set(
		"Authorization",
		"Bearer "+accessToken,
	)

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("AuthMiddleware: %v", err)
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

	claims := c.Get(string(middleware.TokenClaims))
	if claims == nil {
		t.Fatal("expected token claims in echo context")
	}
}

func TestAuthMiddlewareMissingAuthorization(t *testing.T) {
	e := echo.New()
	m := newTestMiddleware()

	nextCalled := false

	next := func(c echo.Context) error {
		nextCalled = true
		return c.NoContent(http.StatusOK)
	}

	handler := m.AuthMiddleware(next)

	req := httptest.NewRequest(
		http.MethodGet,
		"/users",
		nil,
	)

	req.Header.Set("X-Client-Type", "api")

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	err := handler(c)
	if err == nil {
		t.Fatal("expected unauthorized error")
	}

	if nextCalled {
		t.Fatal("next handler should not be called")
	}
}

func TestAuthMiddlewareInvalidToken(t *testing.T) {
	e := echo.New()
	m := newTestMiddleware()

	nextCalled := false

	next := func(c echo.Context) error {
		nextCalled = true
		return c.NoContent(http.StatusOK)
	}

	handler := m.AuthMiddleware(next)

	req := httptest.NewRequest(
		http.MethodGet,
		"/users",
		nil,
	)

	req.Header.Set("X-Client-Type", "api")
	req.Header.Set("Authorization", "Bearer invalid-token")

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	err := handler(c)
	if err == nil {
		t.Fatal("expected unauthorized error")
	}

	if nextCalled {
		t.Fatal("next handler should not be called")
	}
}

func TestAuthMiddlewareWeb(t *testing.T) {
	e := echo.New()
	m := newTestMiddleware()

	accessToken := newTestToken(t)

	nextCalled := false

	next := func(c echo.Context) error {
		nextCalled = true
		return c.NoContent(http.StatusOK)
	}

	handler := m.AuthMiddleware(next)

	req := httptest.NewRequest(
		http.MethodGet,
		"/users",
		nil,
	)

	req.Header.Set("X-Client-Type", "web")

	req.AddCookie(&http.Cookie{
		Name:  "access_token",
		Value: accessToken,
	})

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("AuthMiddleware: %v", err)
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

	claims := c.Get(string(middleware.TokenClaims))
	if claims == nil {
		t.Fatal("expected token claims in echo context")
	}
}

func TestAuthMiddlewareWebMissingCookie(t *testing.T) {
	e := echo.New()
	m := newTestMiddleware()

	nextCalled := false

	next := func(c echo.Context) error {
		nextCalled = true
		return c.NoContent(http.StatusOK)
	}

	handler := m.AuthMiddleware(next)

	req := httptest.NewRequest(
		http.MethodGet,
		"/users",
		nil,
	)

	req.Header.Set("X-Client-Type", "web")

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	err := handler(c)
	if err == nil {
		t.Fatal("expected unauthorized error")
	}

	if nextCalled {
		t.Fatal("next handler should not be called")
	}
}

func TestAuthMiddlewareUnsupportedClient(t *testing.T) {
	e := echo.New()
	m := newTestMiddleware()

	nextCalled := false

	next := func(c echo.Context) error {
		nextCalled = true
		return c.NoContent(http.StatusOK)
	}

	handler := m.AuthMiddleware(next)

	req := httptest.NewRequest(
		http.MethodGet,
		"/users",
		nil,
	)

	req.Header.Set("X-Client-Type", "mobile")

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	err := handler(c)
	if err == nil {
		t.Fatal("expected unauthorized error")
	}

	if nextCalled {
		t.Fatal("next handler should not be called")
	}
}