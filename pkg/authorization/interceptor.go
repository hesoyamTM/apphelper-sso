package authorization

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"log/slog"

	"github.com/hesoyamTM/apphelper-sso/internal/lib/jwt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ServerInterceptor struct {
	log         *slog.Logger
	authMethods map[string]bool

	publicKey *ecdsa.PublicKey
}

func NewServer(log *slog.Logger, authMethods map[string]bool, pubKeyCh <-chan *ecdsa.PublicKey) *ServerInterceptor {
	interceptor := &ServerInterceptor{
		log:         log,
		authMethods: authMethods,
	}

	go func() {
		for {
			interceptor.publicKey = <-pubKeyCh
		}
	}()

	return interceptor
}

func (i *ServerInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		i.log.Debug(info.FullMethod)

		uidCtx, err := i.authorize(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}

		return handler(uidCtx, req)
	}
}

func (i *ServerInterceptor) authorize(ctx context.Context, method string) (context.Context, error) {
	if !i.authMethods[method] {
		return ctx, nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		i.log.Error("metadata is not provided")
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	bearerToken := md["authorization"]
	if len(bearerToken) == 0 {
		i.log.Error("authorization token is not provided")
		return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
	}

	uid, err := jwt.VerifyBearerToken(bearerToken[0], i.publicKey)
	if err != nil {
		if errors.Is(err, jwt.ErrUnauthorized) {
			i.log.Error("token time has expired")
			return nil, status.Errorf(codes.Unauthenticated, "token time has expired")
		}

		i.log.Error("access token is invalid", slog.String("Error", err.Error()))
		return nil, status.Error(codes.Unauthenticated, "access token is invalid")
	}

	return metadata.AppendToOutgoingContext(ctx, "uid", uid), nil
}
