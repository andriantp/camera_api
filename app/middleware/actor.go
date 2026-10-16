package middleware

import (
	"camera_api/pkg/api/request"

	"github.com/labstack/echo/v4"
)

func GetUserActor(c echo.Context) (*request.UserActor, error) {
	claims, err := GetClaims(c)
	if err != nil {
		return nil, err
	}

	return &request.UserActor{
		UserID:     claims.User,
		SessionID:  claims.Session,
		Role:       claims.Role,
		IPAddress:  c.RealIP(),
		UserAgent:  c.Request().UserAgent(),
		ClientType: c.Request().Header.Get("X-Client-Type"),
	}, nil
}
