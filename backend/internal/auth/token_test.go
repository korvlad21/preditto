package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"preditto/internal/config"
)

func TestTokenPair(t *testing.T) {
	cfg := config.AuthConfig{AccessSecret: strings.Repeat("a", 32), RefreshSecret: strings.Repeat("r", 32), AccessTTL: 15 * time.Minute, RefreshTTL: 720 * time.Hour, ShortRefreshTTL: 12 * time.Hour}
	m, err := NewTokenManager(cfg)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Truncate(time.Second)
	m.now = func() time.Time { return now }
	pair, session, err := m.GeneratePair(81, nil)
	if err != nil {
		t.Fatal(err)
	}
	access, err := m.ParseAccessToken(pair.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	refresh, err := m.ParseRefreshToken(pair.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if access.UserID() != 81 || refresh.UserID() != 81 || access.ID == refresh.ID || refresh.ID != session.ID || access.ExpiresAt.Sub(now) != cfg.AccessTTL || refresh.ExpiresAt.Sub(now) != cfg.RefreshTTL || pair.ExpiresIn != 900 || pair.TokenType != "Bearer" {
		t.Fatal("incorrect token pair claims")
	}
	if !bytes.Equal(session.TokenHash, HashRefreshToken(pair.RefreshToken)) || len(session.TokenHash) != 32 {
		t.Fatal("refresh session does not contain a SHA-256 hash")
	}
	encoded, _ := json.Marshal(access)
	var fields map[string]any
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	if len(fields) != 5 {
		t.Fatalf("unexpected JWT fields: %s", encoded)
	}
	for _, key := range []string{"sub", "iat", "exp", "jti", "token_type"} {
		if _, ok := fields[key]; !ok {
			t.Fatalf("missing %s", key)
		}
	}
	other, _, err := m.GeneratePair(81, nil)
	if err != nil || other.RefreshToken == pair.RefreshToken || other.AccessToken == pair.AccessToken {
		t.Fatal("tokens must be unique")
	}
	if _, err := m.ParseAccessToken(pair.RefreshToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatal("refresh accepted as access")
	}
	if _, err := m.ParseRefreshToken(pair.AccessToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatal("access accepted as refresh")
	}
	m.now = func() time.Time { return now.Add(cfg.AccessTTL) }
	if _, err := m.ParseAccessToken(pair.AccessToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatal("expired access accepted")
	}
	if _, err := m.ParseRefreshToken(pair.RefreshToken); err != nil {
		t.Fatal(err)
	}
	m.now = func() time.Time { return now.Add(cfg.RefreshTTL) }
	if _, err := m.ParseRefreshToken(pair.RefreshToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatal("expired refresh accepted")
	}
}

func TestRejectMalformedClaims(t *testing.T) {
	cfg := config.AuthConfig{AccessSecret: strings.Repeat("a", 32), RefreshSecret: strings.Repeat("r", 32), AccessTTL: time.Minute, RefreshTTL: time.Hour, ShortRefreshTTL: 30 * time.Minute}
	m, err := NewTokenManager(cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		change func(*Claims)
		method jwt.SigningMethod
		secret string
	}{
		{"wrong type with correct key", func(c *Claims) { c.TokenType = "refresh" }, jwt.SigningMethodHS256, cfg.AccessSecret},
		{"missing type", func(c *Claims) { c.TokenType = "" }, jwt.SigningMethodHS256, cfg.AccessSecret},
		{"missing expiration", func(c *Claims) { c.ExpiresAt = nil }, jwt.SigningMethodHS256, cfg.AccessSecret},
		{"missing issued at", func(c *Claims) { c.IssuedAt = nil }, jwt.SigningMethodHS256, cfg.AccessSecret},
		{"future issued at", func(c *Claims) { c.IssuedAt = jwt.NewNumericDate(time.Now().Add(time.Hour)) }, jwt.SigningMethodHS256, cfg.AccessSecret},
		{"missing jti", func(c *Claims) { c.ID = "" }, jwt.SigningMethodHS256, cfg.AccessSecret},
		{"invalid subject", func(c *Claims) { c.Subject = "abc" }, jwt.SigningMethodHS256, cfg.AccessSecret},
		{"overflowing subject", func(c *Claims) { c.Subject = "9223372036854775808" }, jwt.SigningMethodHS256, cfg.AccessSecret},
		{"zero subject", func(c *Claims) { c.Subject = "0" }, jwt.SigningMethodHS256, cfg.AccessSecret},
		{"wrong algorithm", func(*Claims) {}, jwt.SigningMethodHS384, cfg.AccessSecret},
		{"wrong signature", func(*Claims) {}, jwt.SigningMethodHS256, strings.Repeat("z", 32)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := Claims{TokenType: "access", RegisteredClaims: jwt.RegisteredClaims{Subject: "81", ID: "unique", IssuedAt: jwt.NewNumericDate(time.Now().Add(-time.Second)), ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))}}
			tc.change(&c)
			raw, err := jwt.NewWithClaims(tc.method, c).SignedString([]byte(tc.secret))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := m.ParseAccessToken(raw); !errors.Is(err, ErrInvalidToken) {
				t.Fatalf("accepted malformed claims: %v", err)
			}
		})
	}
	for _, raw := range []string{"", "not.jwt", strings.Repeat("a", 4097)} {
		if _, err := m.ParseAccessToken(raw); !errors.Is(err, ErrInvalidToken) {
			t.Fatal("accepted malformed JWT")
		}
	}
}

