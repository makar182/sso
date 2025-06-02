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

	storage, err := psql.New(cfg)
	if err != nil {
		log.Error("failed to init storage : %s", sl.Err(err))
		return nil
	}
	auth := authservice.NewAuthService(log, storage, storage, storage, cfg.TokenTTL)
	grpcApp := grpcApplication.NewApp(log, cfg.GRPC.Port, auth)

	return &App{
		GRPCServer: grpcApp,
	}
}
