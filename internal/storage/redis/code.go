package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/hesoyamTM/apphelper-sso/internal/storage"
	"github.com/redis/go-redis/v9"
)

func (s *Storage) CreateVerificationCode(ctx context.Context, email, code string, ttl time.Duration) error {
	const op = "redis.CreateVerificationCode"

	if err := s.client.Set(ctx, email, code, ttl).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) ProvideVerificationCode(ctx context.Context, email string) (string, error) {
	const op = "redis.ProvideVerificationCode"

	var code string
	err := s.client.Get(ctx, email).Scan(&code)
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("%s: %w", op, storage.ErrVerificationCodeNotFound)
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return code, nil
}

func (s *Storage) DeleteVerificationCode(ctx context.Context, email string) error {
	const op = "redis.DeleteVerificationCode"

	if err := s.client.Del(ctx, email).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) CreateChangePasswordToken(ctx context.Context, email, token string, ttl time.Duration) error {
	const op = "redis.CreateChangePasswordToken"

	if err := s.client.Set(ctx, email, token, ttl).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) ProvideChangePasswordToken(ctx context.Context, email string) (string, error) {
	const op = "redis.ProvideChangePasswordToken"

	var token string
	err := s.client.Get(ctx, email).Scan(&token)
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("%s: %w", op, storage.ErrChangePasswordTokenNotFound)
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func (s *Storage) DeleteChangePasswordToken(ctx context.Context, email string) error {
	const op = "redis.DeleteChangePasswordToken"

	if err := s.client.Del(ctx, email).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
