package ecamera

import (
	"camera_api/app/middleware"
	"camera_api/pkg/api/response"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// @Summary Snapshot direct
// @Tags Snapshot
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Camera UUID"
// @Router /camera/{cameraUUID}/snapshot [get]
func (h *handler) GetSnapshot(c echo.Context) error {
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

	filename, size, body, err := h.ucaseCamera.GetSnapshot(
		ctx,
		*actor,
		cameraUUID,
	)
	if err != nil {
		return err
	}
	defer body.Close()

	res := c.Response()
	res.Header().Set(echo.HeaderContentType, "image/png")

	if size != nil {
		res.Header().Set(
			echo.HeaderContentLength,
			strconv.FormatInt(*size, 10),
		)
	}

	res.Header().Set(
		echo.HeaderContentDisposition,
		fmt.Sprintf(`attachment; filename="%s"`, *filename),
	)

	return c.Stream(
		http.StatusOK,
		"image/png",
		body,
	)
}

// @Summary Snapshot s3
// @Tags Snapshot
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Camera UUID"
// @Router /camera/{cameraUUID}/snapshot/s3 [get]
func (h *handler) GetSnapshotS3(c echo.Context) error {
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

	result, err := h.ucaseCamera.GetSnapshotS3(
		ctx,
		*actor,
		cameraUUID,
	)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse("success", result))
}
