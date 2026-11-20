package app

import (
	"camera_api/app/usecase/usystem"
	"camera_api/pkg/adapter/camera"
	"camera_api/pkg/adapter/database/postgres"
	"camera_api/pkg/adapter/s3"
	"camera_api/pkg/utils/token"
	"context"
	"time"
)

type (
	Setting struct {
		Build    usystem.Build
		Database postgres.Setting
		Server   Server
		Token    token.Token
		Camera   camera.Setting
		S3       s3.Setting
	}

	Server struct {
		Port          uint16
		ReadTimeout   time.Duration
		WriteTimeout  time.Duration
		IdleTimeout   time.Duration
		JWTEXPToken   time.Duration
		JWTEXPRefresh time.Duration
		Cors          bool
		AllowOrigins  []string
		Swagger       bool
		// TrustProxy enables reading the client IP from X-Forwarded-For.
		// Only enable it when the API sits behind a reverse proxy on a
		// private/loopback network; otherwise clients can spoof their IP.
		TrustProxy bool
	}
)

type appStruct struct {
	setting Setting
}

type Application interface {
	Run(ctx context.Context) error
}

func NewApp(setting Setting) Application {
	return &appStruct{
		setting: setting,
	}
}
