package middleware

import (
	"errors"
	"net/http"

	"camera_api/pkg/api/response"

	"github.com/labstack/echo/v4"
)

func ErrorMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := next(c)
		if err == nil {
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
