package grpc

import (
	"fmt"
	"log/slog"
	"net"
	"sso/internal/grpc/auth"

	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	host       string
	port       int
}

func New(log *slog.Logger, authServ auth.Auth, host string, port int) *App {
	gRPCServer := grpc.NewServer()

	auth.RegisterServer(gRPCServer, authServ)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		host:       host,
		port:       port,
	}
}

func (a *App) MustRun() {
	if err := a.run(); err != nil {
		panic(err)
	}
}

func (a *App) run() error {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", a.host, a.port))
	if err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	a.log.Info("server is running")

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	return nil
}

func (a *App) Stop() {
	a.log.Info("grpc server is stopping")

	a.gRPCServer.GracefulStop()
}
