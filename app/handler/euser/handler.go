package euser

import (
	"camera_api/app/middleware"
	"camera_api/app/usecase/uuser"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type handler struct {
	validator *validator.Validate
	ucaseUser uuser.Usecase
}

func NewHandler(e *echo.Echo,
	middle *middleware.Middleware,
	rateLimit *middleware.RateLimiter,
	validator *validator.Validate,
	ucaseUser uuser.Usecase) {
	handler := &handler{
		validator: validator,
		ucaseUser: ucaseUser,
	}

	group := e.Group("/user", middleware.ErrorMiddleware)

	loginLimiter := rateLimit.Middleware(
		5,
		time.Minute,
		func(c echo.Context) string {
			return c.RealIP()
		},
	)

	refreshLimiter := rateLimit.Middleware(
		30,
		time.Minute,
		func(c echo.Context) string {
			return c.RealIP()
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

	// =========================================================
	// Authentication
	// =========================================================
	group.POST("/login", handler.Login, loginLimiter)
	group.POST("/refresh", handler.Refresh, refreshLimiter)
	group.POST("/logout", handler.Logout, middle.AuthMiddleware)

	// =========================================================
	// My Profile
	// =========================================================
	group.GET("/profile", handler.GetMyProfile, middle.AuthMiddleware)
	group.PATCH("/profile", handler.UpdateMyProfile, middle.AuthMiddleware)
	group.PATCH("/profile/password", handler.ChangeMyPassword, middle.AuthMiddleware, updateLimiter)

	// =========================================================
	// User Management
	// =========================================================
	group.GET(
		"",
		handler.GetUsers,
		middle.AuthMiddleware,
		middle.RoleMiddleware("super_admin", "admin"),
	)

	group.GET(
		"/roles",
		handler.GetRoles,
		middle.AuthMiddleware,
		middle.RoleMiddleware("super_admin", "admin"),
	)

	group.GET(
		"/:userUUID/profile",
		handler.GetUserProfile,
		middle.AuthMiddleware,
		middle.RoleMiddleware("super_admin", "admin"),
	)

	group.POST(
		"",
		handler.CreateUser,
		middle.AuthMiddleware,
		middle.RoleMiddleware("super_admin", "admin"),
		createLimiter,
	)

	group.PATCH(
		"/:userUUID/profile",
		handler.UpdateUserProfile,
		middle.AuthMiddleware,
		middle.RoleMiddleware("super_admin", "admin"),
		updateLimiter,
	)

	group.POST(
		"/:userUUID/password/reset",
		handler.ResetPassword,
		middle.AuthMiddleware,
		middle.RoleMiddleware("super_admin"),
	)

	group.PATCH(
		"/:userUUID/status",
		handler.UpdateUserStatus,
		middle.AuthMiddleware,
		middle.RoleMiddleware("super_admin"),
	)

	group.DELETE(
		"/:userUUID",
		handler.DeleteUser,
		middle.AuthMiddleware,
		middle.RoleMiddleware("super_admin"),
	)

	// =========================================================
	// Activity
	// =========================================================
	group.GET(
		"/activity",
		handler.GetMyActivities,
		middle.AuthMiddleware,
	)

	group.GET(
		"/:userUUID/activity",
		handler.GetUserActivities,
		middle.AuthMiddleware,
		middle.RoleMiddleware("super_admin"),
	)
}
