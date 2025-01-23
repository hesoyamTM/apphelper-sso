package interceptor

import (
	"context"
	"crypto/ecdsa"
	"log/slog"
	"sso/pkg/lib/jwt"
	"strconv"

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

		uidCtx, err := i.authorize(ctx, info.FullMethod)
		if err != nil {
			return nil, status.Error(codes.Internal, "internal error")
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
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md["auth"]
	if len(values) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
	}

	accessToken := values[0]
	uid, err := jwt.Verify(accessToken, i.publicKey)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "access token is invalid")
	}

	return metadata.AppendToOutgoingContext(ctx, "uid", strconv.Itoa(int(uid))), nil
}
