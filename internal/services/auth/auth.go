package auth

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log/slog"
	"sso/internal/models"
	"sso/internal/services"
	"sso/internal/storage"
	"sso/pkg/lib/jwt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserStorage interface {
	CrateUser(ctx context.Context, name, surname, login string, passHash []byte) (int64, error)
	ProvideUserById(ctx context.Context, id int64) (models.User, error)
	ProvideUserByLogin(ctx context.Context, login string) (models.User, error)
	ProvideUsersById(ctx context.Context, ids []int64) ([]models.User, error)
}

type SessionsStorage interface {
	CreateSession(ctx context.Context, userId int64, refreshToken string, expiration time.Duration) error
	UpdateSession(ctx context.Context, oldRefreshToken, newRefreshToken string, expiration time.Duration) error
	// returns user id
	ProvideUser(ctx context.Context, refreshToken string) (int64, error)
}

type Auth struct {
	log *slog.Logger

	userStorage     UserStorage
	sessionsStorage SessionsStorage

	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration

	privateKey *ecdsa.PrivateKey
}

func New(log *slog.Logger,
	uStorage UserStorage,
	sStorage SessionsStorage,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
	prKeyCh <-chan *ecdsa.PrivateKey) *Auth {
	authService := &Auth{
		log:             log,
		userStorage:     uStorage,
		sessionsStorage: sStorage,

		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}

	go func() {
		for prKey := range prKeyCh {
			authService.privateKey = prKey
		}
	}()

	return authService
}

func (a *Auth) Register(ctx context.Context, name, surname, login, password string) (models.JWTokens, error) {
	const op = "auth.Register"
	log := a.log.With(slog.String("op", op)).With(slog.String("login", login))

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate hash from password", slog.String("Error", err.Error()))

		return models.JWTokens{}, fmt.Errorf("failed to generate hash: %w", err)
	}

	userId, err := a.userStorage.CrateUser(ctx, name, surname, login, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Error("user already exists")
			return models.JWTokens{}, fmt.Errorf("%s: %w", op, services.ErrUserAlreadyExists)
		}

		log.Error("failed to create user", slog.String("Error", err.Error()))
		return models.JWTokens{}, fmt.Errorf("%s: %w", op, err)
	}

	user := models.UserInfo{
		Id:      userId,
		Name:    name,
		Surname: surname,
	}

	tokens, err := jwt.NewTokens(user, a.accessTokenTTL, a.privateKey)
	if err != nil {
		log.Error("failed to generate tokens", slog.String("Error", err.Error()))

		return models.JWTokens{}, fmt.Errorf("%s: %w", op, err)
	}

	if err = a.sessionsStorage.CreateSession(ctx, userId, tokens.RefreshToken, a.refreshTokenTTL); err != nil {
		return models.JWTokens{}, fmt.Errorf("%s: %w", op, err)
	}

	return tokens, nil
}

func (a *Auth) Login(ctx context.Context, login, password string) (models.JWTokens, error) {
	const op = "auth.Login"
	log := a.log.With(slog.String("login", login), slog.String("op", op))
	log.Info("registering user")

	user, err := a.userStorage.ProvideUserByLogin(ctx, login)
	if err != nil {
		log.Error("failed to provide user", slog.String("Error", err.Error()))

		if errors.Is(err, storage.ErrUserNotFound) {
			return models.JWTokens{}, fmt.Errorf("%s: %w", op, services.ErrUserNotFound)
		}

		return models.JWTokens{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Error("incorrect password", slog.String("Error", err.Error()))

		return models.JWTokens{}, fmt.Errorf("%s: %w", op, services.ErrInvalidCredentials)
	}

	tokens, err := jwt.NewTokens(user.UserInfo, a.accessTokenTTL, a.privateKey)
	if err != nil {
		log.Error("failed to generate tokens", slog.String("Error", err.Error()))

		return models.JWTokens{}, fmt.Errorf("failed to generate tokens: %w", err)
	}

	if err = a.sessionsStorage.CreateSession(ctx, user.UserAuth.Id, tokens.RefreshToken, a.refreshTokenTTL); err != nil {
		log.Error("failed to create session", slog.String("Error", err.Error()))

		return models.JWTokens{}, fmt.Errorf("%s: %w", op, err)
	}

	return tokens, nil
}

func (a *Auth) RefreshToken(ctx context.Context, refreshToken string) (models.JWTokens, error) {
	const op = "auth.RefreshToken"
	log := a.log.With(slog.String("op", op))

	userId, err := a.sessionsStorage.ProvideUser(ctx, refreshToken)
	if err != nil {
		log.Error("failed to provide user", slog.String("Error", err.Error()))

		if errors.Is(err, storage.ErrSessionNotFound) {
			return models.JWTokens{}, fmt.Errorf("%s: %w", op, services.ErrNotAuthorized)
		}

		return models.JWTokens{}, fmt.Errorf("%s: %w", op, err)
	}

	user, err := a.userStorage.ProvideUserById(ctx, userId)
	if err != nil {
		log.Error("failed to provide user", slog.String("Error", err.Error()))

		if errors.Is(err, storage.ErrUserNotFound) {
			return models.JWTokens{}, fmt.Errorf("%s: %w", op, services.ErrUserNotFound)
		}

		return models.JWTokens{}, fmt.Errorf("%s: %w", op, err)
	}

	newTokens, err := jwt.NewTokens(user.UserInfo, a.accessTokenTTL, a.privateKey)
	if err != nil {
		log.Error("failed to generate tokens", slog.String("Error", err.Error()))

		return models.JWTokens{}, fmt.Errorf("%s: %w", op, err)
	}

	if err = a.sessionsStorage.UpdateSession(ctx, refreshToken, newTokens.RefreshToken, a.refreshTokenTTL); err != nil {
		log.Error("failed to update session", slog.String("Error", err.Error()))

		if errors.Is(err, storage.ErrSessionNotFound) {
			return models.JWTokens{}, fmt.Errorf("%s: %w", op, services.ErrNotAuthorized)
		}
		return models.JWTokens{}, fmt.Errorf("%s: %w", op, err)
	}

	return newTokens, nil
}

func (a *Auth) GetUser(ctx context.Context, id int64) (models.User, error) {
	const op = "auth.GetUser"
	log := a.log.With(slog.String("op", op))
	_ = log

	user, err := a.userStorage.ProvideUserById(ctx, id)
	if err != nil {
		log.Error("failed to provide user", slog.String("Error", err.Error()))

		if errors.Is(err, storage.ErrUserNotFound) {
			return models.User{}, fmt.Errorf("%s: %w", op, services.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (a *Auth) GetUsers(ctx context.Context, ids []int64) ([]models.User, error) {
	const op = "auth.GetUsers"
	log := a.log.With(slog.String("op", op))
	_ = log

	users, err := a.userStorage.ProvideUsersById(ctx, ids)
	if err != nil {
		log.Error("failed to provide user", slog.String("Error", err.Error()))

		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, fmt.Errorf("%s: %w", op, services.ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return users, nil
}
