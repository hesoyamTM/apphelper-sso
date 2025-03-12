package logger

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func LoggingInterceptor(loggerCtx context.Context) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		guid := uuid.New().String()
		ctx = context.WithValue(ctx, RequestID, guid)
		logger := GetLoggerFromCtx(loggerCtx)

		ctx = context.WithValue(ctx, Key, logger)

		logger.Info(ctx, "request", zap.String("method", info.FullMethod), zap.Time("request time", time.Now()))

		return handler(ctx, req)
	}
}
