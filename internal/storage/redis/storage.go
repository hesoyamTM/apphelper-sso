package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hesoyamTM/apphelper-sso/internal/storage"

	"github.com/redis/go-redis/v9"
)

type Storage struct {
	client *redis.Client
}

type RedisConfig struct {
	Host     string `yaml:"host" env-required:"true" env:"REDIS_HOST"`
	Port     int    `yaml:"port" env-required:"true" env:"REDIS_PORT"`
	Password string `yaml:"password" env-required:"true" env:"REDIS_PASSWORD"`
}

func New(ctx context.Context, cfg RedisConfig) *Storage {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       0,
	})

	return &Storage{
		client: client,
	}
}

func (s *Storage) CreateSession(ctx context.Context, userId uuid.UUID, refreshToken string, tokenTTL time.Duration) error {
	const op = "redis.CreateSession"

	if err := s.client.Set(ctx, refreshToken, userId, tokenTTL).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) UpdateSession(ctx context.Context, oldRefreshToken, newRefreshToken string, tokenTTL time.Duration) error {
	const op = "redis.UpdateSession"

	if err := s.client.Rename(ctx, oldRefreshToken, newRefreshToken).Err(); err != nil {
		if err == redis.Nil {
			return fmt.Errorf("%s: %w", op, storage.ErrSessionNotFound)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.client.Expire(ctx, newRefreshToken, tokenTTL).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// returns user id
func (s *Storage) ProvideUser(ctx context.Context, refreshToken string) (uuid.UUID, error) {
	const op = "redis.ProvideUser"

	var userId uuid.NullUUID
	err := s.client.Get(ctx, refreshToken).Scan(&userId)
	if err != nil {
		if err == redis.Nil {
			return uuid.Nil, fmt.Errorf("%s: %w", op, storage.ErrSessionNotFound)
		}

		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return userId.UUID, nil
}
