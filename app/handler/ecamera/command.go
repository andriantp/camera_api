package ecamera

import (
	"camera_api/app/middleware"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// @Summary Get camera position
// @Tags Movements
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Camera UUID"
// @Router /camera/{cameraUUID}/move/position [get]
func (h *handler) GetPosition(c echo.Context) error {
	cameraUUID := c.Param("cameraUUID")
	if cameraUUID == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "camera UUID is required",
		}
	}

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	result, err := h.ucaseCamera.GetPosition(c.Request().Context(), *actor, cameraUUID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse("success", result))
}

// @Summary Camera calibration move
// @Tags Movements
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Camera UUID"
// @Param request body request.RelativeMove true "RelativeMove request"
// @Router /camera/{cameraUUID}/move/calibration [post]
func (h *handler) CalibrationMove(c echo.Context) error {
	cameraUUID := c.Param("cameraUUID")
	if cameraUUID == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "camera UUID is required",
		}
	}

	var req request.RelativeMove
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

	if req.Pan == nil || req.Tilt == nil {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "Pan Tilt Invalid request",
		}
	}

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	result, err := h.ucaseCamera.CalibrationMove(c.Request().Context(), *actor, cameraUUID, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse("success", result))
}

// @Summary Move camera
// @Tags Movements
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Camera UUID"
// @Param direction path string true "direction"
// @Router /camera/{cameraUUID}/move/{direction} [post]
func (h *handler) MoveTo(c echo.Context) error {
	cameraUUID := c.Param("cameraUUID")
	if cameraUUID == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "UUID camera is required",
		}
	}

	direction := strings.ToUpper(c.Param("direction"))
	if direction == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "Direction is required",
		}
	}

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	status, err := h.ucaseCamera.MoveTo(c.Request().Context(), *actor, cameraUUID, direction)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse(
		"Camera moved successfully",
		map[string]any{
			"uuid":      cameraUUID,
			"direction": direction,
			"status":    *status,
		},
	))
}

// @Summary Move camera to home
// @Tags Movements
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Camera UUID"
// @Router /camera/{cameraUUID}/move/home [post]
func (h *handler) MoveToHome(c echo.Context) error {
	cameraUUID := c.Param("cameraUUID")
	if cameraUUID == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "camera UUID is required",
		}
	}

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	if err := h.ucaseCamera.MoveToHome(c.Request().Context(), *actor, cameraUUID); err != nil {
		return err
	}

	return c.JSON(
		http.StatusOK,
		response.SuccessResponse(
			"Camera moved to home successfully",
			map[string]string{
				"uuid": cameraUUID,
			},
		),
	)
}

// @Summary Get camera presets
// @Tags Preset
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Camera UUID"
// @Router /camera/{cameraUUID}/preset [get]
func (h *handler) GetPresets(c echo.Context) error {
	cameraUUID := c.Param("cameraUUID")
	if cameraUUID == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "camera UUID is required",
		}
	}

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	result, err := h.ucaseCamera.GetPresets(c.Request().Context(), *actor, cameraUUID)
	if err != nil {
		return err
	}

	return c.JSON(
		http.StatusOK,
		response.SuccessResponse("success", result),
	)
}

// @Summary Move camera to preset
// @Tags Preset
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Camera UUID"
// @Param presetID path string true "Preset ID"
// @Router /camera/{cameraUUID}/preset/{presetID} [post]
func (h *handler) MoveToPreset(c echo.Context) error {
	cameraUUID := c.Param("cameraUUID")
	if cameraUUID == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "camera UUID is required",
		}
	}

	presetID, err := strconv.Atoi(c.Param("presetID"))
	if err != nil || presetID < 1 {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid preset ID",
		}
	}

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	if err := h.ucaseCamera.MoveToPreset(c.Request().Context(), *actor, cameraUUID, presetID); err != nil {
		return err
	}

	return c.JSON(
		http.StatusOK,
		response.SuccessResponse(
			"Camera moved to preset successfully",
			map[string]interface{}{
				"uuid":      cameraUUID,
				"preset_id": presetID,
			},
		),
	)
}

// @Summary Get zoom
// @Tags Zoom
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Camera UUID"
// @Router /camera/{cameraUUID}/zoom [get]
func (h *handler) GetZoom(c echo.Context) error {
	cameraUUID := c.Param("cameraUUID")
	if cameraUUID == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "camera UUID is required",
		}
	}

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	result, err := h.ucaseCamera.GetZoom(c.Request().Context(), *actor, cameraUUID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse("success", result))
}

// @Summary Set zoom
// @Tags Zoom
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Camera UUID"
// @Param value path string true "Value Zoom"
// @Router /camera/{cameraUUID}/zoom/{valueZoom} [post]
func (h *handler) SetZoom(c echo.Context) error {
	cameraUUID := c.Param("cameraUUID")
	if cameraUUID == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "camera UUID is required",
		}
	}

	value := c.Param("valueZoom")
	if value == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "value zoom is required",
		}
	}
	val, err := strconv.ParseFloat(strings.ReplaceAll(value, " ", ""), 64)
	if err != nil || val < 10 {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid value zoom",
		}
	}

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	result, err := h.ucaseCamera.SetZoom(c.Request().Context(), *actor, cameraUUID, val)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse("success", result))
}