func TestRememberMeTokenLifetime(t *testing.T) {
	// Non-default durations prove that issuance uses configuration.
	cfg := config.AuthConfig{AccessSecret: strings.Repeat("a", 32), RefreshSecret: strings.Repeat("r", 32), AccessTTL: time.Minute, RefreshTTL: 48 * time.Hour, ShortRefreshTTL: 2 * time.Hour}
	remember, short := true, false
	for _, tc := range []struct {
		name string
		mode *bool
		ttl  time.Duration
	}{
		{"legacy", nil, cfg.RefreshTTL}, {"remember", &remember, cfg.RefreshTTL}, {"short", &short, cfg.ShortRefreshTTL},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, err := NewTokenManager(cfg)
			if err != nil {
				t.Fatal(err)
			}
			now := time.Now().UTC().Truncate(time.Second)
			m.now = func() time.Time { return now }
			pair, session, err := m.GeneratePair(81, tc.mode)
			if err != nil {
				t.Fatal(err)
			}
			claims, err := m.ParseRefreshToken(pair.RefreshToken)
			if err != nil {
				t.Fatal(err)
			}
			if (claims.RememberMe == nil) != (tc.mode == nil) || (tc.mode != nil && *claims.RememberMe != *tc.mode) {
				t.Fatal("mode lost")
			}
			if claims.ExpiresAt.Sub(claims.IssuedAt.Time) != tc.ttl || !session.ExpiresAt.Equal(now.Add(tc.ttl)) {
				t.Fatal("incorrect refresh lifetime")
			}
			access, err := m.ParseAccessToken(pair.AccessToken)
			if err != nil || access.RememberMe != nil || access.ExpiresAt.Sub(access.IssuedAt.Time) != cfg.AccessTTL || pair.ExpiresIn != 60 {
				t.Fatal("access contract changed")
			}
			m.now = func() time.Time { return now.Add(tc.ttl - time.Second) }
			if _, err := m.ParseRefreshToken(pair.RefreshToken); err != nil {
				t.Fatal("refresh expired early")
			}
			m.now = func() time.Time { return now.Add(tc.ttl) }
			if _, err := m.ParseRefreshToken(pair.RefreshToken); !errors.Is(err, ErrInvalidToken) {
				t.Fatal("expired refresh accepted")
			}
		})
	}
}

func TestRememberMeTampering(t *testing.T) {
	cfg := config.AuthConfig{AccessSecret: strings.Repeat("a", 32), RefreshSecret: strings.Repeat("r", 32), AccessTTL: time.Minute, ShortRefreshTTL: 12 * time.Hour, RefreshTTL: 720 * time.Hour}
	m, err := NewTokenManager(cfg)
	if err != nil {
		t.Fatal(err)
	}
	mode := false
	pair, _, err := m.GeneratePair(81, &mode)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(pair.RefreshToken, ".")
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatal(err)
	}
	changed := bytes.Replace(payload, []byte(`"remember_me":false`), []byte(`"remember_me":true`), 1)
	if bytes.Equal(payload, changed) {
		t.Fatal("mode absent from refresh")
	}
	parts[1] = base64.RawURLEncoding.EncodeToString(changed)
	if _, err := m.ParseRefreshToken(strings.Join(parts, ".")); !errors.Is(err, ErrInvalidToken) {
		t.Fatal("tampered mode accepted")
	}
}
