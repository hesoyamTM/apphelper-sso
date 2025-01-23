package auth

import (
	"context"
	"errors"
	"sso/internal/models"
	"sso/internal/services"

	ssov1 "github.com/hesoyamTM/apphelper-protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Register(ctx context.Context, name, surname, login, password string) (models.JWTokens, error)
	Login(ctx context.Context, login, password string) (models.JWTokens, error)
	RefreshToken(ctx context.Context, refreshToken string) (models.JWTokens, error)
	GetUser(ctx context.Context, id int64) (models.User, error)
	GetUsers(ctx context.Context, ids []int64) ([]models.User, error)
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
	login := req.Login
	pass := req.Password
	name := req.Name
	surname := req.Surname

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
	id := req.GetUserId()

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

	if err := validateGetUsers(ids); err != nil {
		return nil, status.Error(codes.Internal, "validation error")
	}

	users, err := s.authService.GetUsers(ctx, ids)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, "user not found")
		}

		return nil, status.Error(codes.Internal, "Internal error")
	}

	usersResp := make([]*ssov1.User, len(users))

	for i := range users {
		usersResp[i] = &ssov1.User{
			Id:      users[i].UserInfo.Id,
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
		return nil, status.Error(codes.Internal, "validation error")
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
