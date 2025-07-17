package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/hesoyamTM/apphelper-sso/internal/models"
	"github.com/hesoyamTM/apphelper-sso/internal/services"

	ssov1 "github.com/hesoyamTM/apphelper-protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Register(ctx context.Context, name, surname, login, password string) (models.JWTokens, error)
	Login(ctx context.Context, login, password string) (models.JWTokens, error)
	Logout(ctx context.Context, refreshToken string) error
	RefreshToken(ctx context.Context, refreshToken string) (models.JWTokens, error)
	GetUser(ctx context.Context, id uuid.UUID) (models.User, error)
	GetUsers(ctx context.Context, ids []uuid.UUID) ([]models.User, error)
	UpdateUser(ctx context.Context, user models.UserInfo) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	ChangePassword(ctx context.Context, email, newPassword, token string) error
	SendVerificationEmail(ctx context.Context, email string) error
	SendPasswordResetEmail(ctx context.Context, email string) error
	VerifyEmail(ctx context.Context, email, code string) error
}

type serverAPI struct {
	authService Auth
	ssov1.UnimplementedAuthServer
}

func RegisterServer(gRpc *grpc.Server, authService Auth) {
	ssov1.RegisterAuthServer(gRpc, &serverAPI{authService: authService})
}

func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	login := req.GetLogin()
	pass := req.GetPassword()

	if err := validateLogin(ctx, login, pass); err != nil {
		return nil, status.Error(codes.InvalidArgument, "validation error")
	}

	tokens, err := s.authService.Login(ctx, login, pass)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		if errors.Is(err, services.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, "user not found")
		}

		return nil, status.Error(codes.InvalidArgument, "internal error")
	}

	return &ssov1.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	login := req.GetLogin()
	pass := req.GetPassword()
	name := req.GetName()
	surname := req.GetSurname()

	if err := validateRegister(ctx, name, surname, login, pass); err != nil {
		return nil, status.Error(codes.InvalidArgument, "validation error")
	}

	tokens, err := s.authService.Register(ctx, name, surname, login, pass)
	if err != nil {
		if errors.Is(err, services.ErrUserAlreadyExists) {
			return nil, status.Error(codes.InvalidArgument, "user already exists")
		}

		return nil, status.Error(codes.InvalidArgument, "internal error")
	}

	return &ssov1.RegisterResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *serverAPI) GetUser(ctx context.Context, req *ssov1.GetUserRequest) (*ssov1.GetUserResponse, error) {
	id, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	user, err := s.authService.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, "user not found")
		}

		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &ssov1.GetUserResponse{
		Name:    user.Name,
		Surname: user.Surname,
	}, nil
}

func (s *serverAPI) GetUsers(ctx context.Context, req *ssov1.GetUsersRequest) (*ssov1.GetUsersResponse, error) {
	ids := req.GetUserIds()
	idsUUID := make([]uuid.UUID, len(ids))
	var err error
	for i := range ids {
		idsUUID[i], err = uuid.Parse(ids[i])
		if err != nil {
			continue
		}
	}

	if err := validateGetUsers(idsUUID); err != nil {
		return nil, status.Error(codes.InvalidArgument, "validation error")
	}

	users, err := s.authService.GetUsers(ctx, idsUUID)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, "user not found")
		}

		return nil, status.Error(codes.Internal, "Internal error")
	}

	usersResp := make([]*ssov1.User, len(users))

	for i := range users {
		usersResp[i] = &ssov1.User{
			Id:      users[i].UserInfo.Id.String(),
			Name:    users[i].Name,
			Surname: users[i].Surname,
		}
	}

	return &ssov1.GetUsersResponse{
		Users: usersResp,
	}, nil
}

func (s *serverAPI) RefreshToken(ctx context.Context, req *ssov1.RefreshTokenRequest) (*ssov1.RefreshTokenResponse, error) {
	token := req.GetRefreshToken()

	if err := validateRefreshToken(ctx, token); err != nil {
		return nil, status.Error(codes.InvalidArgument, "validation error")
	}

	tokens, err := s.authService.RefreshToken(ctx, token)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, "user not found")
		}
		if errors.Is(err, services.ErrNotAuthorized) {
			return nil, status.Error(codes.Unauthenticated, "not authorized")
		}

		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &ssov1.RefreshTokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *serverAPI) UpdateUser(ctx context.Context, req *ssov1.UpdateUserRequest) (*ssov1.UpdateUserResponse, error) {
	id, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	// TODO: check user authentication

	user := models.UserInfo{
		Id:      id,
		Name:    req.GetName(),
		Surname: req.GetSurname(),
	}

	if err := validateUpdateUser(ctx, user); err != nil {
		return nil, status.Error(codes.InvalidArgument, "validation error")
	}

	if err := s.authService.UpdateUser(ctx, user); err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, "user not found")
		}

		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &ssov1.UpdateUserResponse{}, nil
}

func (s *serverAPI) DeleteUser(ctx context.Context, req *ssov1.DeleteUserRequest) (*ssov1.DeleteUserResponse, error) {
	id, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	// TODO: check user authentication

	if err := s.authService.DeleteUser(ctx, id); err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, "user not found")
		}

		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &ssov1.DeleteUserResponse{}, nil
}

func (s *serverAPI) Logout(ctx context.Context, req *ssov1.LogoutRequest) (*ssov1.LogoutResponse, error) {
	token := req.GetRefreshToken()

	if err := validateLogout(ctx, token); err != nil {
		return nil, status.Error(codes.InvalidArgument, "validation error")
	}

	if err := s.authService.Logout(ctx, token); err != nil {
		if errors.Is(err, services.ErrNotAuthorized) {
			return nil, status.Error(codes.Unauthenticated, "not authorized")
		}

		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &ssov1.LogoutResponse{}, nil
}

func (s *serverAPI) ChangePassword(ctx context.Context, req *ssov1.ChangePasswordRequest) (*ssov1.ChangePasswordResponse, error) {
	email := req.GetEmail()
	newPassword := req.GetNewPassword()
	token := req.GetToken()

	if err := validateChangePassword(ctx, newPassword, token); err != nil {
		return nil, status.Error(codes.InvalidArgument, "validation error")
	}

	if err := s.authService.ChangePassword(ctx, email, newPassword, token); err != nil {
		if errors.Is(err, services.ErrNotAuthorized) {
			return nil, status.Error(codes.Unauthenticated, "not authorized")
		}
		if errors.Is(err, services.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		if errors.Is(err, services.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, "user not found")
		}

		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &ssov1.ChangePasswordResponse{}, nil
}

func (s *serverAPI) SendVerificationEmail(ctx context.Context, req *ssov1.SendVerificationEmailRequest) (*ssov1.SendVerificationEmailResponse, error) {
	if err := s.authService.SendVerificationEmail(ctx, req.GetEmail()); err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, "user not found")
		}

		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &ssov1.SendVerificationEmailResponse{}, nil
}

func (s *serverAPI) VerifyEmail(ctx context.Context, req *ssov1.VerifyEmailRequest) (*ssov1.VerifyEmailResponse, error) {
	if err := s.authService.VerifyEmail(ctx, req.GetEmail(), req.GetCode()); err != nil {
		if errors.Is(err, services.ErrNotAuthorized) {
			return nil, status.Error(codes.Unauthenticated, "not authorized")
		}

		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &ssov1.VerifyEmailResponse{}, nil
}
