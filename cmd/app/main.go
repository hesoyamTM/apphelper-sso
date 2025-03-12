package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/hesoyamTM/apphelper-sso/internal/app"
	"github.com/hesoyamTM/apphelper-sso/internal/config"
	"github.com/hesoyamTM/apphelper-sso/pkg/logger"
)

func main() {
	ctx := context.Background()
	cfg := config.MustLoad()
	ctx, err := logger.New(ctx, cfg.Env)
	if err != nil {
		panic(err)
	}

	log := logger.GetLoggerFromCtx(ctx)

	log.Debug(ctx, "logger working")

	gOpts := app.GrpcOpts{
		Host:            cfg.Grpc.Host,
		Port:            cfg.Grpc.Port,
		AccessTokenTTL:  cfg.Grpc.AccessTokenTTL,
		RefreshTokenTTL: cfg.Grpc.RefreshTokenTTL,
	}
	pOpts := app.PsqlOpts{
		Host:     cfg.Psql.Host,
		Port:     cfg.Psql.Port,
		User:     cfg.Psql.User,
		Password: cfg.Psql.Password,
		DB:       cfg.Psql.DB,
	}
	rOpts := app.RedisOpts{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
	}

	application := app.New(ctx, gOpts, pOpts, rOpts)
	go application.GRPCApp.MustRun(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	application.GRPCApp.Stop(ctx)
	log.Info(ctx, "application stopped")
}
