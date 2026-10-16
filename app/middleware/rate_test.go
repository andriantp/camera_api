package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"camera_api/app/middleware"

	"github.com/labstack/echo/v4"
)

func TestRateLimiterAllowed(t *testing.T) {
	e := echo.New()
	rateLimiter := middleware.NewRateLimiter()

	nextCalled := 0

	next := func(c echo.Context) error {
		nextCalled++
		return c.NoContent(http.StatusOK)
	}

	handler := rateLimiter.Middleware(
		3,
		time.Minute,
		func(c echo.Context) string {
			return c.RealIP() + ":user-123"
		},
	)(next)

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(
			http.MethodGet,
			"/",
			nil,
		)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)
		if err != nil {
			t.Fatalf("request %d: %v", i+1, err)
		}

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"request %d: got status %d, want %d",
				i+1,
				rec.Code,
				http.StatusOK,
			)
		}
	}

	if nextCalled != 3 {
		t.Fatalf(
			"next called %d times, want 3",
			nextCalled,
		)
	}
}

func TestRateLimiterExceeded(t *testing.T) {
	e := echo.New()
	rateLimiter := middleware.NewRateLimiter()

	nextCalled := 0

	next := func(c echo.Context) error {
		nextCalled++
		return c.NoContent(http.StatusOK)
	}

	handler := rateLimiter.Middleware(
		2,
		time.Minute,
		func(c echo.Context) string {
			return c.RealIP() + ":user-123"
		},
	)(next)

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(
			http.MethodGet,
			"/",
			nil,
		)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)

		if i < 2 {
			if err != nil {
				t.Fatalf("request %d: %v", i+1, err)
			}

			continue
		}

		if err == nil {
			t.Fatal("expected rate limit error")
		}

		if nextCalled != 2 {
			t.Fatalf(
				"next called %d times, want 2",
				nextCalled,
			)
		}
	}
}

func TestRateLimiterDifferentUser(t *testing.T) {
	e := echo.New()
	rateLimiter := middleware.NewRateLimiter()

	next := func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}

	handler := rateLimiter.Middleware(
		1,
		time.Minute,
		func(c echo.Context) string {
			return c.RealIP() + ":" + c.Request().Header.Get("X-User")
		},
	)(next)

	// User A
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User", "user-a")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler(c); err != nil {
		t.Fatalf("user A first request: %v", err)
	}

	// User A again -> should be blocked
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User", "user-a")

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	if err := handler(c); err == nil {
		t.Fatal("expected user A to be rate limited")
	}

	// User B -> should have its own limit
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User", "user-b")

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	if err := handler(c); err != nil {
		t.Fatalf("user B should not be rate limited: %v", err)
	}
}
func TestRateLimiterDifferentIP(t *testing.T) {
	e := echo.New()
	rateLimiter := middleware.NewRateLimiter()

	next := func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}

	handler := rateLimiter.Middleware(
		1,
		time.Minute,
		func(c echo.Context) string {
			return c.RealIP() + ":user-123"
		},
	)(next)

	// IP A
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler(c); err != nil {
		t.Fatalf("IP A first request: %v", err)
	}

	// IP B
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.2:1234"

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	if err := handler(c); err != nil {
		t.Fatalf("IP B should not be rate limited: %v", err)
	}
}

func TestRateLimiterWindowReset(t *testing.T) {
	e := echo.New()
	rateLimiter := middleware.NewRateLimiter()

	next := func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}

	handler := rateLimiter.Middleware(
		1,
		50*time.Millisecond,
		func(c echo.Context) string {
			return "user-123"
		},
	)(next)

	// First request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler(c); err != nil {
		t.Fatalf("first request: %v", err)
	}

	// Second request immediately
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	if err := handler(c); err == nil {
		t.Fatal("expected second request to be rate limited")
	}

	// Wait until window expires
	time.Sleep(60 * time.Millisecond)

	// Request after window
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	if err := handler(c); err != nil {
		t.Fatalf("request after window: %v", err)
	}
}

func TestRateLimiterInvalidConfiguration(t *testing.T) {
	e := echo.New()
	rateLimiter := middleware.NewRateLimiter()

	next := func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}

	tests := []struct {
		name   string
		limit  int
		window time.Duration
	}{
		{
			name:   "zero limit",
			limit:  0,
			window: time.Minute,
		},
		{
			name:   "negative limit",
			limit:  -1,
			window: time.Minute,
		},
		{
			name:   "zero window",
			limit:  5,
			window: 0,
		},
		{
			name:   "negative window",
			limit:  5,
			window: -time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := rateLimiter.Middleware(
				tt.limit,
				tt.window,
				func(c echo.Context) string {
					return "user-123"
				},
			)(next)

			req := httptest.NewRequest(
				http.MethodGet,
				"/",
				nil,
			)

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler(c)

			if err == nil {
				t.Fatal("expected configuration error")
			}
		})
	}
}

func TestRateLimiterEmptyKey(t *testing.T) {
	e := echo.New()
	rateLimiter := middleware.NewRateLimiter()

	next := func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}

	handler := rateLimiter.Middleware(
		5,
		time.Minute,
		func(c echo.Context) string {
			return ""
		},
	)(next)

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)

	if err == nil {
		t.Fatal("expected invalid key error")
	}
}