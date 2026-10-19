package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"camera_api/app/handler/ecamera"
	"camera_api/app/handler/esystem"
	"camera_api/app/handler/euser"
	"camera_api/app/usecase/ucamera"
	"camera_api/app/usecase/usystem"
	"camera_api/app/usecase/uuser"
	"camera_api/pkg/adapter/database/postgres"
	"camera_api/pkg/adapter/database/repository/user"
	"camera_api/pkg/adapter/s3"
	"camera_api/pkg/utils/logger"

	internalMiddleware "camera_api/app/middleware"
	adaptercamera "camera_api/pkg/adapter/camera"
	repocamera "camera_api/pkg/adapter/database/repository/camera"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	_ "camera_api/docs"

	echoSwagger "github.com/swaggo/echo-swagger"
)

func (a *appStruct) Run(ctx context.Context) error {
	db, err := postgres.NewPostgres(a.setting.Database)
	if err != nil {
		return fmt.Errorf("NewPostgres: %w", err)
	}

	logger.Info("Run", "success connect to database")

	// create validator
	validate := validator.New()

	// repository
	repoUser := user.NewRepository(db)
	repoCamera := repocamera.NewRepository(db)
	camera, err := adaptercamera.NewCamera(a.setting.Camera)
	if err != nil {
		return fmt.Errorf("NewCamera: %w", err)
	}

	// s3
	storage, err := s3.NewStore(ctx, a.setting.S3)
	if err != nil {
		return fmt.Errorf("NewStore: %w", err)
	}
	logger.Info("Run", "success connect to storage")

	// usecase
	uUser := uuser.NewUsecase(
		uuser.Expire{
			Token:   a.setting.Server.JWTEXPToken,
			Refresh: a.setting.Server.JWTEXPRefresh,
		},
		a.setting.Token,
		repoUser,
	)
	uSystem := usystem.NewUsecase(a.setting.Build)
	uCamera := ucamera.NewUsecase(repoUser, repoCamera, camera, storage)

	// internalMiddleware
	middle := internalMiddleware.NewMiddleware(a.setting.Token)
	rateLimit := internalMiddleware.NewRateLimiter()

	// Echo
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	logger.Info("Run", fmt.Sprintf("swagger:%v", a.setting.Server.Swagger))
	if a.setting.Server.Swagger {
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	// CORS
	if a.setting.Build.ENV != "production" {
		e.Use(middleware.CORS())
	} else if a.setting.Server.Cors {
		e.Use(middleware.CORSWithConfig(
			middleware.CORSConfig{
				AllowOriginFunc: func(origin string) (bool, error) {
					if origin == "" {
						return true, nil
					}

					for _, allowed := range a.setting.Server.AllowOrigins {
						if origin == allowed {
							return true, nil
						}
					}

					return false, nil
				},
				AllowMethods: []string{
					http.MethodGet,
					http.MethodPost,
					http.MethodPut,
					http.MethodDelete,
					http.MethodOptions,
				},
				AllowHeaders: []string{
					echo.HeaderOrigin,
					echo.HeaderContentType,
					echo.HeaderAccept,
					echo.HeaderAuthorization,
				},
				AllowCredentials: true,
			},
		))
	}

	// LOGGER
	e.Use(middleware.RequestLoggerWithConfig(
		middleware.RequestLoggerConfig{
			LogStatus:    true,
			LogMethod:    true,
			LogURI:       true,
			LogUserAgent: true,
			LogLatency:   true,
			LogValuesFunc: func(c echo.Context, values middleware.RequestLoggerValues) error {
				logEcho := fmt.Sprintf(
					"{status:%v} {method:%v} {latency:%v} {uri:%v} {user_agent:%v}",
					values.Status,
					values.Method,
					values.Latency,
					values.URI,
					values.UserAgent,
				)

				if values.Status >= 200 && values.Status <= 202 {
					logger.Trace("logEcho", logEcho)
				} else {
					logger.Error("logEcho", logEcho)
				}

				return nil
			},
		},
	))

	// Server timeout
	e.Server.ReadTimeout = a.setting.Server.ReadTimeout
	e.Server.WriteTimeout = a.setting.Server.WriteTimeout
	e.Server.IdleTimeout = a.setting.Server.IdleTimeout

	e.Use(middleware.Recover())

	// handlers
	euser.NewHandler(e, middle, rateLimit, validate, uUser)
	esystem.NewHandler(e, middle, uSystem)
	ecamera.NewHandler(e, middle, rateLimit, validate, uCamera)

	e.Server.Addr = fmt.Sprintf(":%d", a.setting.Server.Port)

	// server lifecycle
	errServer := make(chan error, 1)

	go func() {
		logger.Info("Run", fmt.Sprintf("server running on %s", e.Server.Addr))
		log.Println("============= run =============")

		err := e.StartServer(e.Server)

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errServer <- err
			return
		}

		errServer <- nil
	}()

	select {
	case <-ctx.Done():
		logger.Info("Run", fmt.Sprintf("shutdown triggered: %v", ctx.Err()))

		ctxShutdown, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		if err := e.Shutdown(ctxShutdown); err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}

		if err := <-errServer; err != nil {
			return fmt.Errorf("server error: %w", err)
		}

		logger.Info("Run", "server exited properly")

		return nil

	case err := <-errServer:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}

		return nil
	}
}
