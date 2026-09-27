package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"preditto/internal/model"
)

var ErrNotFound = errors.New("record not found")

type authDB interface {
	executor
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type AuthRepository struct {
	db    *sql.DB
	query authDB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db, query: db}
}

// InTransaction gives the service transaction-scoped persistence operations.
func (r *AuthRepository) InTransaction(ctx context.Context, fn func(*AuthRepository) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin auth transaction: %w", err)
	}
	defer tx.Rollback()
	if err := fn(&AuthRepository{query: tx}); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit auth transaction: %w", err)
	}
	return nil
}

func (r *AuthRepository) CreateRefreshToken(ctx context.Context, token model.RefreshToken) error {
	_, err := r.query.ExecContext(ctx, `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.CreatedAt)
	return err
}

// LockRefreshToken serializes refresh and logout for the same token.
func (r *AuthRepository) LockRefreshToken(ctx context.Context, id string) (model.RefreshToken, error) {
	var token model.RefreshToken
	err := r.query.QueryRowContext(ctx, `
		SELECT id, user_id, token_hash, expires_at, created_at, revoked_at
		FROM refresh_tokens WHERE id = $1 FOR UPDATE
	`, id).Scan(&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt, &token.CreatedAt, &token.RevokedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.RefreshToken{}, ErrNotFound
	}
	return token, err
}

func (r *AuthRepository) RevokeRefreshToken(ctx context.Context, id string) error {
	_, err := r.query.ExecContext(ctx, `
		UPDATE refresh_tokens SET revoked_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND revoked_at IS NULL
	`, id)
	return err
}
