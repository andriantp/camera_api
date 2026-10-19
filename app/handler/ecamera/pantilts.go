package ecamera

import (
	"camera_api/app/middleware"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"
	"net/http"

	"github.com/labstack/echo/v4"
)

// @Summary Get pan tilts
// @Tags PanTilt
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Router /camera/pan-tilts [get]
func (h *handler) GetPanTilts(c echo.Context) error {
	result, err := h.ucaseCamera.GetPanTilts(
		c.Request().Context(),
	)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse("success", result))
}

// @Summary Create pan tilt
// @Tags PanTilt
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param request body request.CreatePanTilt true "CreatePanTilt request"
// @Router /camera/pan-tilts [post]
func (h *handler) CreatePanTilt(c echo.Context) error {
	var req request.CreatePanTilt

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

	result, err := h.ucaseCamera.CreatePanTilt(c.Request().Context(), *actor, req)
	if err != nil {
		return err
	}

	return c.JSON(
		http.StatusCreated,
		response.SuccessResponse("success", result),
	)
}

// @Summary Get pan tilt
// @Tags PanTilt
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Pantilt UUID"
// @Router /camera/pan-tilts/{pantiltUUID} [get]
func (h *handler) GetPanTilt(c echo.Context) error {
	panTiltUUID := c.Param("pantiltUUID")

	result, err := h.ucaseCamera.GetPanTiltByUUID(c.Request().Context(), panTiltUUID)
	if err != nil {
		return err
	}

	return c.JSON(
		http.StatusOK,
		response.SuccessResponse("success", result),
	)
}

// @Summary Update pan tilt
// @Tags PanTilt
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Pantilt UUID"
// @Param request body request.UpdatePanTilt true "UpdatePanTilt request"
// @Router /camera/pan-tilts/{pantiltUUID} [patch]
func (h *handler) UpdatePanTilt(c echo.Context) error {
	panTiltUUID := c.Param("pantiltUUID")
	if panTiltUUID == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "UUID PanTilt is required",
		}
	}

	var req request.UpdatePanTilt
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

	if req.Name == nil && req.Pan == nil && req.Tilt == nil && req.Duration == nil {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "At least one field is required",
		}
	}

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	if err := h.ucaseCamera.UpdatePanTilt(c.Request().Context(), *actor, panTiltUUID, req); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse(
		"PanTilt updated successfully",
		map[string]string{
			"uuid": panTiltUUID,
		},
	))
}
