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

func TestErrorMiddlewareSuccess(t *testing.T) {
	e := echo.New()

	handler := middleware.ErrorMiddleware(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("ErrorMiddleware: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"unexpected status: got %d, want %d",
			rec.Code,
			http.StatusOK,
		)
	}
}

func TestErrorMiddlewareAPIError(t *testing.T) {
	e := echo.New()

	handler := middleware.ErrorMiddleware(func(c echo.Context) error {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "bad request",
		}
	})

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("ErrorMiddleware: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"unexpected status: got %d, want %d",
			rec.Code,
			http.StatusBadRequest,
		)
	}
}

func TestErrorMiddlewareUnknownError(t *testing.T) {
	e := echo.New()

	handler := middleware.ErrorMiddleware(func(c echo.Context) error {
		return errors.New("database connection failed")
	})

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatalf("ErrorMiddleware: %v", err)
	}

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"unexpected status: got %d, want %d",
			rec.Code,
			http.StatusInternalServerError,
		)
	}
}
