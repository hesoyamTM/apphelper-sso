package logger

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Middleware func(next http.Handler) http.Handler

func LoggingMiddleware(loggerCtx context.Context) Middleware {
	logger := GetLoggerFromCtx(loggerCtx)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), Key, logger)
			r = r.WithContext(ctx)

			logger.Info(ctx, "request", zap.String("method", r.Method), zap.Time("request time", time.Now()))

			next.ServeHTTP(w, r)
		})
	}
}
