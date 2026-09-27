package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Postgres PostgresConfig
	Auth     AuthConfig
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
	postgres, err := LoadPostgres()
	if err != nil {
		return Config{}, err
	}
	auth, err := loadAuth()
	if err != nil {
		return Config{}, err
	}
	return Config{Postgres: postgres, Auth: auth}, nil
}

// LoadPostgres keeps database-only tools independent of API signing secrets.
func LoadPostgres() (PostgresConfig, error) {
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
			return PostgresConfig{}, fmt.Errorf("%s must be set", field.name)
		}
	}

	port, err := strconv.Atoi(cfg.Postgres.Port)
	if err != nil || port < 1 || port > 65535 {
		return PostgresConfig{}, fmt.Errorf("POSTGRES_PORT must be an integer between 1 and 65535")
	}
	return cfg.Postgres, nil
}

type AuthConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

func loadAuth() (AuthConfig, error) {
	cfg := AuthConfig{
		AccessSecret:  os.Getenv("JWT_ACCESS_SECRET"),
		RefreshSecret: os.Getenv("JWT_REFRESH_SECRET"),
	}
	for _, field := range []struct {
		name  string
		value *time.Duration
	}{
		{"JWT_ACCESS_TOKEN_TTL", &cfg.AccessTTL},
		{"JWT_REFRESH_TOKEN_TTL", &cfg.RefreshTTL},
	} {
		value, err := time.ParseDuration(os.Getenv(field.name))
		if err != nil {
			return AuthConfig{}, fmt.Errorf("%s must be a duration", field.name)
		}
		*field.value = value
	}
	return cfg, cfg.Validate()
}

func (c AuthConfig) Validate() error {
	if len(c.AccessSecret) < 32 || len(c.RefreshSecret) < 32 {
		return fmt.Errorf("JWT_ACCESS_SECRET and JWT_REFRESH_SECRET must each contain at least 32 bytes")
	}
	if c.AccessSecret == c.RefreshSecret {
		return fmt.Errorf("JWT access and refresh secrets must differ")
	}
	if c.AccessTTL < time.Second || c.AccessTTL > time.Hour || c.AccessTTL%time.Second != 0 {
		return fmt.Errorf("JWT_ACCESS_TOKEN_TTL must be whole seconds between 1s and 1h")
	}
	if c.RefreshTTL <= c.AccessTTL || c.RefreshTTL%time.Second != 0 {
		return fmt.Errorf("JWT_REFRESH_TOKEN_TTL must be whole seconds and longer than the access TTL")
	}
	return nil
}
