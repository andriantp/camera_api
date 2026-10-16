package middleware

import (
	"camera_api/pkg/api/response"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type Key string

const TokenClaims Key = "tokenClaims"

func (m *Middleware) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		clientType := c.Request().Header.Get("X-Client-Type")

		var accessToken string
		switch strings.ToLower(clientType) {
		case "api":
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return unauthorizedResponse(c)
			}

			parts := strings.Fields(authHeader)

			if len(parts) != 2 ||
				!strings.EqualFold(parts[0], "Bearer") ||
				parts[1] == "" {
				return unauthorizedResponse(c)
			}

			accessToken = parts[1]

		case "web":
			cookie, err := c.Cookie("access")
			if err != nil || cookie.Value == "" {
				return unauthorizedResponse(c)
			}

			accessToken = cookie.Value

		default:
			return unauthorizedResponse(c)
		}

		claims, err := m.token.VerifyToken(accessToken)
		if err != nil {
			return unauthorizedResponse(c)
		}

		// Store claims in Echo context.
		c.Set("user", claims)

		// if use context with value
		/*
			ctxWith := context.WithValue(
					c.Request().Context(),
					TokenClaims,
					claims,
				)

				c.SetRequest(c.Request().WithContext(ctxWith))
		*/

		return next(c)
	}
}

func unauthorizedResponse(c echo.Context) error {
	return response.APIError{
		StatusCode: http.StatusUnauthorized,
		Message:    "Unauthorized",
	}
}
