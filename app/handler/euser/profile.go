package euser

import (
	"camera_api/app/middleware"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"
	"net/http"

	"github.com/labstack/echo/v4"
)

// @Summary Get my profile
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Router /user/profile [get]
func (h *handler) GetMyProfile(c echo.Context) error {
	claims, err := middleware.GetClaims(c)
	if err != nil {
		return err
	}

	profile, err := h.ucaseUser.GetMyProfile(c.Request().Context(), claims.User)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse("success", profile))
}

// @Summary Update my profile
// @Tags Profile
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param request body request.UpdateProfile true "UpdateProfile request"
// @Router /user/profile [patch]
func (h *handler) UpdateMyProfile(c echo.Context) error {
	ctx := c.Request().Context()

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	var req request.UpdateProfile
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

	if req.Username == nil {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "username is required",
		}
	}

	if err := h.ucaseUser.UpdateMyProfile(ctx, *actor, req); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse("Profile updated successfully", nil))
}

// @Summary Change my password
// @Tags Profile
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param request body request.ChangePassword true "ChangePassword request"
// @Router /user/profile/password [patch]
func (h *handler) ChangeMyPassword(c echo.Context) error {
	ctx := c.Request().Context()

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	var req request.ChangePassword
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

	if err := h.ucaseUser.ChangeMyPassword(ctx, *actor, req); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse("Password changed successfully", nil))
}
