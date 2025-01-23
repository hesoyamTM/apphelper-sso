package interceptor

// import (
// 	"context"
// 	"crypto/ecdsa"
// 	"log/slog"
// 	"sso/internal/lib/jwt"
// 	"strconv"

// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/metadata"
// 	"google.golang.org/grpc/status"
// )

// type ClientInterceptor struct {
// 	log         *slog.Logger
// 	authMethods map[string]bool

// 	publicKey *ecdsa.PublicKey
// }

// func NewClient(log *slog.Logger, authMethods map[string]bool, pubKeyCh <-chan *ecdsa.PublicKey) *ClientInterceptor {
// 	interceptor := &ServerInterceptor{
// 		log:         log,
// 		authMethods: authMethods,
// 	}

// 	go func() {
// 		for {
// 			interceptor.publicKey = <-pubKeyCh
// 		}
// 	}()

// 	return interceptor
// }

// func (i *ServerInterceptor) Unary() grpc.UnaryClientInterceptor {
// 	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

// 	}
// }

// func (i *ServerInterceptor) authorize(ctx context.Context, method string) (context.Context, error) {
// 	if !i.authMethods[method] {
// 		return ctx, nil
// 	}

// 	md, ok := metadata.FromIncomingContext(ctx)
// 	if !ok {
// 		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
// 	}

// 	values := md["auth"]
// 	if len(values) == 0 {
// 		return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
// 	}

// 	accessToken := values[0]
// 	uid, err := jwt.Verify(accessToken, i.publicKey)
// 	if err != nil {
// 		return nil, status.Error(codes.Unauthenticated, "access token is invalid")
// 	}

// 	return metadata.AppendToOutgoingContext(ctx, "uid", strconv.Itoa(int(uid))), nil
// }
