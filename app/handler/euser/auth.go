package euser

import (
	"camera_api/app/middleware"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"
	"camera_api/pkg/utils/token"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// @Summary Login
// @Tags Authentication
// @Accept json
// @Produce json
// @Security []
// @Param X-Client-Type header string true "Client type"
// @Param request body request.Login true "Login request"
// @Router /user/login [post]
func (h *handler) Login(c echo.Context) error {
	ctx := c.Request().Context()

	var req request.Login
	if err := c.Bind(&req); err != nil {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "The given data was invalid",
		}
	}
	if err := h.validator.Struct(req); err != nil {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid request",
		}
	}

	req.ClientType = c.Request().Header.Get("X-Client-Type")
	req.SessionID = uuid.New().String()
	req.IPAddress = c.RealIP()
	req.UserAgent = c.Request().UserAgent()

	access, refresh, err := h.ucaseUser.Login(ctx, req)
	if err != nil {
		return err
	}

	return h.authResponse(
		c,
		req.ClientType,
		*access,
		*refresh,
		"login success",
	)
}

// @Summary Refresh token
// @Tags Authentication
// @Produce json
// @Param X-Client-Type header string true "Client type"
// @Param request body token.Refresh true "Refresh request"
// @Router /user/refresh [post]
func (h *handler) Refresh(c echo.Context) error {
	ctx := c.Request().Context()
	actor := request.UserActor{
		IPAddress:  c.RealIP(),
		UserAgent:  c.Request().UserAgent(),
		ClientType: c.Request().Header.Get("X-Client-Type"),
	}

	refreshToken, err := h.getClientRefreshToken(c, actor.ClientType)
	if err != nil {
		return err
	}

	access, refresh, err := h.ucaseUser.Refresh(ctx, refreshToken, actor)
	if err != nil {
		return err
	}

	return h.authResponse(
		c,
		actor.ClientType,
		*access,
		*refresh,
		"refresh success",
	)
}

// @Summary Logout
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Router /user/logout [post]
func (h *handler) Logout(c echo.Context) error {
	ctx := c.Request().Context()

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	if err := h.ucaseUser.Logout(ctx, *actor); err != nil {
		return err
	}

	switch actor.ClientType {
	case "api":
		res := response.SuccessResponse(
			"logout success",
			map[string]string{
				"message": "token cleared on client side",
			},
		)

		return c.JSON(http.StatusOK, res)

	default:
		expiredCookie := func(name string) *http.Cookie {
			return &http.Cookie{
				Name:     name,
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
				Expires:  time.Unix(0, 0),
				MaxAge:   -1,
			}
		}

		c.SetCookie(expiredCookie("access"))
		c.SetCookie(expiredCookie("refresh"))

		res := response.SuccessResponse(
			"logout success",
			map[string]string{
				"message": "cookies cleared",
			},
		)

		return c.JSON(http.StatusOK, res)
	}
}

func (h *handler) getClientRefreshToken(c echo.Context, clientType string) (string, error) {
	switch clientType {
	case "api":
		var req token.Refresh

		if err := c.Bind(&req); err != nil {
			return "", response.APIError{
				StatusCode: http.StatusBadRequest,
				Message:    "The given data was invalid",
			}
		}

		if req.Refresh == "" {
			return "", response.APIError{
				StatusCode: http.StatusUnauthorized,
				Message:    "Refresh unauthorized",
			}
		}

		return req.Refresh, nil

	default:
		cookie, err := c.Cookie("refresh")
		if err != nil || cookie.Value == "" {
			return "", response.APIError{
				StatusCode: http.StatusUnauthorized,
				Message:    "Refresh unauthorized",
			}
		}

		return cookie.Value, nil
	}
}

func (h *handler) authResponse(c echo.Context, clientType string, access string, refresh string, message string) error {
	switch clientType {
	case "api":
		c.Response().Header().Set(
			"Authorization",
			"Bearer "+access,
		)

		res := response.SuccessResponse(
			message,
			map[string]string{
				"access":  access,
				"refresh": refresh,
			},
		)

		return c.JSON(http.StatusOK, res)

	default:
		expire := h.ucaseUser.GetExpire()

		c.SetCookie(&http.Cookie{
			Name:     "access",
			Value:    access,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   int(expire.Token.Seconds()),
		})

		c.SetCookie(&http.Cookie{
			Name:     "refresh",
			Value:    refresh,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   int(expire.Refresh.Seconds()),
		})

		res := response.SuccessResponse(
			message,
			map[string]string{
				"message": "cookie set",
			},
		)

		return c.JSON(http.StatusOK, res)
	}
}
