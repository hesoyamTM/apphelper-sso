package redis

import (
	"context"
	"fmt"
	"sso/internal/storage"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Storage struct {
	client *redis.Client
}

func New(host, pass string, port int) *Storage {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: pass,
		DB:       0,
	})

	return &Storage{
		client: client,
	}
}

func (s *Storage) CreateSession(ctx context.Context, userId int64, refreshToken string, tokenTTL time.Duration) error {
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
func (s *Storage) ProvideUser(ctx context.Context, refreshToken string) (int64, error) {
	const op = "redis.ProvideUser"

	res, err := s.client.Get(ctx, refreshToken).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrSessionNotFound)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	userId, err := strconv.ParseInt(res, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return userId, nil
}
