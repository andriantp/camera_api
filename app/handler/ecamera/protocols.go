package ecamera

import (
	"camera_api/app/middleware"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"
	"net/http"

	"github.com/labstack/echo/v4"
)

// @Summary Get protocols
// @Tags Protocol
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Router /camera/protocols [get]
func (h *handler) GetProtocols(c echo.Context) error {
	ctx := c.Request().Context()

	protocols, err := h.ucaseCamera.GetProtocols(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.SuccessResponse("success", protocols))
}

// @Summary Create protocol
// @Tags Protocol
// @Security BearerAuth
// @Param X-Client-Type header string true "Client type"
// @Param request body request.CreateProtocol true "CreateProtocol request"
// @Router /camera/protocols [post]
func (h *handler) CreateProtocol(c echo.Context) error {
	ctx := c.Request().Context()

	actor, err := middleware.GetUserActor(c)
	if err != nil {
		return err
	}

	var req request.CreateProtocol
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

	protocolUUID, err := h.ucaseCamera.CreateProtocol(ctx, *actor, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated,
		response.SuccessResponse(
			"Protocol created successfully",
			map[string]string{
				"uuid": *protocolUUID,
			},
		),
	)
}
