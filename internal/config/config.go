package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Env                     string        `yaml:"env" env-required:"true"`
	TokenTTL                time.Duration `yaml:"token_ttl" env-required:"true"`
	GRPC                    `yaml:"grpc" env-required:"true"`
	Storage                 `yaml:"storage" env-required:"true"`
	MigrationSourceFilePath string `yaml:"migration_source_file_path" env-required:"true"`
}

type GRPC struct {
	Port        int           `yaml:"port" env-required:"true"`
	Timeout     time.Duration `yaml:"timeout" env-required:"true"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-required:"true"`
}

type Storage struct {
	DBType      string `yaml:"db_type" env-required:"true"`
	DBHost      string `yaml:"db_host" env-required:"true"`
	DBPort      int    `yaml:"db_port" env-required:"true"`
	DBSSL       string `yaml:"db_ssl" env-required:"true"`
	DBName      string `yaml:"db_name" env-required:"true"`
	DBUser      string `yaml:"db_user" env-required:"true"`
	DBPass      string `yaml:"db_pass" env:"DB_PASS"`
	StoragePath string
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("config path is not set")
	}

	return MustLoadByPath(configPath)
}

func MustLoadByPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cleanenv can not read config: %s", configPath)
	}

	cfg.StoragePath = fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Storage.DBType,
		cfg.Storage.DBUser,
		cfg.Storage.DBPass,
		cfg.Storage.DBHost,
		cfg.Storage.DBPort,
		cfg.Storage.DBName,
		cfg.Storage.DBSSL,
	)

	return &cfg
}
