package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"preditto/internal/config"
	authdto "preditto/internal/dto/auth"
	"preditto/internal/model"
)

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func (c Claims) UserID() int64 {
	id, err := strconv.ParseInt(c.Subject, 10, 64)
	if err != nil {
		return 0
	}
	return id
}

type TokenManager struct {
	cfg config.AuthConfig
	now func() time.Time
}

func NewTokenManager(cfg config.AuthConfig) (*TokenManager, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &TokenManager{cfg: cfg, now: time.Now}, nil
}

// GeneratePair returns the wire tokens and the only refresh data persisted to SQL.
func (m *TokenManager) GeneratePair(userID int64) (authdto.TokenResponse, model.RefreshToken, error) {
	now := m.now().UTC().Truncate(time.Second)
	access, _, err := m.generate(userID, "access", m.cfg.AccessSecret, now, m.cfg.AccessTTL)
	if err != nil {
		return authdto.TokenResponse{}, model.RefreshToken{}, err
	}
	refresh, claims, err := m.generate(userID, "refresh", m.cfg.RefreshSecret, now, m.cfg.RefreshTTL)
	if err != nil {
		return authdto.TokenResponse{}, model.RefreshToken{}, err
	}
	return authdto.TokenResponse{
			AccessToken: access, RefreshToken: refresh, TokenType: "Bearer", ExpiresIn: int64(m.cfg.AccessTTL / time.Second),
		}, model.RefreshToken{
			ID: claims.ID, UserID: userID, TokenHash: HashRefreshToken(refresh), ExpiresAt: claims.ExpiresAt.Time, CreatedAt: now,
		}, nil
}

func (m *TokenManager) generate(userID int64, kind, secret string, now time.Time, ttl time.Duration) (string, Claims, error) {
	if userID <= 0 {
		return "", Claims{}, ErrInvalidToken
	}
	var random [32]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", Claims{}, err
	}
	claims := Claims{TokenType: kind, RegisteredClaims: jwt.RegisteredClaims{
		Subject: strconv.FormatInt(userID, 10), ID: hex.EncodeToString(random[:]),
		IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	}}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	return token, claims, err
}

func (m *TokenManager) ParseAccessToken(raw string) (*Claims, error) {
	return m.parse(raw, "access", m.cfg.AccessSecret)
}

func (m *TokenManager) ParseRefreshToken(raw string) (*Claims, error) {
	return m.parse(raw, "refresh", m.cfg.RefreshSecret)
}

func (m *TokenManager) parse(raw, kind, secret string) (*Claims, error) {
	if len(raw) > 4096 {
		return nil, ErrInvalidToken
	}
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(_ *jwt.Token) (any, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithTimeFunc(m.now))
	if err != nil || !token.Valid || claims.TokenType != kind || claims.UserID() <= 0 || claims.ID == "" || claims.IssuedAt == nil || !claims.ExpiresAt.After(claims.IssuedAt.Time) {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// SHA-256 is appropriate for high-entropy signed tokens; passwords use bcrypt.
func HashRefreshToken(raw string) []byte {
	hash := sha256.Sum256([]byte(raw))
	return hash[:]
}
