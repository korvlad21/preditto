package model

import "time"

type RefreshToken struct {
	ID        string
	UserID    int64
	TokenHash []byte `json:"-"`
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt *time.Time
}
