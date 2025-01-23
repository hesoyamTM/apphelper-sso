package main

import (
	"log/slog"
	"os"
	"os/signal"
	"sso/internal/app"
	"sso/internal/config"
	"syscall"
)

const (
	localEnv = "local"
	devEnv   = "dev"
	prodEnv  = "prod"
)

func main() {
	cfg := config.MustLoad()
	log := mustSetupLogger(cfg.Env)
	log.Debug("logger working")

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

	application := app.New(log, gOpts, pOpts, rOpts, cfg.KeysUpdateInterval)
	go application.GRPCApp.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	application.GRPCApp.Stop()
	application.KGApp.Stop()
	log.Info("application stopped")
}

func mustSetupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case localEnv:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case devEnv:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case prodEnv:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		panic("environment reading error")
	}

	return log
}
