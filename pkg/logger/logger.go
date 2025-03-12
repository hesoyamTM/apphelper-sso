package logger

import (
	"context"

	"go.uber.org/zap"
)

type ctxKey string

const (
	Key ctxKey = "logger"

	RequestID ctxKey = "request_id"

	localEnv = "local"
	devEnv   = "dev"
	prodEnv  = "prod"
)

type Logger struct {
	l *zap.Logger
}

func New(ctx context.Context, env string) (context.Context, error) {
	var logger *zap.Logger
	var err error

	switch env {
	case localEnv:
		logger, err = zap.NewDevelopment()
	case devEnv:
		logger, err = zap.NewDevelopment()
	case prodEnv:
		logger, err = zap.NewProduction()
	default:
		logger, err = zap.NewProduction()
	}

	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, Key, &Logger{logger})
	return ctx, nil
}

func GetLoggerFromCtx(ctx context.Context) *Logger {
	return ctx.Value(Key).(*Logger)
}

func (l *Logger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(string(RequestID), ctx.Value(RequestID).(string)))
	}
	l.l.Debug(msg, fields...)
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(string(RequestID), ctx.Value(RequestID).(string)))
	}
	l.l.Info(msg, fields...)
}

func (l *Logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(string(RequestID), ctx.Value(RequestID).(string)))
	}
	l.l.Error(msg, fields...)
}
