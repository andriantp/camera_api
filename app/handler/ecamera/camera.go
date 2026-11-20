package ecamera

import (
	"camera_api/app/middleware"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"
	"net/http"

	"github.com/google/uuid"
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

// @Summary Grant camera access
// @Description Gives an operator or user access to a camera, or re-enables a revoked access. Idempotent: granting an active access again returns 200 without changes.
// @Tags Camera
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Camera UUID"
// @Param userUUID path string true "User UUID"
// @Router /camera/{cameraUUID}/access/{userUUID} [put]
func (h *handler) GrantCameraAccess(c echo.Context) error {
	cameraUUID, userUUID, err := accessParams(c)
	if err != nil {
		return err
	}

	return h.grantCameraAccess(c, cameraUUID, userUUID)
}

// @Summary Assign camera to user (deprecated)
// @Description Deprecated: use PUT /camera/{cameraUUID}/access/{userUUID}.
// @Tags Camera
// @Security BearerAuth
// @Deprecated
// @Param X-Client-Type header string true "Client type"
// @Param request body request.AssignCameraRequest true "Assign camera request"
// @Router /camera/assign [post]
func (h *handler) AssignCamera(c echo.Context) error {
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

	return h.grantCameraAccess(c, req.CameraUUID, req.UserUUID)
}

func (h *handler) grantCameraAccess(c echo.Context, cameraUUID, userUUID string) error {
	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	changed, err := h.ucaseCamera.AssignCamera(c.Request().Context(), *actor, cameraUUID, userUUID)
	if err != nil {
		return err
	}

	message := "Camera access granted successfully"
	if !changed {
		message = "Camera access already granted"
	}

	return c.JSON(http.StatusOK, response.SuccessResponse(message, nil))
}

// @Summary Revoke camera access
// @Description Deactivates a user's access to a camera; PUT on the same path re-enables it. Idempotent: revoking an already revoked access returns 200 without changes.
// @Tags Camera
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Camera UUID"
// @Param userUUID path string true "User UUID"
// @Router /camera/{cameraUUID}/access/{userUUID} [delete]
func (h *handler) RevokeCameraAccess(c echo.Context) error {
	cameraUUID, userUUID, err := accessParams(c)
	if err != nil {
		return err
	}

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	changed, err := h.ucaseCamera.RevokeCameraAccess(c.Request().Context(), *actor, cameraUUID, userUUID)
	if err != nil {
		return err
	}

	message := "Camera access revoked successfully"
	if !changed {
		message = "Camera access already revoked"
	}

	return c.JSON(http.StatusOK, response.SuccessResponse(message, nil))
}

// accessParams reads and validates the camera and user UUIDs from the path.
// Malformed UUIDs are rejected here; PostgreSQL would fail them with a 500.
func accessParams(c echo.Context) (string, string, error) {
	cameraUUID := c.Param("cameraUUID")
	userUUID := c.Param("userUUID")

	if _, err := uuid.Parse(cameraUUID); err != nil {
		return "", "", response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid camera UUID",
		}
	}

	if _, err := uuid.Parse(userUUID); err != nil {
		return "", "", response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid user UUID",
		}
	}

	return cameraUUID, userUUID, nil
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
