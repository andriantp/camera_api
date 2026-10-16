package middleware

import (
	"net/http"
	"strings"

	"camera_api/pkg/api/response"

	"github.com/labstack/echo/v4"
)

func (m *Middleware) RoleMiddleware(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, err := GetClaims(c)
			if err != nil {
				return err
			}

			for _, role := range roles {
				if strings.EqualFold(claims.Role, role) {
					return next(c)
				}
			}

			return response.APIError{
				StatusCode: http.StatusForbidden,
				Message:    "forbidden",
			}
		}
	}
}

func (m *Middleware) RoleExceptMiddleware(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, err := GetClaims(c)
			if err != nil {
				return err
			}

			for _, role := range roles {
				if strings.EqualFold(claims.Role, role) {
					return response.APIError{
						StatusCode: http.StatusForbidden,
						Message:    "forbidden",
					}
				}
			}

			return next(c)
		}
	}
}
