package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/hesoyamTM/apphelper-sso/internal/app"
	"github.com/hesoyamTM/apphelper-sso/internal/clients/report"
	"github.com/hesoyamTM/apphelper-sso/internal/config"
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

	repClient, err := report.New(log, cfg.Report.Addr)
	if err != nil {
		panic(err)
	}

	cOpts := app.Clients{
		ReportClient: repClient,
	}

	application := app.New(log, gOpts, pOpts, rOpts, cOpts, cfg.KeysUpdateInterval)
	go application.GRPCApp.MustRun()
	go application.KGApp.Run()

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
