package ecamera

import (
	"camera_api/app/middleware"
	"camera_api/pkg/api/response"
	"net/http"

	"github.com/labstack/echo/v4"
)

// @Summary Get capabilities
// @Tags Device
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Camera UUID"
// @Router /camera/{cameraUUID}/capabilities [get]
func (h *handler) GetCapabilities(c echo.Context) error {
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

	result, err := h.ucaseCamera.GetCapabilities(c.Request().Context(), *actor, cameraUUID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse("success", result))
}

// @Summary Get capability PTZ
// @Tags Device
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Camera UUID"
// @Router /camera/{cameraUUID}/capabilities/ptz [get]
func (h *handler) GetPTZCapability(c echo.Context) error {
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

	result, err := h.ucaseCamera.GetPTZCapability(c.Request().Context(), *actor, cameraUUID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse("success", result))
}

// @Summary Get camera capability storage
// @Tags Device
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Camera UUID"
// @Router /camera/{cameraUUID}/capabilities/storage [get]
func (h *handler) GetStorage(c echo.Context) error {
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

	result, err := h.ucaseCamera.GetStorage(c.Request().Context(), *actor, cameraUUID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse("success", result))
}
