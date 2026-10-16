package middleware

import (
	"camera_api/pkg/api/response"
	"camera_api/pkg/utils/token"
	"net/http"

	"github.com/labstack/echo/v4"
)

func GetClaims(c echo.Context) (*token.CustomClaims, error) {
	value := c.Get("user")
	if value == nil {
		return nil, response.APIError{
			StatusCode: http.StatusUnauthorized,
			Message:    "Unauthorized",
		}
	}

	claims, ok := value.(*token.CustomClaims)
	if !ok {
		return nil, response.APIError{
			StatusCode: http.StatusUnauthorized,
			Message:    "Invalid claims",
		}
	}

	return claims, nil
}
