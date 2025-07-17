package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/hesoyamTM/apphelper-sso/internal/config"
	"github.com/hesoyamTM/apphelper-sso/internal/grpc/auth"
	"github.com/hesoyamTM/apphelper-sso/pkg/logger"
	"go.opentelemetry.io/otel"

	"google.golang.org/grpc"
	oteltracing "google.golang.org/grpc/experimental/opentelemetry"
	"google.golang.org/grpc/stats/opentelemetry"
)

const ()

type App struct {
	log        *logger.Logger
	gRPCServer *grpc.Server
	config     config.GRPC
}

func New(ctx context.Context, authServ auth.Auth, config config.GRPC) *App {
	so := opentelemetry.ServerOption(opentelemetry.Options{
		MetricsOptions: opentelemetry.MetricsOptions{
			MeterProvider: otel.GetMeterProvider(),
			Metrics:       opentelemetry.DefaultMetrics(),
		},
		TraceOptions: oteltracing.TraceOptions{
			TracerProvider:    otel.GetTracerProvider(),
			TextMapPropagator: otel.GetTextMapPropagator(),
		},
	})

	gRPCServer := grpc.NewServer(
		so,
		grpc.UnaryInterceptor(logger.LoggingInterceptor(ctx)),
	)

	auth.RegisterServer(gRPCServer, authServ)

	return &App{
		log:        logger.GetLoggerFromCtx(ctx),
		gRPCServer: gRPCServer,
		config:     config,
	}
}

func (a *App) MustRun(ctx context.Context) {
	if err := a.run(ctx); err != nil {
		panic(err)
	}
}

func (a *App) run(ctx context.Context) error {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", a.config.Host, a.config.Port))
	if err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	a.log.Info(ctx, fmt.Sprintf("grpc server is running on %s:%d", a.config.Host, a.config.Port))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	return nil
}

func (a *App) Stop(ctx context.Context) {
	a.log.Info(ctx, "grpc server is stopping")

	a.gRPCServer.GracefulStop()
}
