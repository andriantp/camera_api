package app

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"camera_api/app/handler/ecamera"
	"camera_api/app/handler/esystem"
	"camera_api/app/handler/euser"
	internalMiddleware "camera_api/app/middleware"
	"camera_api/pkg/utils/token"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const allowedOrigin = "https://app.example.com"

func newCORSServer() *echo.Echo {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(productionCORS([]string{allowedOrigin})))
	e.PATCH("/user/profile", func(c echo.Context) error { return c.NoContent(http.StatusOK) })
	return e
}

func preflight(e *echo.Echo, origin, method, headers string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodOptions, "/user/profile", nil)
	req.Header.Set(echo.HeaderOrigin, origin)
	req.Header.Set(echo.HeaderAccessControlRequestMethod, method)
	req.Header.Set(echo.HeaderAccessControlRequestHeaders, headers)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	return rec
}

func TestProductionCORSPreflight(t *testing.T) {
	rec := preflight(newCORSServer(), allowedOrigin, http.MethodPatch, "Content-Type, Authorization, X-Client-Type")

	if got := rec.Header().Get(echo.HeaderAccessControlAllowOrigin); got != allowedOrigin {
		t.Fatalf("Allow-Origin = %q, want %q", got, allowedOrigin)
	}

	methods := rec.Header().Get(echo.HeaderAccessControlAllowMethods)
	if !strings.Contains(methods, http.MethodPatch) {
		t.Errorf("Allow-Methods = %q, want it to include PATCH", methods)
	}

	headers := strings.ToLower(rec.Header().Get(echo.HeaderAccessControlAllowHeaders))
	if !strings.Contains(headers, strings.ToLower(HeaderClientType)) {
		t.Errorf("Allow-Headers = %q, want it to include %s", headers, HeaderClientType)
	}
}

func TestProductionCORSRejectsUnknownOrigin(t *testing.T) {
	rec := preflight(newCORSServer(), "https://evil.example.com", http.MethodPatch, "X-Client-Type")

	if got := rec.Header().Get(echo.HeaderAccessControlAllowOrigin); got != "" {
		t.Fatalf("Allow-Origin = %q for an unknown origin, want none", got)
	}
}

// TestProductionCORSAllowsEveryRouteMethod registers the real API routes and
// fails if any route uses a method that browsers would be blocked from
// sending.
func TestProductionCORSAllowsEveryRouteMethod(t *testing.T) {
	e := echo.New()
	middle := internalMiddleware.NewMiddleware(token.Token{}, nil)
	rate := internalMiddleware.NewRateLimiter()
	validate := validator.New()

	// Handlers only register routes here; the nil usecases are never called.
	euser.NewHandler(e, middle, rate, validate, nil)
	esystem.NewHandler(e, middle, nil)
	ecamera.NewHandler(e, middle, rate, validate, nil)

	allowed := productionCORS(nil).AllowMethods

	checked := 0

	for _, route := range e.Routes() {
		// Echo adds internal "echo_route_not_found" entries for groups.
		if strings.HasPrefix(route.Method, "echo_") {
			continue
		}

		checked++

		if !slices.Contains(allowed, route.Method) {
			t.Errorf("%s %s uses a method missing from CORS AllowMethods %v", route.Method, route.Path, allowed)
		}
	}

	if checked < 40 {
		t.Fatalf("only %d routes checked; expected the full API to be registered", checked)
	}
}
