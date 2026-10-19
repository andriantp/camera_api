package euser

import (
	"camera_api/app/middleware"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// @Summary Get users
// @Tags User Management
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Router /user [get]
func (h *handler) GetUsers(c echo.Context) error {
	ctx := c.Request().Context()
	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(c.QueryParam("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	users, err := h.ucaseUser.GetUsers(ctx, *actor, limit, offset)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.PaginationResponse("success", response.Meta{
		TotalRecords: users.Total,
		Limit:        limit,
		Offset:       offset,
		Returned:     users.Returned,
	}, users.Users))
}

// @Summary Get roles
// @Tags User Management
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Router /user/roles [get]
func (h *handler) GetRoles(c echo.Context) error {
	ctx := c.Request().Context()

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	roles, err := h.ucaseUser.GetRoles(ctx, *actor)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse("success", roles))
}

// @Summary Get user profile
// @Tags User Management
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param userUUID path string true "User UUID"
// @Router /user/{userUUID}/profile [get]
func (h *handler) GetUserProfile(c echo.Context) error {
	ctx := c.Request().Context()

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	userUUID := c.Param("userUUID")
	if userUUID == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "user UUID is required",
		}
	}

	profile, err := h.ucaseUser.GetUserProfile(ctx, *actor, userUUID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse("success", profile))
}

// @Summary Create user
// @Tags User Management
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param request body request.CreateUser true "CreateUser request"
// @Router /user [post]
func (h *handler) CreateUser(c echo.Context) error {
	ctx := c.Request().Context()

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	var req request.CreateUser
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

	userUUID, err := h.ucaseUser.CreateUser(ctx, *actor, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated,
		response.SuccessResponse(
			"User created successfully",
			map[string]string{
				"uuid": *userUUID,
			},
		),
	)
}

// @Summary Update user profile
// @Tags User Management
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param userUUID path string true "User UUID"
// @Param request body request.UpdateProfile true "UpdateProfile request"
// @Router /user/{userUUID}/profile [patch]
func (h *handler) UpdateUserProfile(c echo.Context) error {
	ctx := c.Request().Context()

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	userUUID := c.Param("userUUID")
	if userUUID == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "user UUID is required",
		}
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

	if req.Username == nil && req.Role == nil {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "No data to update",
		}
	}

	if err := h.ucaseUser.UpdateUserProfile(ctx, *actor, userUUID, req); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse("Profile updated successfully", nil))
}

// @Summary Reset user password
// @Tags User Management
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param userUUID path string true "User UUID"
// @Param request body request.ResetPassword true "ResetPassword request"
// @Router /user/{userUUID}/password/reset [post]
func (h *handler) ResetPassword(c echo.Context) error {
	ctx := c.Request().Context()

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	// Get target user UUID.
	userUUID := c.Param("userUUID")
	if userUUID == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "user UUID is required",
		}
	}

	// Bind request.
	var req request.ResetPassword
	if err := c.Bind(&req); err != nil {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid request",
		}
	}
	if err := h.validator.Struct(req); err != nil {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid request",
		}
	}

	if err := h.ucaseUser.ResetPassword(ctx, *actor, userUUID, req); err != nil {
		return err
	}

	return c.JSON(http.StatusOK,
		response.SuccessResponse(
			"Password reset successfully",
			nil,
		),
	)
}

// @Summary Update user status
// @Tags User Management
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param userUUID path string true "User UUID"
// @Param request body request.UpdateUserStatus true "UpdateUserStatus request"
// @Router /user/{userUUID}/status [patch]
func (h *handler) UpdateUserStatus(c echo.Context) error {
	ctx := c.Request().Context()

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	userUUID := c.Param("userUUID")
	if userUUID == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "user UUID is required",
		}
	}

	var req request.UpdateUserStatus
	if err := c.Bind(&req); err != nil {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid request",
		}
	}
	if err := h.validator.Struct(req); err != nil {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid request",
		}
	}

	if err := h.ucaseUser.UpdateUserStatus(ctx, *actor, userUUID, req); err != nil {
		return err
	}

	return c.JSON(http.StatusOK,
		response.SuccessResponse(
			"User status updated successfully",
			nil,
		),
	)
}

// @Summary Delete user
// @Tags User Management
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param userUUID path string true "User UUID"
// @Router /user/{userUUID} [delete]
func (h *handler) DeleteUser(c echo.Context) error {
	ctx := c.Request().Context()

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	userUUID := c.Param("userUUID")
	if userUUID == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "user UUID is required",
		}
	}

	if err := h.ucaseUser.DeleteUser(ctx, *actor, userUUID); err != nil {
		return err
	}

	return c.JSON(http.StatusOK,
		response.SuccessResponse(
			"User deleted successfully",
			nil,
		),
	)
}
