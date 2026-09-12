package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Postgres PostgresConfig
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

func Load() (Config, error) {
	cfg := Config{}
	for _, field := range []struct {
		name  string
		value *string
	}{
		{"POSTGRES_HOST", &cfg.Postgres.Host},
		{"POSTGRES_PORT", &cfg.Postgres.Port},
		{"POSTGRES_USER", &cfg.Postgres.User},
		{"POSTGRES_PASSWORD", &cfg.Postgres.Password},
		{"POSTGRES_DB", &cfg.Postgres.Database},
		{"POSTGRES_SSLMODE", &cfg.Postgres.SSLMode},
	} {
		*field.value = os.Getenv(field.name)
		if *field.value == "" {
			return Config{}, fmt.Errorf("%s must be set", field.name)
		}
	}

	port, err := strconv.Atoi(cfg.Postgres.Port)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("POSTGRES_PORT must be an integer between 1 and 65535")
	}
	return cfg, nil
}
