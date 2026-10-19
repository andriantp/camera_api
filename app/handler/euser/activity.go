package euser

import (
	"camera_api/app/middleware"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// @Summary Get my activities
// @Tags Activity
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param from query string false "Start date (RFC3339, timezone-aware)" example(2026-08-18T00:00:00+07:00)
// @Param to query string false "End date (RFC3339, timezone-aware)" example(2026-08-18T23:59:59+07:00)
// @Param limit query string false "Limit"
// @Param offset query string false "Offset"
// @Param order query string false "Order"
// @Router /user/activity [get]
func (h *handler) GetMyActivities(c echo.Context) error {
	ctx := c.Request().Context()

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	filter, err := h.parseActivityFilter(c)
	if err != nil {
		return err
	}

	result, err := h.ucaseUser.GetMyActivities(ctx, *actor, filter)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.PaginationResponse(
		"success",
		response.Meta{
			TotalRecords: result.Total,
			Limit:        filter.Limit,
			Offset:       filter.Offset,
			Returned:     result.Returned,
		},
		result.Activities,
	))
}

// @Summary Get user activities
// @Tags Activity
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param userUUID path string true "User UUID"
// @Param target query string false "Target Activity" example(all, user, camera)
// @Param from query string false "Start date (RFC3339, timezone-aware)" example(2026-08-18T00:00:00+07:00)
// @Param to query string false "End date (RFC3339, timezone-aware)" example(2026-08-18T23:59:59+07:00)
// @Param limit query string false "Limit"
// @Param offset query string false "Offset"
// @Param order query string false "Order"
// @Router /user/{userUUID}/activity [get]
func (h *handler) GetUserActivities(c echo.Context) error {
	ctx := c.Request().Context()

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	targetUser := c.Param("userUUID")
	if targetUser == "" {
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "user UUID is required",
		}
	}

	filter, err := h.parseActivityFilter(c)
	if err != nil {
		return err
	}

	switch filter.Target {
	case "all", "user", "camera":
	default:
		return response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid activity target",
		}
	}

	result, err := h.ucaseUser.GetUserActivities(ctx, *actor, targetUser, filter)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.PaginationResponse(
		"success",
		response.Meta{
			TotalRecords: result.Total,
			Limit:        filter.Limit,
			Offset:       filter.Offset,
			Returned:     result.Returned,
		},
		result.Activities,
	))
}

func (h *handler) parseActivityFilter(c echo.Context) (request.Pagination, error) {
	filter := request.Pagination{
		Target: "all",
		Limit:  20,
		Offset: 0,
		Order:  "desc",
	}

	if value := c.QueryParam("target"); value != "" {
		filter.Target = strings.ToLower(value)
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

	if value := c.QueryParam("order"); value != "" {
		if value != "asc" && value != "desc" {
			return filter, response.APIError{
				StatusCode: http.StatusBadRequest,
				Message:    "Invalid order",
			}
		}

		filter.Order = value
	}

	return filter, nil
}
