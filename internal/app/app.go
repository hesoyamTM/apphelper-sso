package app

import (
	"context"

	grpcapp "github.com/hesoyamTM/apphelper-sso/internal/app/grpc"
	"github.com/hesoyamTM/apphelper-sso/internal/clients/redpanda"
	"github.com/hesoyamTM/apphelper-sso/internal/config"
	"github.com/hesoyamTM/apphelper-sso/internal/lib/jwt"
	"github.com/hesoyamTM/apphelper-sso/internal/migrations"
	"github.com/hesoyamTM/apphelper-sso/internal/services/auth"
	"github.com/hesoyamTM/apphelper-sso/internal/storage/psql"
	"github.com/hesoyamTM/apphelper-sso/internal/storage/redis"
)

const migrationsDir = "migrations"

type App struct {
	GRPCApp        *grpcapp.App
	RedpandaClient *redpanda.RedPandaClient
}

func New(ctx context.Context, cfg *config.Config) *App {
	psqlDB, err := psql.New(ctx, cfg.Psql)
	if err != nil {
		panic(err)
	}

	migrationCfg := migrations.Config{
		Host:     cfg.Psql.Host,
		Port:     cfg.Psql.Port,
		User:     cfg.Psql.User,
		Password: cfg.Psql.Password,
		DB:       cfg.Psql.DB,
	}

	if err := migrations.RunMigrations(ctx, migrationCfg, migrationsDir); err != nil {
		panic(err)
	}

	rDB := redis.New(ctx, cfg.Redis)

	privKey, err := jwt.DecodePrivateKey(cfg.PrivateKey)
	if err != nil {
		panic(err)
	}

	redpandaClient, err := redpanda.NewRedPandaClient(ctx, cfg.Redpanda)
	if err != nil {
		panic(err)
	}

	authService := auth.New(
		ctx,
		redpandaClient,
		psqlDB,
		rDB,
		rDB,
		rDB,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
		cfg.CodeTTL,
		cfg.TokenTTL,
		privKey,
	)

	if err != nil {
		panic(err)
	}

	grpcApp := grpcapp.New(ctx, authService, cfg.Grpc)

	return &App{
		GRPCApp:        grpcApp,
		RedpandaClient: redpandaClient,
	}
}
