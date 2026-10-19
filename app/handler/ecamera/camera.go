package ecamera

import (
	"camera_api/app/middleware"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"
	"net/http"

	"github.com/labstack/echo/v4"
)

// @Summary Get cameras
// @Tags Camera
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Router /camera [get]
func (h *handler) GetCameras(c echo.Context) error {
	ctx := c.Request().Context()

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	cameras, err := h.ucaseCamera.GetCameras(ctx, *actor)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse("success", cameras))
}

// @Summary Create camera
// @Tags Camera
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param request body request.CreateCamera true "CreateCamera request"
// @Router /camera [post]
func (h *handler) CreateCamera(c echo.Context) error {
	ctx := c.Request().Context()

	var req request.CreateCamera
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

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	cameraUUID, err := h.ucaseCamera.CreateCamera(ctx, *actor, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated,
		response.SuccessResponse(
			"Camera created successfully",
			map[string]string{
				"uuid": *cameraUUID,
			},
		),
	)
}

// @Summary Assign camera to user
// @Tags Camera
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param request body request.AssignCameraRequest true "Assign camera request"
// @Router /camera/assign [post]
func (h *handler) AssignCamera(c echo.Context) error {
	ctx := c.Request().Context()

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	var req request.AssignCameraRequest

	if err := c.Bind(&req); err != nil {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid request body",
		}
	}

	if err := h.validator.Struct(req); err != nil {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid request",
		}
	}

	if err := h.ucaseCamera.AssignCamera(ctx, *actor, req); err != nil {
		return err
	}

	return c.JSON(
		http.StatusOK,
		response.SuccessResponse(
			"Camera assigned successfully",
			nil,
		),
	)
}

// @Summary Get cameras by protocol
// @Tags Camera
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param protocolUUID path string true "Protocol UUID"
// @Router /camera/protocol/{protocolUUID} [get]
func (h *handler) GetCamerasByProtocol(c echo.Context) error {
	ctx := c.Request().Context()

	protocolUUID := c.Param("protocolUUID")
	if protocolUUID == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "protocol UUID is required",
		}
	}

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	cameras, err := h.ucaseCamera.GetCamerasByProtocol(ctx, *actor, protocolUUID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse("success", cameras))
}

// @Summary Get camera
// @Tags Camera
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Camera UUID"
// @Router /camera/{cameraUUID} [get]
func (h *handler) GetCamera(c echo.Context) error {
	ctx := c.Request().Context()

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

	camera, err := h.ucaseCamera.GetCamera(ctx, *actor, cameraUUID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse("success", camera))
}

// @Summary Get camera credential
// @Tags Camera
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Camera UUID"
// @Router /camera/{cameraUUID}/credential [get]
func (h *handler) GetCameraWithCredential(c echo.Context) error {
	ctx := c.Request().Context()

	cameraUUID := c.Param("cameraUUID")
	if cameraUUID == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "camera UUID is required",
		}
	}

	camera, err := h.ucaseCamera.GetCameraWithCredential(ctx, cameraUUID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse("success", camera))
}
