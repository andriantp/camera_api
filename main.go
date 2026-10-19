// @title Camera API
// @version 1.0
// @description Camera management API
// @host localhost:8085
// @BasePath /
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"camera_api/app"
	"camera_api/pkg/adapter/vault"
	"camera_api/pkg/utils/logger"
)

/*
author : @tp
*/

var (
	version = "v.1.0"
	release = "2026-08-31T13:22:02+07:00"
)

func init() {
	log.Println("========   camera_api  =========")
	log.Printf("version : %s", version)
	log.Printf("release : %s", release)
	log.Println("========    initialization    =========")
}

func main() {
	rootCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	env := os.Getenv("ENV")
	conf, apps, err := vault.ReadEnv(env)
	if err != nil {
		logger.Error("main", fmt.Sprintf("ReadEnv: %v", err))
		panic(err)
	}

	apps.Build.Version = version
	apps.Build.Release = release

	log.Println("============= start =============")
	if conf.General.Core > 0 {
		runtime.GOMAXPROCS(conf.General.Core)
	}

	application := app.NewApp(*apps)
	if err := application.Run(ctx); err != nil {
		logger.Error("main", fmt.Sprintf("FAILED on Run: %v", err))
		panic(err)
	}
}
