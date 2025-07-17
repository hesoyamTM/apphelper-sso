package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hesoyamTM/apphelper-sso/internal/models"
	"github.com/hesoyamTM/apphelper-sso/pkg/logger"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister(t *testing.T) {
	// Mock setup
	mockUserStorage := &MockUserStorage{}
	mockSessionsStorage := &MockSessionsStorage{}
	mockCodeStorage := &MockCodeStorage{}
	mockTokenStorage := &MockTokenStorage{}
	mockRedpandaClient := &MockRedpandaClient{}

	mockUserStorage.On("CrateUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(uuid.New(), nil)
	mockSessionsStorage.On("CreateSession", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockCodeStorage.On("CreateVerificationCode", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockRedpandaClient.On("UserRegistered", mock.Anything, mock.Anything).Return(nil)

	privKey, err := genRandomPrivateKey()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test setup
	ctx, err := logger.New(context.Background(), "dev")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	userStorage := mockUserStorage
	sessionsStorage := mockSessionsStorage
	codeStorage := mockCodeStorage
	tokenStorage := mockTokenStorage
	redpandaClient := mockRedpandaClient

	authService := New(
		ctx,
		redpandaClient,
		userStorage,
		sessionsStorage,
		codeStorage,
		tokenStorage,
		time.Hour,
		time.Hour,
		time.Minute,
		time.Minute,
		privKey,
	)

	// Test
	name := "John"
	surname := "Doe"
	email := "john.doe@example.com"
	password := "password"

	tokens, err := authService.Register(ctx, name, surname, email, password)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if tokens.AccessToken == "" {
		t.Errorf("access token is empty")
	}

	if tokens.RefreshToken == "" {
		t.Errorf("refresh token is empty")
	}

	// assertions
	mockUserStorage.AssertExpectations(t)
	mockSessionsStorage.AssertExpectations(t)
	mockCodeStorage.AssertExpectations(t)
	mockRedpandaClient.AssertExpectations(t)
}

func TestLogin(t *testing.T) {
	// Mock setup
	mockUserStorage := &MockUserStorage{}
	mockSessionsStorage := &MockSessionsStorage{}
	mockCodeStorage := &MockCodeStorage{}
	mockTokenStorage := &MockTokenStorage{}
	mockRedpandaClient := &MockRedpandaClient{}

	passHash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	mockUserStorage.On("ProvideUserByEmail", mock.Anything, "john.doe@example.com").Return(models.User{
		UserInfo: models.UserInfo{
			Id:      uuid.New(),
			Name:    "John",
			Surname: "Doe",
		},
		UserAuth: models.UserAuth{
			Id:       uuid.New(),
			Email:    "john.doe@example.com",
			PassHash: passHash,
		},
	}, nil)
	mockSessionsStorage.On("CreateSession", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	privKey, err := genRandomPrivateKey()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test setup
	ctx, err := logger.New(context.Background(), "dev")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	authService := New(
		ctx,
		mockRedpandaClient,
		mockUserStorage,
		mockSessionsStorage,
		mockCodeStorage,
		mockTokenStorage,
		time.Hour,
		time.Hour,
		time.Minute,
		time.Minute,
		privKey,
	)

	// Test
	email := "john.doe@example.com"
	password := "password"

	tokens, err := authService.Login(ctx, email, password)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if tokens.AccessToken == "" {
		t.Errorf("access token is empty")
	}

	if tokens.RefreshToken == "" {
		t.Errorf("refresh token is empty")
	}

	// assertions
	mockUserStorage.AssertExpectations(t)
	mockSessionsStorage.AssertExpectations(t)
	mockCodeStorage.AssertExpectations(t)
	mockTokenStorage.AssertExpectations(t)
	mockRedpandaClient.AssertExpectations(t)
}

func TestLogout(t *testing.T) {
	// Mock setup
	mockUserStorage := &MockUserStorage{}
	mockSessionsStorage := &MockSessionsStorage{}
	mockCodeStorage := &MockCodeStorage{}
	mockTokenStorage := &MockTokenStorage{}
	mockRedpandaClient := &MockRedpandaClient{}

	refreshToken := "refresh-token"

	mockSessionsStorage.On("DeleteSession", mock.Anything, refreshToken).Return(nil)

	privKey, err := genRandomPrivateKey()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test setup
	ctx, err := logger.New(context.Background(), "dev")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	authService := New(
		ctx,
		mockRedpandaClient,
		mockUserStorage,
		mockSessionsStorage,
		mockCodeStorage,
		mockTokenStorage,
		time.Hour,
		time.Hour,
		time.Minute,
		time.Minute,
		privKey,
	)

	// Test
	if err := authService.Logout(ctx, refreshToken); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// assertions
	mockSessionsStorage.AssertExpectations(t)
}

func TestRefreshToken(t *testing.T) {
	// Mock setup
	mockUserStorage := &MockUserStorage{}
	mockSessionsStorage := &MockSessionsStorage{}
	mockCodeStorage := &MockCodeStorage{}
	mockTokenStorage := &MockTokenStorage{}
	mockRedpandaClient := &MockRedpandaClient{}

	refreshToken := "refresh-token"

	mockSessionsStorage.On("ProvideUser", mock.Anything, refreshToken).Return(uuid.New(), nil)
	mockUserStorage.On("ProvideUserById", mock.Anything, mock.Anything).Return(models.User{
		UserInfo: models.UserInfo{
			Id:      uuid.New(),
			Name:    "John",
			Surname: "Doe",
		},
		UserAuth: models.UserAuth{
			Id:       uuid.New(),
			Email:    "john.doe@example.com",
			PassHash: []byte{},
		},
	}, nil)
	mockSessionsStorage.On("UpdateSession", mock.Anything, refreshToken, mock.Anything, mock.Anything).Return(nil)

	privKey, err := genRandomPrivateKey()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test setup
	ctx, err := logger.New(context.Background(), "dev")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	authService := New(
		ctx,
		mockRedpandaClient,
		mockUserStorage,
		mockSessionsStorage,
		mockCodeStorage,
		mockTokenStorage,
		time.Hour,
		time.Hour,
		time.Minute,
		time.Minute,
		privKey,
	)

	// Test
	tokens, err := authService.RefreshToken(ctx, refreshToken)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if tokens.AccessToken == "" {
		t.Errorf("access token is empty")
	}

	if tokens.RefreshToken == "" {
		t.Errorf("refresh token is empty")
	}

	// assertions
	mockSessionsStorage.AssertExpectations(t)
	mockUserStorage.AssertExpectations(t)
}

func TestGetUser(t *testing.T) {
	// Mock setup
	mockUserStorage := &MockUserStorage{}
	mockSessionsStorage := &MockSessionsStorage{}
	mockCodeStorage := &MockCodeStorage{}
	mockTokenStorage := &MockTokenStorage{}
	mockRedpandaClient := &MockRedpandaClient{}

	userId := uuid.New()

	mockUserStorage.On("ProvideUserById", mock.Anything, userId).Return(models.User{
		UserInfo: models.UserInfo{
			Id:      userId,
			Name:    "John",
			Surname: "Doe",
		},
		UserAuth: models.UserAuth{
			Id:       uuid.New(),
			Email:    "john.doe@example.com",
			PassHash: []byte{},
		},
	}, nil)

	privKey, err := genRandomPrivateKey()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test setup
	ctx, err := logger.New(context.Background(), "dev")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	authService := New(
		ctx,
		mockRedpandaClient,
		mockUserStorage,
		mockSessionsStorage,
		mockCodeStorage,
		mockTokenStorage,
		time.Hour,
		time.Hour,
		time.Minute,
		time.Minute,
		privKey,
	)

	// Test
	user, err := authService.GetUser(ctx, userId)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if user.Name != "John" {
		t.Errorf("unexpected user name: %v", user.Name)
	}

	if user.Surname != "Doe" {
		t.Errorf("unexpected user surname: %v", user.Surname)
	}

	// assertions
	mockUserStorage.AssertExpectations(t)
}

func TestGetUsers(t *testing.T) {
	// Mock setup
	mockUserStorage := &MockUserStorage{}
	mockSessionsStorage := &MockSessionsStorage{}
	mockCodeStorage := &MockCodeStorage{}
	mockTokenStorage := &MockTokenStorage{}
	mockRedpandaClient := &MockRedpandaClient{}

	userId := uuid.New()

	mockUserStorage.On("ProvideUsersById", mock.Anything, []uuid.UUID{userId}).Return([]models.User{
		{
			UserInfo: models.UserInfo{
				Id:      userId,
				Name:    "John",
				Surname: "Doe",
			},
			UserAuth: models.UserAuth{
				Id:       uuid.New(),
				Email:    "john.doe@example.com",
				PassHash: []byte{},
			},
		},
	}, nil)

	privKey, err := genRandomPrivateKey()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test setup
	ctx, err := logger.New(context.Background(), "dev")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	authService := New(
		ctx,
		mockRedpandaClient,
		mockUserStorage,
		mockSessionsStorage,
		mockCodeStorage,
		mockTokenStorage,
		time.Hour,
		time.Hour,
		time.Minute,
		time.Minute,
		privKey,
	)

	// Test
	users, err := authService.GetUsers(ctx, []uuid.UUID{userId})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("unexpected user count: %v", len(users))
	}

	if users[0].Name != "John" {
		t.Errorf("unexpected user name: %v", users[0].Name)
	}

	if users[0].Surname != "Doe" {
		t.Errorf("unexpected user surname: %v", users[0].Surname)
	}

	// assertions
	mockUserStorage.AssertExpectations(t)
}

func TestUpdateUser(t *testing.T) {
	// Mock setup
	mockUserStorage := &MockUserStorage{}
	mockSessionsStorage := &MockSessionsStorage{}
	mockCodeStorage := &MockCodeStorage{}
	mockTokenStorage := &MockTokenStorage{}
	mockRedpandaClient := &MockRedpandaClient{}

	userId := uuid.New()

	mockUserStorage.On("UpdateUser", mock.Anything,
		models.UserInfo{
			Id:      userId,
			Name:    "John",
			Surname: "Doe",
		}).Return(nil)

	privKey, err := genRandomPrivateKey()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test setup
	ctx, err := logger.New(context.Background(), "dev")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	authService := New(
		ctx,
		mockRedpandaClient,
		mockUserStorage,
		mockSessionsStorage,
		mockCodeStorage,
		mockTokenStorage,
		time.Hour,
		time.Hour,
		time.Minute,
		time.Minute,
		privKey,
	)

	// Test
	if err := authService.UpdateUser(ctx,
		models.UserInfo{
			Id:      userId,
			Name:    "John",
			Surname: "Doe",
		}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// assertions
	mockUserStorage.AssertExpectations(t)
}

func TestDeleteUser(t *testing.T) {
	// Mock setup
	mockUserStorage := &MockUserStorage{}
	mockSessionsStorage := &MockSessionsStorage{}
	mockCodeStorage := &MockCodeStorage{}
	mockTokenStorage := &MockTokenStorage{}
	mockRedpandaClient := &MockRedpandaClient{}

	userId := uuid.New()

	mockUserStorage.On("DeleteUser", mock.Anything, userId).Return(nil)

	privKey, err := genRandomPrivateKey()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test setup
	ctx, err := logger.New(context.Background(), "dev")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	authService := New(
		ctx,
		mockRedpandaClient,
		mockUserStorage,
		mockSessionsStorage,
		mockCodeStorage,
		mockTokenStorage,
		time.Hour,
		time.Hour,
		time.Minute,
		time.Minute,
		privKey,
	)

	// Test
	if err := authService.DeleteUser(ctx, userId); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// assertions
	mockUserStorage.AssertExpectations(t)
}

func TestSendVerificationEmail(t *testing.T) {
	// Mock setup
	mockUserStorage := &MockUserStorage{}
	mockSessionsStorage := &MockSessionsStorage{}
	mockCodeStorage := &MockCodeStorage{}
	mockTokenStorage := &MockTokenStorage{}
	mockRedpandaClient := &MockRedpandaClient{}

	email := "john.doe@example.com"

	mockCodeStorage.On("CreateVerificationCode", mock.Anything, email, mock.Anything, mock.Anything).Return(nil)
	mockRedpandaClient.On("VerificationCodeUpdated", mock.Anything, mock.Anything).Return(nil)

	privKey, err := genRandomPrivateKey()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test setup
	ctx, err := logger.New(context.Background(), "dev")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	authService := New(
		ctx,
		mockRedpandaClient,
		mockUserStorage,
		mockSessionsStorage,
		mockCodeStorage,
		mockTokenStorage,
		time.Hour,
		time.Hour,
		time.Minute,
		time.Minute,
		privKey,
	)

	// Test
	if err := authService.SendVerificationEmail(ctx, email); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// assertions
	mockCodeStorage.AssertExpectations(t)
	mockRedpandaClient.AssertExpectations(t)
}

func TestSendPasswordResetEmail(t *testing.T) {
	// Mock setup
	mockUserStorage := &MockUserStorage{}
	mockSessionsStorage := &MockSessionsStorage{}
	mockCodeStorage := &MockCodeStorage{}
	mockTokenStorage := &MockTokenStorage{}
	mockRedpandaClient := &MockRedpandaClient{}

	email := "john.doe@example.com"

	mockTokenStorage.On("CreateChangePasswordToken", mock.Anything, email, mock.Anything, mock.Anything).Return(nil)
	mockRedpandaClient.On("PasswordChanged", mock.Anything, mock.Anything).Return(nil)

	privKey, err := genRandomPrivateKey()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test setup
	ctx, err := logger.New(context.Background(), "dev")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	authService := New(
		ctx,
		mockRedpandaClient,
		mockUserStorage,
		mockSessionsStorage,
		mockCodeStorage,
		mockTokenStorage,
		time.Hour,
		time.Hour,
		time.Minute,
		time.Minute,
		privKey,
	)

	// Test
	if err := authService.SendPasswordResetEmail(ctx, email); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// assertions
	mockTokenStorage.AssertExpectations(t)
	mockRedpandaClient.AssertExpectations(t)
}

func TestVerifyEmail(t *testing.T) {
	// Mock setup
	mockUserStorage := &MockUserStorage{}
	mockSessionsStorage := &MockSessionsStorage{}
	mockCodeStorage := &MockCodeStorage{}
	mockTokenStorage := &MockTokenStorage{}
	mockRedpandaClient := &MockRedpandaClient{}

	email := "john.doe@example.com"
	code := "123456"

	mockCodeStorage.On("ProvideVerificationCode", mock.Anything, email).Return(code, nil)
	mockCodeStorage.On("DeleteVerificationCode", mock.Anything, email).Return(nil)

	privKey, err := genRandomPrivateKey()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test setup
	ctx, err := logger.New(context.Background(), "dev")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	authService := New(
		ctx,
		mockRedpandaClient,
		mockUserStorage,
		mockSessionsStorage,
		mockCodeStorage,
		mockTokenStorage,
		time.Hour,
		time.Hour,
		time.Minute,
		time.Minute,
		privKey,
	)

	// Test
	if err := authService.VerifyEmail(ctx, email, code); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// assertions
	mockCodeStorage.AssertExpectations(t)
}

func TestChangePassword(t *testing.T) {
	// Mock setup
	mockUserStorage := &MockUserStorage{}
	mockSessionsStorage := &MockSessionsStorage{}
	mockCodeStorage := &MockCodeStorage{}
	mockTokenStorage := &MockTokenStorage{}
	mockRedpandaClient := &MockRedpandaClient{}

	email := "john.doe@example.com"
	newPassword := "new-password"
	token := "123456"

	mockTokenStorage.On("ProvideChangePasswordToken", mock.Anything, email).Return(token, nil)
	mockUserStorage.On("ChangePassword", mock.Anything, email, mock.Anything).Return(nil)
	mockTokenStorage.On("DeleteChangePasswordToken", mock.Anything, email).Return(nil)

	privKey, err := genRandomPrivateKey()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test setup
	ctx, err := logger.New(context.Background(), "dev")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	authService := New(
		ctx,
		mockRedpandaClient,
		mockUserStorage,
		mockSessionsStorage,
		mockCodeStorage,
		mockTokenStorage,
		time.Hour,
		time.Hour,
		time.Minute,
		time.Minute,
		privKey,
	)

	// Test
	if err := authService.ChangePassword(ctx, email, newPassword, token); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// assertions
	mockTokenStorage.AssertExpectations(t)
	mockUserStorage.AssertExpectations(t)
}
