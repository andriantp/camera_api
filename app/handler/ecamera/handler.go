package ecamera

import (
	"camera_api/app/middleware"
	"camera_api/app/usecase/ucamera"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type handler struct {
	validator   *validator.Validate
	ucaseCamera ucamera.Usecase
}

func NewHandler(
	e *echo.Echo,
	middle *middleware.Middleware,
	rateLimit *middleware.RateLimiter,
	validator *validator.Validate,
	ucaseCamera ucamera.Usecase,
) {
	handler := &handler{
		validator:   validator,
		ucaseCamera: ucaseCamera,
	}

	readLimiter := rateLimit.Middleware(
		10,
		time.Minute,
		func(c echo.Context) string {
			return "camera:" + c.Param("uuid")
		},
	)

	createLimiter := rateLimit.Middleware(
		10,
		time.Minute,
		func(c echo.Context) string {
			claims, err := middleware.GetClaims(c)
			if err != nil {
				return "unknown"
			}

			return "user:" + claims.User
		},
	)

	updateLimiter := rateLimit.Middleware(
		20,
		time.Minute,
		func(c echo.Context) string {
			claims, err := middleware.GetClaims(c)
			if err != nil {
				return "unknown"
			}

			return "user:" + claims.User
		},
	)

	moveLimiter := rateLimit.Middleware(
		10,
		time.Second,
		func(c echo.Context) string {
			return "camera:" + c.Param("uuid")
		},
	)

	downloadLimiter := rateLimit.Middleware(
		2,
		time.Minute,
		func(c echo.Context) string {
			return "camera:" + c.Param("cameraUUID")
		},
	)

	group := e.Group("/camera", middleware.ErrorMiddleware, middle.AuthMiddleware)

	// =========================================================
	// Protocol
	// =========================================================
	group.GET("/protocols", handler.GetProtocols, middle.RoleMiddleware("super_admin"))
	group.POST("/protocols", handler.CreateProtocol, middle.RoleMiddleware("super_admin"))

	// =========================================================
	// PanTilt
	// =========================================================
	group.GET("/pan-tilts", handler.GetPanTilts)
	group.POST("/pan-tilts", handler.CreatePanTilt, middle.RoleMiddleware("super_admin", "admin"), createLimiter)
	group.GET("/pan-tilts/:pantiltUUID", handler.GetPanTilt)
	group.PATCH("/pan-tilts/:pantiltUUID", handler.UpdatePanTilt, middle.RoleMiddleware("super_admin", "admin"), updateLimiter)

	// =========================================================
	// Direction
	// =========================================================
	group.GET("/directions", handler.GetDirections)
	group.POST("/directions", handler.CreateDirection, middle.RoleMiddleware("super_admin", "admin"), createLimiter)
	group.GET("/directions/:directionUUID", handler.GetDirection)
	group.PATCH("/directions/:directionUUID", handler.UpdateDirection, middle.RoleMiddleware("super_admin", "admin"), updateLimiter)

	// =========================================================
	// Camera
	// =========================================================
	group.GET("", handler.GetCameras)
	group.POST("", handler.CreateCamera, middle.RoleMiddleware("super_admin", "admin"), createLimiter)
	group.POST("/assign", handler.AssignCamera, middle.RoleMiddleware("super_admin", "admin"), updateLimiter)
	group.GET("/protocol/:protocolUUID", handler.GetCamerasByProtocol)
	// Camera detail
	group.GET("/:cameraUUID", handler.GetCamera)
	// Camera credential
	group.GET("/:cameraUUID/credential", handler.GetCameraWithCredential, middle.RoleMiddleware("super_admin"))

	// =========================================================
	// Configuration
	// =========================================================
	group.GET("/:cameraUUID/configuration", handler.GetConfiguration, middle.RoleMiddleware("super_admin", "admin"), readLimiter)
	group.PATCH("/:cameraUUID/configuration", handler.UpdateConfiguration, middle.RoleMiddleware("super_admin", "admin"), updateLimiter)
	group.PATCH("/:cameraUUID/sn", handler.UpdateCameraSN, middle.RoleMiddleware("super_admin", "admin"), updateLimiter)
	group.PATCH("/:cameraUUID/status", handler.UpdateCameraStatus, middle.RoleMiddleware("super_admin", "admin"), updateLimiter)
	group.PATCH("/:cameraUUID/credential", handler.UpdateCameraCredential, middle.RoleMiddleware("super_admin", "admin"), updateLimiter)
	group.PATCH("/:cameraUUID/direction", handler.UpdateCameraDirection, middle.RoleMiddleware("super_admin", "admin"), updateLimiter)

	// =========================================================
	// Device
	// =========================================================
	group.GET("/:cameraUUID/capabilities", handler.GetCapabilities, middle.RoleMiddleware("super_admin"), readLimiter)
	group.GET("/:cameraUUID/capabilities/ptz", handler.GetPTZCapability, middle.RoleMiddleware("super_admin"), readLimiter)
	group.GET("/:cameraUUID/capabilities/storage", handler.GetStorage, middle.RoleExceptMiddleware("user"), readLimiter)

	// =========================================================
	// Movements
	// =========================================================
	group.GET("/:cameraUUID/move/position", handler.GetPosition, middle.RoleExceptMiddleware("user"), readLimiter)
	group.POST("/:cameraUUID/move/calibration", handler.CalibrationMove, middle.RoleMiddleware("super_admin"), moveLimiter)
	group.POST("/:cameraUUID/move/:direction", handler.MoveTo, middle.RoleExceptMiddleware("user"), moveLimiter)
	group.POST("/:cameraUUID/move/home", handler.MoveToHome, middle.RoleExceptMiddleware("user"), moveLimiter)

	// =========================================================
	// Preset
	// =========================================================
	group.GET("/:cameraUUID/preset", handler.GetPresets, readLimiter)
	group.POST("/:cameraUUID/preset/:presetID", handler.MoveToPreset, middle.RoleExceptMiddleware("user"), moveLimiter)

	// =========================================================
	// Zoom
	// =========================================================
	group.GET("/:cameraUUID/zoom", handler.GetZoom, readLimiter)
	group.POST("/:cameraUUID/zoom/:valueZoom", handler.SetZoom, middle.RoleExceptMiddleware("user"), moveLimiter)

	// =========================================================
	// Content
	// =========================================================
	// content
	group.GET("/:cameraUUID/content", handler.GetContent, middle.RoleExceptMiddleware("user"), readLimiter)
	group.GET("/:cameraUUID/content/:filename/info", handler.FindContent, middle.RoleExceptMiddleware("user"), readLimiter)

	// download
	group.GET("/:cameraUUID/content/:filename", handler.DownloadContent, middle.RoleMiddleware("super_admin"), downloadLimiter)
	group.GET("/:cameraUUID/content/:filename/s3", handler.DownloadContentS3, middle.RoleExceptMiddleware("user"), downloadLimiter)

	// snapshot
	group.GET("/:cameraUUID/snapshot", handler.GetSnapshot, middle.RoleMiddleware("super_admin"), downloadLimiter)
	group.GET("/:cameraUUID/snapshot/s3", handler.GetSnapshotS3, middle.RoleExceptMiddleware("user"), downloadLimiter)
}
