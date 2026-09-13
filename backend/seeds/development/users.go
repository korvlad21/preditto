package development

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type user struct {
	username string
	email    string
	status   string
}

var users = []user{
	{"korvlad21", "vladek2000@inbox.ru", "ACTIVE"},
	{"hanna_stoma", "hannastoma9@gmail.com", "ACTIVE"},
	{"boroda", "borodin_nikita@yandex.ru", "BLOCKED"},
}

func seedUsers(ctx context.Context, tx *sql.Tx) error {
	for _, item := range users {
		hash, err := bcrypt.GenerateFromPassword([]byte("12345678"), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash password for user %q: %w", item.username, err)
		}

		var id int64
		err = tx.QueryRowContext(ctx, `
			SELECT id FROM users WHERE username = $1 OR email = $2
		`, item.username, item.email).Scan(&id)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			_, err = tx.ExecContext(ctx, `
				INSERT INTO users (username, email, password_hash, status)
				VALUES ($1, $2, $3, $4)
			`, item.username, item.email, string(hash), item.status)
		case err != nil:
			return fmt.Errorf("find user %q: %w", item.username, err)
		default:
			_, err = tx.ExecContext(ctx, `
				UPDATE users
				SET username = $1, email = $2, password_hash = $3, status = $4
				WHERE id = $5
			`, item.username, item.email, string(hash), item.status, id)
		}
		if err != nil {
			return fmt.Errorf("upsert user %q: %w", item.username, err)
		}
	}
	return nil
}
