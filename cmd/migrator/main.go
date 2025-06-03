package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log/slog"
	"os"
	"sso/internal/config"
	"sso/internal/lib/logger/handlers/slogpretty"
	"sso/internal/lib/logger/sl"
)

type MigrationsMode string

const (
	None      MigrationsMode = "none"
	DownOnly  MigrationsMode = "down-only"
	UpOnly    MigrationsMode = "up-only"
	UpAndDown MigrationsMode = "up-and-down"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	//Переводим флаги в переменные окружения
	MustSetupEnvVars()

	//TODO: инициализировать объект конфига
	cfg := config.MustLoad()

	//TODO: инициализировать логгер
	log := setupLogger(cfg.Env)
	log = log.With(slog.String("env", cfg.Env))

	const op = "Storage.PostgreSQL.Migrations"

	storagePath := cfg.StoragePath
	migrationsTable := os.Getenv("MIGRATIONS_TABLE")
	if migrationsTable != "migrations" {
		storagePath = fmt.Sprintf("%s&x-migrations-table=%s", storagePath, migrationsTable)
	}

	mode := MigrationsMode(os.Getenv("MIGRATIONS_MODE"))
	if mode != None && mode != DownOnly && mode != UpOnly && mode != UpAndDown {
		log.Error("invalid migration mode", slog.String("op", op), slog.String("mode", string(mode)))
	}

	m, err := migrate.New(cfg.MigrationSourceFilePath, cfg.StoragePath)
	if err != nil {
		log.Error("failed to create migrate instance", slog.String("op", op), sl.Err(err))
	}

	if mode == DownOnly || mode == UpAndDown {
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Error("failed to run down migration", slog.String("op", op), sl.Err(err))
		}
	}
	if mode == UpOnly || mode == UpAndDown {
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Error("failed to run up migration", slog.String("op", op), sl.Err(err))
		}
	}

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}

func MustSetupEnvVars() {
	var configPath string
	var migrationsMode, migrationsTable string
	flag.StringVar(&configPath, "config", "", "Path to config file")
	flag.StringVar(&migrationsMode, "migrations-mode", "none", "Migration mode: How to migrate the database. Options: up, down, up-and-down and none")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "Name of the migrations table")
	flag.Parse()
	if configPath != "" {
		err := os.Setenv("CONFIG_PATH", configPath)
		if err != nil {
			panic(err)
		}
	}
	err := os.Setenv("MIGRATIONS_MODE", migrationsMode)
	if err != nil {
		panic(err)
	}
	err = os.Setenv("MIGRATIONS_TABLE", migrationsTable)
	if err != nil {
		panic(err)
	}
}
