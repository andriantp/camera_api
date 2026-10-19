package ecamera

import (
	"camera_api/app/middleware"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"
	"net/http"

	"github.com/labstack/echo/v4"
)

// @Summary Get Camera Configuration
// @Tags Configuration
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Camera UUID"
// @Router /camera/{cameraUUID}/configuration [get]
func (h *handler) GetConfiguration(c echo.Context) error {
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

	result, err := h.ucaseCamera.GetConfiguration(c.Request().Context(), *actor, cameraUUID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse("success", result))
}

// @Summary Update Camera Configuration
// @Tags Configuration
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Camera UUID"
// @Param request body request.UpdateCameraConfiguration true "UpdateCameraConfiguration request"
// @Router /camera/{cameraUUID}/configuration [patch]
func (h *handler) UpdateConfiguration(c echo.Context) error {
	cameraUUID := c.Param("cameraUUID")
	if cameraUUID == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "UUID camera is required",
		}
	}

	var req request.UpdateCameraConfiguration

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

	if err := h.ucaseCamera.UpdateConfiguration(c.Request().Context(), *actor, cameraUUID, req); err != nil {
		return err
	}

	return c.JSON(
		http.StatusOK,
		response.SuccessResponse(
			"Camera channel updated successfully",
			map[string]string{
				"uuid": cameraUUID,
			},
		),
	)
}

// @Summary Update camera serial number
// @Tags Configuration
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Camera UUID"
// @Param request body request.UpdateCameraSN true "UpdateCameraSN request"
// @Router /camera/{cameraUUID}/sn [patch]
func (h *handler) UpdateCameraSN(c echo.Context) error {
	ctx := c.Request().Context()
	cameraUUID := c.Param("cameraUUID")
	if cameraUUID == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "UUID camera is required",
		}
	}

	var req request.UpdateCameraSN
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

	if err := h.ucaseCamera.UpdateCameraSN(ctx, *actor, cameraUUID, req); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse(
		"Camera serial number updated successfully",
		map[string]string{
			"uuid": cameraUUID,
		},
	))
}

// @Summary Update camera status
// @Tags Configuration
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Camera UUID"
// @Param request body request.UpdateCameraStatus true "UpdateCameraStatus request"
// @Router /camera/{cameraUUID}/status [patch]
func (h *handler) UpdateCameraStatus(c echo.Context) error {
	ctx := c.Request().Context()

	cameraUUID := c.Param("cameraUUID")
	if cameraUUID == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "UUID camera is required",
		}
	}

	var req request.UpdateCameraStatus
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

	if err := h.ucaseCamera.UpdateCameraStatus(ctx, *actor, cameraUUID, req); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse(
		"Camera status updated successfully",
		map[string]string{
			"uuid": cameraUUID,
		},
	))
}

// @Summary Update camera credential
// @Tags Configuration
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Camera UUID"
// @Param request body request.UpdateCameraCredential true "UpdateCameraCredential request"
// @Router /camera/{cameraUUID}/credential [patch]
func (h *handler) UpdateCameraCredential(c echo.Context) error {
	ctx := c.Request().Context()

	cameraUUID := c.Param("cameraUUID")
	if cameraUUID == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "UUID camera is required",
		}
	}

	var req request.UpdateCameraCredential

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

	if req.Host == nil && req.Username == nil && req.Password == nil {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "At least one credential field is required",
		}
	}

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	if err := h.ucaseCamera.UpdateCameraCredential(ctx, *actor, cameraUUID, req); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse(
		"Camera credential updated successfully",
		map[string]string{
			"uuid": cameraUUID,
		},
	))
}

// @Summary Update camera direction
// @Tags Configuration
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Camera UUID"
// @Param request body request.UpdateCameraDirection true "UpdateCameraDirection request"
// @Router /camera/{cameraUUID}/direction [patch]
func (h *handler) UpdateCameraDirection(c echo.Context) error {
	cameraUUID := c.Param("cameraUUID")
	if cameraUUID == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "UUID camera is required",
		}
	}

	var req request.UpdateCameraDirection

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

	if err := h.ucaseCamera.UpdateCameraDirection(c.Request().Context(), *actor, cameraUUID, req); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse(
		"Camera direction updated successfully",
		map[string]string{
			"uuid": cameraUUID,
		},
	))
}
