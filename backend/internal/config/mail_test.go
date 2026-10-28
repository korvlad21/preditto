package config

import (
	"strings"
	"testing"
	"time"
)

func setMailEnv(t *testing.T) {
	t.Helper()
	for key, value := range map[string]string{
		"MAIL_ENABLED": "true", "MAIL_FROM_ADDRESS": "sender@example.com", "MAIL_FROM_NAME": "Preditto",
		"SMTP_HOST": "smtp.mail.ru", "SMTP_PORT": "465", "SMTP_USERNAME": "sender@example.com",
		"SMTP_PASSWORD": "test-secret-$", "SMTP_TLS_MODE": "tls", "SMTP_AUTH_TYPE": "plain", "SMTP_TIMEOUT": "10s",
	} {
		t.Setenv(key, value)
	}
}

func TestLoadMail(t *testing.T) {
	setMailEnv(t)
	cfg, err := LoadMail()
	if err != nil || !cfg.Enabled || cfg.Host != "smtp.mail.ru" || cfg.Port != 465 ||
		cfg.Timeout != 10*time.Second || cfg.Password != "test-secret-$" || cfg.TLSMode != "tls" {
		t.Fatal("mail configuration was not preserved", err)
	}
	// Invalid configuration errors identify settings without printing their values.
	for _, tc := range []struct{ key, value string }{
		{"MAIL_ENABLED", "invalid"}, {"SMTP_HOST", ""}, {"SMTP_HOST", "smtp://mail.ru"},
		{"SMTP_PORT", "bad"}, {"SMTP_PORT", "0"}, {"SMTP_PORT", "65536"},
		{"SMTP_USERNAME", ""}, {"SMTP_PASSWORD", ""}, {"MAIL_FROM_ADDRESS", "bad"},
		{"MAIL_FROM_ADDRESS", "Name <sender@example.com>"}, {"MAIL_FROM_NAME", "Name\r\nBcc: victim@example.com"},
		{"SMTP_TLS_MODE", "none"}, {"SMTP_TLS_MODE", "opportunistic"}, {"SMTP_AUTH_TYPE", "none"},
		{"SMTP_TIMEOUT", "bad"}, {"SMTP_TIMEOUT", "0s"}, {"SMTP_TIMEOUT", "2m"},
	} {
		t.Run(tc.key+"/"+tc.value, func(t *testing.T) {
			setMailEnv(t)
			t.Setenv(tc.key, tc.value)
			_, err := LoadMail()
			if err == nil || strings.Contains(err.Error(), "test-secret-$") {
				t.Fatal("invalid configuration accepted or secret exposed")
			}
		})
	}
}

func TestMailDisabledAndIndependentLoaders(t *testing.T) {
	setMailEnv(t)
	t.Setenv("MAIL_ENABLED", "false")
	t.Setenv("SMTP_PASSWORD", "")
	if cfg, err := LoadMail(); err != nil || cfg.Enabled {
		t.Fatal("disabled mail must not require credentials", err)
	}
	t.Setenv("MAIL_ENABLED", "")
	if cfg, err := LoadMail(); err != nil || cfg.Enabled {
		t.Fatal("unset mail must be disabled", err)
	}
	setMailEnv(t)
	t.Setenv("JWT_ACCESS_SECRET", "")
	t.Setenv("POSTGRES_HOST", "")
	if _, err := LoadMail(); err != nil {
		t.Fatal("mail must not require database or JWT settings", err)
	}
	for key, value := range postgresEnv() {
		t.Setenv(key, value)
	}
	t.Setenv("SMTP_PASSWORD", "")
	if _, err := LoadPostgres(); err != nil {
		t.Fatal("seed configuration must not require SMTP credentials", err)
	}
}

func TestLoadIncludesMail(t *testing.T) {
	setMailEnv(t)
	setAuthEnv(t)
	for key, value := range postgresEnv() {
		t.Setenv(key, value)
	}
	cfg, err := Load()
	if err != nil || !cfg.Mail.Enabled {
		t.Fatal("API configuration did not load mail", err)
	}
	t.Setenv("SMTP_PASSWORD", "")
	if _, err := Load(); err == nil {
		t.Fatal("API must reject enabled but incomplete mail configuration")
	}
}
