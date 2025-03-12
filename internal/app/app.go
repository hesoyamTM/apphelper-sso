package app

import (
	"context"
	"time"

	grpcapp "github.com/hesoyamTM/apphelper-sso/internal/app/grpc"
	"github.com/hesoyamTM/apphelper-sso/internal/clients/report"
	"github.com/hesoyamTM/apphelper-sso/internal/clients/schedule"
	"github.com/hesoyamTM/apphelper-sso/internal/services/auth"
	"github.com/hesoyamTM/apphelper-sso/internal/storage/psql"
	"github.com/hesoyamTM/apphelper-sso/internal/storage/redis"
)

type App struct {
	GRPCApp *grpcapp.App
}

type GrpcOpts struct {
	Host            string
	Port            int
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type PsqlOpts struct {
	Host     string
	Port     int
	User     string
	Password string
	DB       string
}

type RedisOpts struct {
	Host     string
	Port     int
	Password string
}

type Clients struct {
	ReportClient   *report.Client
	ScheduleClient *schedule.Client
}

func New(ctx context.Context, grpcOpts GrpcOpts, psqlOpts PsqlOpts, rOpts RedisOpts) *App {
	psqlDB, err := psql.New(psqlOpts.User, psqlOpts.Password, psqlOpts.Host, psqlOpts.DB, psqlOpts.Port)
	if err != nil {
		panic(err)
	}

	rDB := redis.New(rOpts.Host, rOpts.Password, rOpts.Port)

	authService := auth.New(
		ctx,
		psqlDB,
		rDB,
		grpcOpts.AccessTokenTTL,
		grpcOpts.RefreshTokenTTL,
	)

	grpcCfg := grpcapp.Config{
		Host: grpcOpts.Host,
		Port: grpcOpts.Port,
	}

	grpcApp := grpcapp.New(ctx, authService, grpcCfg)

	return &App{
		GRPCApp: grpcApp,
	}
}
