package app

import (
	"context"

	grpcapp "github.com/hesoyamTM/apphelper-sso/internal/app/grpc"
	"github.com/hesoyamTM/apphelper-sso/internal/clients/report"
	"github.com/hesoyamTM/apphelper-sso/internal/clients/schedule"
	"github.com/hesoyamTM/apphelper-sso/internal/config"
	"github.com/hesoyamTM/apphelper-sso/internal/lib/jwt"
	"github.com/hesoyamTM/apphelper-sso/internal/migrations"
	"github.com/hesoyamTM/apphelper-sso/internal/services/auth"
	"github.com/hesoyamTM/apphelper-sso/internal/storage/psql"
	"github.com/hesoyamTM/apphelper-sso/internal/storage/redis"
	"github.com/hesoyamTM/apphelper-sso/pkg/metrics"
)

type App struct {
	GRPCApp *grpcapp.App
}

type Clients struct {
	ReportClient   *report.Client
	ScheduleClient *schedule.Client
}

func New(ctx context.Context, cfg *config.Config) *App {
	psqlDB, err := psql.New(ctx, cfg.Psql)
	if err != nil {
		panic(err)
	}

	if err := migrations.RunMigrations(ctx, cfg.Psql); err != nil {
		panic(err)
	}

	rDB := redis.New(ctx, cfg.Redis)

	privKey, err := jwt.DecodePrivateKey(cfg.PrivateKey)
	if err != nil {
		panic(err)
	}

	authService := auth.New(
		ctx,
		psqlDB,
		rDB,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
		privKey,
	)

	grpcMetrics, err := metrics.NewMetrics(ctx, cfg.Metrics)
	if err != nil {
		panic(err)
	}

	grpcApp := grpcapp.New(ctx, authService, grpcMetrics, cfg.Grpc)

	return &App{
		GRPCApp: grpcApp,
	}
}
