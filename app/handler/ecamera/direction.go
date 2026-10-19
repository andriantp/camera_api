package ecamera

import (
	"camera_api/app/middleware"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"
	"net/http"

	"github.com/labstack/echo/v4"
)

// @Summary Get directions
// @Tags Direction
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Router /camera/directions [get]
func (h *handler) GetDirections(c echo.Context) error {
	result, err := h.ucaseCamera.GetDirections(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse("success", result))
}

// @Summary Create direction
// @Tags Direction
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param request body request.CreateDirection true "CreateDirection request"
// @Router /camera/directions [post]
func (h *handler) CreateDirection(c echo.Context) error {
	var req request.CreateDirection

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

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	result, err := h.ucaseCamera.CreateDirection(c.Request().Context(), *actor, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, response.SuccessResponse("Direction created successfully", result))
}

// @Summary Get direction
// @Tags Direction
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Direction UUID"
// @Router /camera/directions/{directionUUID} [get]
func (h *handler) GetDirection(c echo.Context) error {
	directionUUID := c.Param("directionUUID")
	if directionUUID == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "UUID Direction is required",
		}
	}

	result, err := h.ucaseCamera.GetDirectionByUUID(c.Request().Context(), directionUUID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse("success", result))
}

// @Summary Update direction
// @Tags Direction
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Direction UUID"
// @Param request body request.UpdateDirection true "UpdateDirection request"
// @Router /camera/directions/{directionUUID} [patch]
func (h *handler) UpdateDirection(c echo.Context) error {
	directionUUID := c.Param("directionUUID")
	if directionUUID == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "UUID Direction is required",
		}
	}

	var req request.UpdateDirection
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

	if req.Name == nil && req.Left == nil && req.Right == nil && req.Up == nil && req.Down == nil {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "At least one field is required",
		}
	}

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	if err := h.ucaseCamera.UpdateDirection(c.Request().Context(), *actor, directionUUID, req); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse(
		"Direction updated successfully",
		map[string]string{
			"uuid": directionUUID,
		},
	))
}
