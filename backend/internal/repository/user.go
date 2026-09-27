package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"preditto/internal/model"
)

var (
	ErrUsernameExists       = errors.New("username already exists")
	ErrEmailExists          = errors.New("email already exists")
	ErrFavoriteTeamNotFound = errors.New("favorite team not found")
)

// CreateUser must be called within InTransaction so profile failures roll back users.
func (r *AuthRepository) CreateUser(ctx context.Context, user *model.User) error {
	err := r.query.QueryRowContext(ctx, `
		INSERT INTO users (username, email, password_hash, status)
		VALUES ($1, $2, $3, $4) RETURNING id
	`, user.Username, user.Email, user.PasswordHash, user.Status).Scan(&user.ID)
	if err != nil {
		return userWriteError(err)
	}
	_, err = r.query.ExecContext(ctx, `
		INSERT INTO user_info (user_id, first_name, last_name, favorite_team_id)
		VALUES ($1, $2, $3, $4)
	`, user.ID, user.FirstName, user.LastName, user.FavoriteTeamID)
	return userWriteError(err)
}

func userWriteError(err error) error {
	var pg *pq.Error
	if errors.As(err, &pg) {
		switch {
		case pg.Code == "23505" && pg.Constraint == "users_username_key":
			return ErrUsernameExists
		case pg.Code == "23505" && pg.Constraint == "users_email_key":
			return ErrEmailExists
		case pg.Code == "23503" && pg.Constraint == "user_info_favorite_team_id_fkey":
			return ErrFavoriteTeamNotFound
		}
	}
	return err
}

const userSelect = `SELECT u.id, u.username, u.email, u.password_hash, u.status,
	COALESCE(i.first_name, ''), COALESCE(i.last_name, ''), i.favorite_team_id
	FROM users u LEFT JOIN user_info i ON i.user_id = u.id `

func (r *AuthRepository) FindUserByLogin(ctx context.Context, login string) (model.User, error) {
	return scanUser(r.query.QueryRowContext(ctx, userSelect+`WHERE u.username = $1 OR u.email = $1`, login))
}

// LockUser prevents status changes while a new session is being issued.
func (r *AuthRepository) LockUser(ctx context.Context, id int64) (model.User, error) {
	return scanUser(r.query.QueryRowContext(ctx, userSelect+`WHERE u.id = $1 FOR SHARE OF u`, id))
}

func scanUser(row *sql.Row) (model.User, error) {
	var user model.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Status,
		&user.FirstName, &user.LastName, &user.FavoriteTeamID)
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, ErrNotFound
	}
	return user, err
}
