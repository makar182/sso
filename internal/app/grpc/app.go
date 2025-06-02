package grpcApplication

import (
	"fmt"
	"google.golang.org/grpc"
	"log/slog"
	"net"
	authgrpc "sso/internal/grpc/auth"
	"sso/internal/lib/logger/sl"
	authservice "sso/internal/services/auth"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func NewApp(log *slog.Logger, port int, auth *authservice.Auth) *App {
	gRPCServer := grpc.NewServer()

	authgrpc.RegisterServerAPI(gRPCServer, auth)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic("failed to run gRPC server")
	}
}

func (a *App) Run() error {
	const op = "app.gRPC.Application.Run"
	log := a.log.With(
		slog.String("operation", op),
		slog.Int("port", a.port),
	)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		log.Error("failed to listen", sl.Err(err))
	}

	log.Info("gRPC server is running", slog.String("address", lis.Addr().String()))

	if err := a.gRPCServer.Serve(lis); err != nil {
		log.Error("failed to serve gRPC server", sl.Err(err))
		return err
	}

	return nil
}

func (a *App) Stop() {
	const op = "app.gRPC.Application.Stop"
	log := a.log.With(
		slog.String("operation", op),
	)

	log.Info("stopping gRPC server")
	a.gRPCServer.GracefulStop()
	log.Info("gRPC server stopped")
}
