package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/hesoyamTM/apphelper-sso/internal/grpc/auth"
	"github.com/hesoyamTM/apphelper-sso/pkg/logger"

	"google.golang.org/grpc"
)

const ()

type App struct {
	log        *logger.Logger
	gRPCServer *grpc.Server
	config     Config
}

type Config struct {
	Host string
	Port int
}

func New(ctx context.Context, authServ auth.Auth, config Config) *App {
	gRPCServer := grpc.NewServer(grpc.UnaryInterceptor(logger.LoggingInterceptor(ctx)))

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

	a.log.Info(ctx, "server is running")

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	return nil
}

func (a *App) Stop(ctx context.Context) {
	a.log.Info(ctx, "grpc server is stopping")

	a.gRPCServer.GracefulStop()
}
