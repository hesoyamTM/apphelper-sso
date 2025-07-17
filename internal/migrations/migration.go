package migrations

import (
	"context"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/hesoyamTM/apphelper-sso/pkg/logger"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DB       string
}

func RunMigrations(ctx context.Context, cfg Config, migrationsDir string) error {
	const op = "migrations.RunMigrations"
	log := logger.GetLoggerFromCtx(ctx)
	log.Info(ctx, "running migrations")

	m, err := migrate.New(
		fmt.Sprintf("file://migratinns"),
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DB),
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Info(ctx, "no migrations to run")
			return nil
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info(ctx, "migrations done")

	return nil
}
