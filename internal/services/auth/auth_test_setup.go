package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"time"

	"github.com/google/uuid"
	"github.com/hesoyamTM/apphelper-sso/internal/clients/redpanda"
	"github.com/hesoyamTM/apphelper-sso/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockUserStorage struct {
	mock.Mock
}

func (m *MockUserStorage) CrateUser(ctx context.Context, name, surname, email string, passHash []byte) (uuid.UUID, error) {
	args := m.Called(ctx, name, surname, email, passHash)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockUserStorage) ProvideUserById(ctx context.Context, id uuid.UUID) (models.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockUserStorage) ProvideUserByEmail(ctx context.Context, email string) (models.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockUserStorage) ProvideUsersById(ctx context.Context, ids []uuid.UUID) ([]models.User, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserStorage) UpdateUser(ctx context.Context, user models.UserInfo) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserStorage) ChangePassword(ctx context.Context, email string, newPassword []byte) error {
	args := m.Called(ctx, email, newPassword)
	return args.Error(0)
}

func (m *MockUserStorage) DeleteUser(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockSessionsStorage struct {
	mock.Mock
}

func (m *MockSessionsStorage) CreateSession(ctx context.Context, userId uuid.UUID, refreshToken string, expiration time.Duration) error {
	args := m.Called(ctx, userId, refreshToken, expiration)
	return args.Error(0)
}

func (m *MockSessionsStorage) UpdateSession(ctx context.Context, oldRefreshToken, newRefreshToken string, expiration time.Duration) error {
	args := m.Called(ctx, oldRefreshToken, newRefreshToken, expiration)
	return args.Error(0)
}

func (m *MockSessionsStorage) ProvideUser(ctx context.Context, refreshToken string) (uuid.UUID, error) {
	args := m.Called(ctx, refreshToken)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockSessionsStorage) DeleteSession(ctx context.Context, refreshToken string) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

type MockCodeStorage struct {
	mock.Mock
}

func (m *MockCodeStorage) CreateVerificationCode(ctx context.Context, email, code string, ttl time.Duration) error {
	args := m.Called(ctx, email, code, ttl)
	return args.Error(0)
}

func (m *MockCodeStorage) ProvideVerificationCode(ctx context.Context, email string) (string, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockCodeStorage) DeleteVerificationCode(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

type MockTokenStorage struct {
	mock.Mock
}

func (m *MockTokenStorage) CreateChangePasswordToken(ctx context.Context, email, token string, ttl time.Duration) error {
	args := m.Called(ctx, email, token, ttl)
	return args.Error(0)
}

func (m *MockTokenStorage) ProvideChangePasswordToken(ctx context.Context, email string) (string, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockTokenStorage) DeleteChangePasswordToken(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

type MockRedpandaClient struct {
	mock.Mock
}

func (m *MockRedpandaClient) UserRegistered(ctx context.Context, user *redpanda.UserRegisteredEvent) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRedpandaClient) PasswordChanged(ctx context.Context, user *redpanda.UserRegisteredEvent) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRedpandaClient) VerificationCodeUpdated(ctx context.Context, user *redpanda.VerificationCodeUpdatedEvent) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func genRandomPrivateKey() (*ecdsa.PrivateKey, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	return key, nil
}
