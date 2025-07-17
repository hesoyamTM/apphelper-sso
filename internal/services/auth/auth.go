package auth

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hesoyamTM/apphelper-sso/internal/clients/redpanda"
	"github.com/hesoyamTM/apphelper-sso/internal/lib/jwt"
	"github.com/hesoyamTM/apphelper-sso/internal/models"
	"github.com/hesoyamTM/apphelper-sso/internal/services"
	"github.com/hesoyamTM/apphelper-sso/internal/storage"
	"github.com/hesoyamTM/apphelper-sso/pkg/logger"
	"go.uber.org/zap"

	"golang.org/x/crypto/bcrypt"
)

type UserStorage interface {
	CrateUser(ctx context.Context, name, surname, email string, passHash []byte) (uuid.UUID, error)
	ProvideUserById(ctx context.Context, id uuid.UUID) (models.User, error)
	ProvideUserByEmail(ctx context.Context, email string) (models.User, error)
	ProvideUsersById(ctx context.Context, ids []uuid.UUID) ([]models.User, error)
	UpdateUser(ctx context.Context, user models.UserInfo) error
	ChangePassword(ctx context.Context, email string, newPassword []byte) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type SessionsStorage interface {
	CreateSession(ctx context.Context, userId uuid.UUID, refreshToken string, expiration time.Duration) error
	UpdateSession(ctx context.Context, oldRefreshToken, newRefreshToken string, expiration time.Duration) error
	ProvideUser(ctx context.Context, refreshToken string) (uuid.UUID, error) //returns user id
	DeleteSession(ctx context.Context, refreshToken string) error
}

type CodeStorage interface {
	CreateVerificationCode(ctx context.Context, email, code string, ttl time.Duration) error
	ProvideVerificationCode(ctx context.Context, email string) (string, error)
	DeleteVerificationCode(ctx context.Context, email string) error
}

type TokenStorage interface {
	CreateChangePasswordToken(ctx context.Context, email, token string, ttl time.Duration) error
	ProvideChangePasswordToken(ctx context.Context, email string) (string, error)
	DeleteChangePasswordToken(ctx context.Context, email string) error
}

type RedpandaClient interface {
	UserRegistered(ctx context.Context, user *redpanda.UserRegisteredEvent) error
	PasswordChanged(ctx context.Context, user *redpanda.UserRegisteredEvent) error
	VerificationCodeUpdated(ctx context.Context, user *redpanda.VerificationCodeUpdatedEvent) error
}

type Auth struct {
	log *logger.Logger

	redpandaClient RedpandaClient

	userStorage     UserStorage
	sessionsStorage SessionsStorage
	codeStorage     CodeStorage
	tokenStorage    TokenStorage

	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	codeTTL         time.Duration
	tokenTTL        time.Duration

	privateKey *ecdsa.PrivateKey
}

func New(ctx context.Context,
	redpandaClient RedpandaClient,
	uStorage UserStorage,
	sStorage SessionsStorage,
	cStorage CodeStorage,
	tStorage TokenStorage,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
	codeTTL time.Duration,
	tokenTTL time.Duration,
	privateKey *ecdsa.PrivateKey,
) *Auth {
	authService := &Auth{
		log: logger.GetLoggerFromCtx(ctx),

		redpandaClient: redpandaClient,

		userStorage:     uStorage,
		sessionsStorage: sStorage,
		codeStorage:     cStorage,
		tokenStorage:    tStorage,

		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		codeTTL:         codeTTL,
		tokenTTL:        tokenTTL,

		privateKey: privateKey,
	}

	return authService
}

func (a *Auth) Register(ctx context.Context, name, surname, email, password string) (models.JWTokens, error) {
	const op = "auth.Register"
	log := logger.GetLoggerFromCtx(ctx)

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error(ctx, "failed to generate hash from password", zap.Error(err))

		return models.JWTokens{}, fmt.Errorf("failed to generate hash: %w", err)
	}

	userId, err := a.userStorage.CrateUser(ctx, name, surname, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Error(ctx, "user already exists", zap.Error(err))
			return models.JWTokens{}, fmt.Errorf("%s: %w", op, services.ErrUserAlreadyExists)
		}

		log.Error(ctx, "failed to create user", zap.Error(err))
		return models.JWTokens{}, fmt.Errorf("%s: %w", op, err)
	}

	user := models.UserInfo{
		Id:      userId,
		Name:    name,
		Surname: surname,
	}

	tokens, err := jwt.NewTokens(user, a.accessTokenTTL, a.privateKey)
	if err != nil {
		log.Error(ctx, "failed to generate tokens", zap.Error(err))

		return models.JWTokens{}, fmt.Errorf("%s: %w", op, err)
	}

	if err = a.sessionsStorage.CreateSession(ctx, userId, tokens.RefreshToken, a.refreshTokenTTL); err != nil {
		return models.JWTokens{}, fmt.Errorf("%s: %w", op, err)
	}

	code := uuid.New().String()
	if err := a.codeStorage.CreateVerificationCode(ctx, email, code, a.codeTTL); err != nil {
		return models.JWTokens{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := a.redpandaClient.UserRegistered(ctx, &redpanda.UserRegisteredEvent{
		UserID:  userId.String(),
		Email:   email,
		Name:    name,
		Surname: surname,
		Code:    code,
	}); err != nil {
		log.Error(ctx, "failed to send user registered event", zap.Error(err))
		return models.JWTokens{}, fmt.Errorf("%s: %w", op, err)
	}

	return tokens, nil
}

func (a *Auth) Login(ctx context.Context, email, password string) (models.JWTokens, error) {
	const op = "auth.Login"
	log := logger.GetLoggerFromCtx(ctx) //a.log.With(slog.String("login", login), slog.String("op", op))
	log.Info(ctx, "authorize user")

	user, err := a.userStorage.ProvideUserByEmail(ctx, email)
	if err != nil {
		log.Error(ctx, "failed to provide user", zap.Error(err))

		if errors.Is(err, storage.ErrUserNotFound) {
			return models.JWTokens{}, fmt.Errorf("%s: %w", op, services.ErrUserNotFound)
		}

		return models.JWTokens{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Error(ctx, "incorrect password", zap.Error(err))

		return models.JWTokens{}, fmt.Errorf("%s: %w", op, services.ErrInvalidCredentials)
	}

	tokens, err := jwt.NewTokens(user.UserInfo, a.accessTokenTTL, a.privateKey)
	if err != nil {
		log.Error(ctx, "failed to generate tokens", zap.Error(err))

		return models.JWTokens{}, fmt.Errorf("failed to generate tokens: %w", err)
	}

	if err = a.sessionsStorage.CreateSession(ctx, user.UserAuth.Id, tokens.RefreshToken, a.refreshTokenTTL); err != nil {
		log.Error(ctx, "failed to create session", zap.Error(err))

		return models.JWTokens{}, fmt.Errorf("%s: %w", op, err)
	}

	return tokens, nil
}

func (a *Auth) Logout(ctx context.Context, refreshToken string) error {
	const op = "auth.Logout"
	log := logger.GetLoggerFromCtx(ctx)

	if err := a.sessionsStorage.DeleteSession(ctx, refreshToken); err != nil {
		log.Error(ctx, "failed to delete session", zap.Error(err))

		if errors.Is(err, storage.ErrSessionNotFound) {
			return fmt.Errorf("%s: %w", op, services.ErrNotAuthorized)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *Auth) RefreshToken(ctx context.Context, refreshToken string) (models.JWTokens, error) {
	const op = "auth.RefreshToken"
	log := logger.GetLoggerFromCtx(ctx) //a.log.With(slog.String("op", op))

	userId, err := a.sessionsStorage.ProvideUser(ctx, refreshToken)
	if err != nil {
		log.Error(ctx, "failed to provide user", zap.Error(err))

		if errors.Is(err, storage.ErrSessionNotFound) {
			return models.JWTokens{}, fmt.Errorf("%s: %w", op, services.ErrNotAuthorized)
		}

		return models.JWTokens{}, fmt.Errorf("%s: %w", op, err)
	}

	user, err := a.userStorage.ProvideUserById(ctx, userId)
	if err != nil {
		log.Error(ctx, "failed to provide user", zap.Error(err))

		if errors.Is(err, storage.ErrUserNotFound) {
			return models.JWTokens{}, fmt.Errorf("%s: %w", op, services.ErrUserNotFound)
		}

		return models.JWTokens{}, fmt.Errorf("%s: %w", op, err)
	}

	newTokens, err := jwt.NewTokens(user.UserInfo, a.accessTokenTTL, a.privateKey)
	if err != nil {
		log.Error(ctx, "failed to generate tokens", zap.Error(err))

		return models.JWTokens{}, fmt.Errorf("%s: %w", op, err)
	}

	if err = a.sessionsStorage.UpdateSession(ctx, refreshToken, newTokens.RefreshToken, a.refreshTokenTTL); err != nil {
		log.Error(ctx, "failed to update session", zap.Error(err))

		if errors.Is(err, storage.ErrSessionNotFound) {
			return models.JWTokens{}, fmt.Errorf("%s: %w", op, services.ErrNotAuthorized)
		}
		return models.JWTokens{}, fmt.Errorf("%s: %w", op, err)
	}

	return newTokens, nil
}

func (a *Auth) GetUser(ctx context.Context, id uuid.UUID) (models.User, error) {
	const op = "auth.GetUser"
	log := logger.GetLoggerFromCtx(ctx) //a.log.With(slog.String("op", op))
	_ = log

	user, err := a.userStorage.ProvideUserById(ctx, id)
	if err != nil {
		log.Error(ctx, "failed to provide user", zap.Error(err))

		if errors.Is(err, storage.ErrUserNotFound) {
			return models.User{}, fmt.Errorf("%s: %w", op, services.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (a *Auth) GetUsers(ctx context.Context, ids []uuid.UUID) ([]models.User, error) {
	const op = "auth.GetUsers"
	log := logger.GetLoggerFromCtx(ctx) //a.log.With(slog.String("op", op))

	users, err := a.userStorage.ProvideUsersById(ctx, ids)
	if err != nil {
		log.Error(ctx, "failed to provide user", zap.Error(err))

		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, fmt.Errorf("%s: %w", op, services.ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return users, nil
}

func (a *Auth) UpdateUser(ctx context.Context, user models.UserInfo) error {
	const op = "auth.UpdateUser"
	log := logger.GetLoggerFromCtx(ctx)

	if err := a.userStorage.UpdateUser(ctx, user); err != nil {
		log.Error(ctx, "failed to update user", zap.Error(err))

		if errors.Is(err, storage.ErrUserNotFound) {
			return fmt.Errorf("%s: %w", op, services.ErrUserNotFound)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *Auth) DeleteUser(ctx context.Context, id uuid.UUID) error {
	const op = "auth.DeleteUser"
	log := logger.GetLoggerFromCtx(ctx)

	if err := a.userStorage.DeleteUser(ctx, id); err != nil {
		log.Error(ctx, "failed to delete user", zap.Error(err))

		if errors.Is(err, storage.ErrUserNotFound) {
			return fmt.Errorf("%s: %w", op, services.ErrUserNotFound)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Auth) SendVerificationEmail(ctx context.Context, email string) error {
	const op = "auth.SendVerificationEmail"

	code := uuid.New().String()

	if err := s.codeStorage.CreateVerificationCode(ctx, email, code, s.codeTTL); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.redpandaClient.VerificationCodeUpdated(ctx, &redpanda.VerificationCodeUpdatedEvent{
		Email: email,
		Code:  code,
	}); err != nil {
		s.log.Error(ctx, "failed to send verification email", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Auth) VerifyEmail(ctx context.Context, email, code string) error {
	const op = "auth.VerifyEmail"

	providedCode, err := s.codeStorage.ProvideVerificationCode(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrVerificationCodeNotFound) {
			return fmt.Errorf("%s: %w", op, services.ErrNotAuthorized)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	if providedCode != code {
		return fmt.Errorf("%s: %w", op, services.ErrInvalidCredentials)
	}

	if err := s.codeStorage.DeleteVerificationCode(ctx, email); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Auth) SendPasswordResetEmail(ctx context.Context, email string) error {
	const op = "auth.SendPasswordResetEmail"

	code := uuid.New().String()

	if err := s.tokenStorage.CreateChangePasswordToken(ctx, email, code, s.tokenTTL); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.redpandaClient.PasswordChanged(ctx, &redpanda.UserRegisteredEvent{
		UserID:  email,
		Email:   email,
		Name:    "",
		Surname: "",
		Code:    code,
	}); err != nil {
		s.log.Error(ctx, "failed to send password reset email", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Auth) ChangePassword(ctx context.Context, email, newPassword, token string) error {
	const op = "auth.ChangePassword"

	tok, err := s.tokenStorage.ProvideChangePasswordToken(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrChangePasswordTokenNotFound) {
			return fmt.Errorf("%s: %w", op, services.ErrNotAuthorized)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	if tok != token {
		return fmt.Errorf("%s: %w", op, services.ErrInvalidCredentials)
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.userStorage.ChangePassword(ctx, email, passHash); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.tokenStorage.DeleteChangePasswordToken(ctx, email); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
