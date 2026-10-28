package config

import (
	"fmt"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"
)

// MailConfig contains SMTP settings. Never log this value: it holds credentials.
type MailConfig struct {
	Enabled     bool
	Host        string
	Port        int
	Username    string
	Password    string
	FromAddress string
	FromName    string
	TLSMode     string
	AuthType    string
	Timeout     time.Duration
}

// LoadMail is independent of database and JWT configuration for command-line tools.
// Like the other loaders, it reads process environment; Compose loads the root .env.
func LoadMail() (MailConfig, error) {
	enabled := os.Getenv("MAIL_ENABLED")
	if enabled == "" {
		enabled = "false"
	}
	on, err := strconv.ParseBool(enabled)
	if err != nil {
		return MailConfig{}, fmt.Errorf("MAIL_ENABLED must be a boolean")
	}
	if !on {
		return MailConfig{}, nil
	}
	cfg := MailConfig{
		Enabled: true, Host: os.Getenv("SMTP_HOST"), Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"), FromAddress: os.Getenv("MAIL_FROM_ADDRESS"),
		FromName: os.Getenv("MAIL_FROM_NAME"), TLSMode: os.Getenv("SMTP_TLS_MODE"),
		AuthType: os.Getenv("SMTP_AUTH_TYPE"),
	}
	cfg.Port, err = strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return MailConfig{}, fmt.Errorf("SMTP_PORT must be an integer between 1 and 65535")
	}
	cfg.Timeout, err = time.ParseDuration(os.Getenv("SMTP_TIMEOUT"))
	if err != nil {
		return MailConfig{}, fmt.Errorf("SMTP_TIMEOUT must be a duration")
	}
	return cfg, cfg.Validate()
}

func (c MailConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.Host == "" || strings.ContainsAny(c.Host, " \t\r\n/:@") {
		return fmt.Errorf("SMTP_HOST must be a hostname without a scheme or port")
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("SMTP_PORT must be an integer between 1 and 65535")
	}
	if c.Username == "" || c.Password == "" {
		return fmt.Errorf("SMTP_USERNAME and SMTP_PASSWORD must be set")
	}
	address, err := mail.ParseAddress(c.FromAddress)
	if err != nil || address.Address != c.FromAddress || strings.ContainsAny(c.FromAddress+c.FromName, "\r\n\x00") {
		return fmt.Errorf("MAIL_FROM_ADDRESS must be a single email address and sender fields must not contain control characters")
	}
	if c.TLSMode != "tls" && c.TLSMode != "starttls" {
		return fmt.Errorf("SMTP_TLS_MODE must be tls or starttls")
	}
	if c.AuthType != "plain" && c.AuthType != "login" {
		return fmt.Errorf("SMTP_AUTH_TYPE must be plain or login")
	}
	if c.Timeout <= 0 || c.Timeout > time.Minute {
		return fmt.Errorf("SMTP_TIMEOUT must be positive and no longer than 1m")
	}
	return nil
}
