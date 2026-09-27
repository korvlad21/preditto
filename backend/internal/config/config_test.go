package config

import (
	"strings"
	"testing"
	"time"
)

func postgresEnv() map[string]string {
	return map[string]string{
		"POSTGRES_HOST": "db.example.test", "POSTGRES_PORT": "5544",
		"POSTGRES_USER": "app", "POSTGRES_PASSWORD": " p@ss:'\\/?#% ",
		"POSTGRES_DB": "app_db", "POSTGRES_SSLMODE": "verify-full",
	}
}

func TestLoad(t *testing.T) {
	setAuthEnv(t)
	for key, value := range postgresEnv() {
		t.Setenv(key, value)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	want := PostgresConfig{
		Host: "db.example.test", Port: "5544", User: "app",
		Password: " p@ss:'\\/?#% ", Database: "app_db", SSLMode: "verify-full",
	}
	if cfg.Postgres != want {
		t.Fatal("PostgreSQL configuration does not preserve environment values")
	}
}

func setAuthEnv(t *testing.T) {
	t.Helper()
	t.Setenv("JWT_ACCESS_SECRET", strings.Repeat("a", 32))
	t.Setenv("JWT_REFRESH_SECRET", strings.Repeat("r", 32))
	t.Setenv("JWT_ACCESS_TOKEN_TTL", "15m")
	t.Setenv("JWT_REFRESH_TOKEN_TTL", "720h")
}

func TestAuthConfig(t *testing.T) {
	setAuthEnv(t)
	cfg, err := loadAuth()
	if err != nil || cfg.AccessTTL != 15*time.Minute || cfg.RefreshTTL != 720*time.Hour {
		t.Fatalf("unexpected auth configuration: %v", err)
	}
	for _, tc := range []struct{ key, value string }{
		{"JWT_ACCESS_SECRET", ""}, {"JWT_ACCESS_SECRET", "short"},
		{"JWT_REFRESH_SECRET", ""}, {"JWT_REFRESH_SECRET", strings.Repeat("a", 32)},
		{"JWT_ACCESS_TOKEN_TTL", ""}, {"JWT_ACCESS_TOKEN_TTL", "bad"},
		{"JWT_ACCESS_TOKEN_TTL", "0s"}, {"JWT_ACCESS_TOKEN_TTL", "-1m"},
		{"JWT_ACCESS_TOKEN_TTL", "2h"}, {"JWT_ACCESS_TOKEN_TTL", "1500ms"},
		{"JWT_REFRESH_TOKEN_TTL", "15m"}, {"JWT_REFRESH_TOKEN_TTL", "1m"},
		{"JWT_REFRESH_TOKEN_TTL", ""},
	} {
		t.Run(tc.key+"/"+tc.value, func(t *testing.T) {
			setAuthEnv(t)
			t.Setenv(tc.key, tc.value)
			if _, err := loadAuth(); err == nil {
				t.Fatal("invalid configuration accepted")
			}
		})
	}
}

func TestDatabaseOnlyConfigDoesNotRequireJWT(t *testing.T) {
	for key, value := range postgresEnv() {
		t.Setenv(key, value)
	}
	t.Setenv("JWT_ACCESS_SECRET", "")
	t.Setenv("JWT_REFRESH_SECRET", "")
	if _, err := LoadPostgres(); err != nil {
		t.Fatal(err)
	}
}

func TestLoadRequiresPostgresEnvironment(t *testing.T) {
	for missing := range postgresEnv() {
		t.Run(missing, func(t *testing.T) {
			for key, value := range postgresEnv() {
				t.Setenv(key, value)
			}
			t.Setenv(missing, "")
			if _, err := Load(); err == nil || !strings.Contains(err.Error(), missing) {
				t.Fatalf("expected an error identifying %s, got %v", missing, err)
			}
		})
	}
}

func TestLoadRejectsInvalidPort(t *testing.T) {
	for _, port := range []string{"abc", "0", "-1", "65536"} {
		t.Run(port, func(t *testing.T) {
			for key, value := range postgresEnv() {
				t.Setenv(key, value)
			}
			t.Setenv("POSTGRES_PORT", port)
			if _, err := Load(); err == nil || !strings.Contains(err.Error(), "POSTGRES_PORT") {
				t.Fatalf("expected a port validation error, got %v", err)
			}
		})
	}
}
