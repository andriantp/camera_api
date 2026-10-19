package ecamera

import (
	"camera_api/app/middleware"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

// @Summary Get content available
// @Tags Content
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Camera UUID"
// @Param from query string false "Start date (RFC3339, timezone-aware)" example(2026-08-18T00:00:00+07:00)
// @Param to query string false "End date (RFC3339, timezone-aware)" example(2026-08-18T23:59:59+07:00)
// @Param limit query string false "Limit"
// @Param offset query string false "Offset"
// @Router /camera/{cameraUUID}/content [get]
func (h *handler) GetContent(c echo.Context) error {
	ctx := c.Request().Context()

	cameraUUID := c.Param("cameraUUID")
	if cameraUUID == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "camera UUID is required",
		}
	}

	filter, err := h.parseContentFilter(c)
	if err != nil {
		return err
	}

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	result, err := h.ucaseCamera.GetContents(ctx, *actor, cameraUUID, filter)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.PaginationResponse(
		"success",
		response.Meta{
			TotalRecords: result.Content.TotalMatches,
			Limit:        filter.Limit,
			Offset:       filter.Offset,
			Returned:     result.Content.NumOfMatches,
		},
		result.Content.Matches,
	))
}

// @Summary Find content
// @Tags Content
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Camera UUID"
// @Param filename path string true "Filename"
// @Router /camera/{cameraUUID}/content/{filename}/info [get]
func (h *handler) FindContent(c echo.Context) error {
	ctx := c.Request().Context()

	cameraUUID := c.Param("cameraUUID")
	if cameraUUID == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "camera UUID is required",
		}
	}

	filename := c.Param("filename")
	if filename == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "filename is required",
		}
	}

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	result, err := h.ucaseCamera.FindContent(
		ctx,
		*actor,
		cameraUUID,
		filename,
	)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse("success", result))
}

// @Summary Download content
// @Tags Content
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Camera UUID"
// @Param filename path string true "Filename"
// @Router /camera/{cameraUUID}/content/{filename} [get]
func (h *handler) DownloadContent(c echo.Context) error {
	ctx := c.Request().Context()

	cameraUUID := c.Param("cameraUUID")
	if cameraUUID == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "camera UUID is required",
		}
	}

	filename := c.Param("filename")
	if filename == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "filename is required",
		}
	}

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	size, body, err := h.ucaseCamera.DownloadContent(
		ctx,
		*actor,
		cameraUUID,
		filename,
	)
	if err != nil {
		return err
	}
	defer body.Close()

	res := c.Response()
	res.Header().Set(echo.HeaderContentType, "video/mp4")
	if size != nil {
		res.Header().Set(echo.HeaderContentLength, strconv.FormatInt(*size, 10))
	}
	res.Header().Set(echo.HeaderContentDisposition, fmt.Sprintf(`attachment; filename="%s"`, filename))

	res.WriteHeader(http.StatusOK)

	return c.Stream(
		http.StatusOK,
		"video/mp4",
		body,
	)
	// written, err := io.Copy(res, body)
	// if err != nil {
	// 	logger.Error(
	// 		"DownloadContent-Stream",
	// 		fmt.Sprintf("written[%d]: %s", written, err.Error()),
	// 	)
	// 	return err
	// }
	// return nil
}

// @Summary Download content s3
// @Tags Content
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param cameraUUID path string true "Camera UUID"
// @Param filename path string true "Filename"
// @Router /camera/{cameraUUID}/content/{filename}/s3 [get]
func (h *handler) DownloadContentS3(c echo.Context) error {
	ctx := c.Request().Context()

	cameraUUID := c.Param("cameraUUID")
	if cameraUUID == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "camera UUID is required",
		}
	}

	filename := c.Param("filename")
	if filename == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "filename is required",
		}
	}

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	result, err := h.ucaseCamera.DownloadContentAtS3(
		ctx,
		*actor,
		cameraUUID,
		filename,
	)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse("success", result))
}

func (h *handler) parseContentFilter(c echo.Context) (request.SearchContent, error) {
	filter := request.SearchContent{

		Limit:  10,
		Offset: 0,
	}

	if value := c.QueryParam("from"); value != "" {
		from, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return filter, response.APIError{
				StatusCode: http.StatusBadRequest,
				Message:    "Invalid from format",
			}
		}

		filter.From = from
	}

	if value := c.QueryParam("to"); value != "" {
		to, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return filter, response.APIError{
				StatusCode: http.StatusBadRequest,
				Message:    "Invalid to format",
			}
		}

		filter.To = to
	}

	if value := c.QueryParam("limit"); value != "" {
		limit, err := strconv.Atoi(value)
		if err != nil || limit <= 0 {
			return filter, response.APIError{
				StatusCode: http.StatusBadRequest,
				Message:    "Invalid limit",
			}
		}

		if limit > 100 {
			limit = 100
		}

		filter.Limit = limit
	}

	if value := c.QueryParam("offset"); value != "" {
		offset, err := strconv.Atoi(value)
		if err != nil || offset < 0 {
			return filter, response.APIError{
				StatusCode: http.StatusBadRequest,
				Message:    "Invalid offset",
			}
		}

		filter.Offset = offset
	}

	return filter, nil
}
