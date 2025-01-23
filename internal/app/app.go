package app

import (
	"log/slog"
	"time"

	grpcapp "github.com/hesoyamTM/apphelper-sso/internal/app/grpc"
	"github.com/hesoyamTM/apphelper-sso/internal/app/key"
	"github.com/hesoyamTM/apphelper-sso/internal/clients/report"
	"github.com/hesoyamTM/apphelper-sso/internal/services/auth"
	"github.com/hesoyamTM/apphelper-sso/internal/storage/psql"
	"github.com/hesoyamTM/apphelper-sso/internal/storage/redis"
)

type App struct {
	KGApp   *key.App
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
	ReportClient *report.Client
}

func New(log *slog.Logger, grpcOpts GrpcOpts, psqlOpts PsqlOpts, rOpts RedisOpts, clients Clients, updInterval time.Duration) *App {
	psqlDB, err := psql.New(psqlOpts.User, psqlOpts.Password, psqlOpts.Host, psqlOpts.DB, psqlOpts.Port)
	if err != nil {
		panic(err)
	}

	rDB := redis.New(rOpts.Host, rOpts.Password, rOpts.Port)

	kgApp := key.New(log, updInterval, *clients.ReportClient)

	authService := auth.New(
		log,
		psqlDB,
		rDB,
		grpcOpts.AccessTokenTTL,
		grpcOpts.RefreshTokenTTL,
		kgApp.GetPrivateKeyChan(),
	)

	grpcApp := grpcapp.New(log, authService, grpcOpts.Host, grpcOpts.Port)

	return &App{
		KGApp:   kgApp,
		GRPCApp: grpcApp,
	}
}
