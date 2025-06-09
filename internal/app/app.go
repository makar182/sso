package app

import (
	"log/slog"
	grpcApplication "sso/internal/app/grpc"
	"sso/internal/config"
	"sso/internal/lib/logger/sl"
	authservice "sso/internal/services/auth"
	psql "sso/internal/storage/postgreSQL"
)

type App struct {
	GRPCServer *grpcApplication.App
}

func NewApp(
	log *slog.Logger,
	cfg *config.Config) *App {
	const op = "app.Application.Run"
	log = log.With(
		slog.String("operation", op),
	)

	storage, err := psql.New(cfg)
	if err != nil {
		log.Error("failed to init storage : %s", sl.Err(err))
		return nil
	}
	log.Info("storage initialized", slog.Any("db", cfg))
	auth := authservice.NewAuthService(log, storage, storage, storage, storage, cfg.TokenTTL)
	log.Info("auth service initialized")
	grpcApp := grpcApplication.NewApp(log, cfg.GRPC.Port, auth)
	log.Info("gRPC server initialized", slog.Int("port", cfg.GRPC.Port))

	return &App{
		GRPCServer: grpcApp,
	}
}
