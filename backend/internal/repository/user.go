package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"preditto/internal/model"
)

const (
	pgCodeUniqueViolation     = "23505"
	pgCodeForeignKeyViolation = "23503"
)

var (
	ErrUsernameExists       = errors.New("username already exists")
	ErrEmailExists          = errors.New("email already exists")
	ErrFavoriteTeamNotFound = errors.New("favorite team not found")
)

type UserRepository struct {
	query queryDB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{query: db}
}

// CreateUser must be called within InTransaction so profile failures roll back users.
func (r *UserRepository) CreateUser(ctx context.Context, user *model.User) error {
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
		case pg.Code == pgCodeUniqueViolation && pg.Constraint == "users_username_key":
			return ErrUsernameExists
		case pg.Code == pgCodeUniqueViolation && pg.Constraint == "users_email_key":
			return ErrEmailExists
		case pg.Code == pgCodeForeignKeyViolation && pg.Constraint == "user_info_favorite_team_id_fkey":
			return ErrFavoriteTeamNotFound
		case pg.Code == pgCodeForeignKeyViolation && pg.Constraint == "user_info_selected_tournament_id_fkey":
			return ErrTournamentNotFound
		}
	}
	return err
}

// UpdateSelectedTournament updates an existing profile; nil clears the selection.
func (r *UserRepository) UpdateSelectedTournament(ctx context.Context, userID int64, tournamentID *int64) error {
	result, err := r.query.ExecContext(ctx, `
		UPDATE user_info SET selected_tournament_id = $1 WHERE user_id = $2
	`, tournamentID, userID)
	if err != nil {
		return fmt.Errorf("update selected tournament: %w", userWriteError(err))
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("selected tournament rows affected: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

const userSelect = `SELECT u.id, u.username, u.email, u.password_hash, u.status,
	COALESCE(i.first_name, ''), COALESCE(i.last_name, ''), i.favorite_team_id
	FROM users u LEFT JOIN user_info i ON i.user_id = u.id `

func (r *UserRepository) FindUserByLogin(ctx context.Context, login string) (model.User, error) {
	return scanUser(r.query.QueryRowContext(ctx, userSelect+`WHERE u.username = $1 OR u.email = $1`, login))
}

func (r *UserRepository) GetUserInfo(ctx context.Context, id int64) (model.UserInfo, error) {
	var user model.UserInfo
	err := r.query.QueryRowContext(ctx, `
		SELECT u.id, u.username, u.email, u.status, u.created_at,
			i.first_name, i.last_name, i.avatar_url, i.favorite_team_id, i.selected_tournament_id
		FROM users u LEFT JOIN user_info i ON i.user_id = u.id
		WHERE u.id = $1
	`, id).Scan(&user.ID, &user.Username, &user.Email, &user.Status, &user.CreatedAt,
		&user.FirstName, &user.LastName, &user.AvatarURL, &user.FavoriteTeamID, &user.SelectedTournamentID)
	if errors.Is(err, sql.ErrNoRows) {
		return model.UserInfo{}, ErrNotFound
	}
	if err != nil {
		return model.UserInfo{}, fmt.Errorf("get user info: %w", err)
	}
	return user, nil
}

// LockUser prevents status changes while a new session is being issued.
func (r *UserRepository) LockUser(ctx context.Context, id int64) (model.User, error) {
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
