package config

import (
	"strings"
	"testing"
)

func postgresEnv() map[string]string {
	return map[string]string{
		"POSTGRES_HOST": "db.example.test", "POSTGRES_PORT": "5544",
		"POSTGRES_USER": "app", "POSTGRES_PASSWORD": " p@ss:'\\/?#% ",
		"POSTGRES_DB": "app_db", "POSTGRES_SSLMODE": "verify-full",
	}
}

func TestLoad(t *testing.T) {
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
