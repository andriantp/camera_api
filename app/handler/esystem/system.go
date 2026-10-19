package esystem

import (
	"camera_api/app/middleware"
	"camera_api/app/usecase/usystem"
	"camera_api/pkg/api/response"
	"net/http"

	"github.com/labstack/echo/v4"
)

type handler struct {
	ucaseSystem usystem.Usecase
}

func NewHandler(e *echo.Echo,
	middle *middleware.Middleware,
	ucaseSystem usystem.Usecase) {
	handler := &handler{
		ucaseSystem: ucaseSystem,
	}

	group := e.Group("/system", middleware.ErrorMiddleware, middle.AuthMiddleware, middle.RoleMiddleware("super_admin"))

	group.GET("/version", handler.GetVersion)
}

// GetVersion
//
// @Summary Get application version
// @Tags System
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Router /system/version [get]
func (h *handler) GetVersion(c echo.Context) error {
	version := h.ucaseSystem.GetVersion()
	res := response.SuccessResponse(
		"success",
		version,
	)

	return c.JSON(http.StatusOK, res)
}
