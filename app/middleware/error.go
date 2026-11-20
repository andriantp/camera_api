package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"camera_api/pkg/api/response"
	"camera_api/pkg/utils/logger"

	"github.com/labstack/echo/v4"
)

func ErrorMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := next(c)
		if err == nil {
			return nil
		}

		// Once the status and part of the body are sent (e.g. a camera
		// connection drops mid-download), writing a JSON error would append
		// it to the streamed file. Log it and leave the response as is.
		if c.Response().Committed {
			logger.Error(
				"ErrorMiddleware",
				fmt.Sprintf("%s %s: error after response was committed: %v",
					c.Request().Method, c.Request().URL.Path, err),
			)

			return nil
		}

		var apiErr response.APIError

		if !errors.As(err, &apiErr) {
			resp := response.FailedResponse(
				"internal server error",
				nil,
			)

			return c.JSON(
				http.StatusInternalServerError,
				resp,
			)
		}

		resp := response.FailedResponse(
			apiErr.Message,
			apiErr.Cause,
		)

		return c.JSON(
			apiErr.StatusCode,
			resp,
		)
	}
}
