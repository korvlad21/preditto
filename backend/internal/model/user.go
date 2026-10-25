package model

import "time"

const (
	UserStatusActive  = "ACTIVE"
	UserStatusBlocked = "BLOCKED"
)

type User struct {
	ID             int64  `json:"id" format:"int64"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	PasswordHash   string `json:"-"`
	Status         string `json:"status"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	FavoriteTeamID *int64 `json:"favorite_team_id" extensions:"x-nullable" format:"int64"`
}

type UserInfo struct {
	ID                   int64     `json:"id" format:"int64"`
	Username             string    `json:"username"`
	Email                string    `json:"email"`
	Status               string    `json:"status"`
	FirstName            *string   `json:"first_name" extensions:"x-nullable"`
	LastName             *string   `json:"last_name" extensions:"x-nullable"`
	AvatarURL            *string   `json:"avatar_url" extensions:"x-nullable"`
	FavoriteTeamID       *int64    `json:"favorite_team_id" extensions:"x-nullable" format:"int64"`
	SelectedTournamentID *int64    `json:"selected_tournament_id" extensions:"x-nullable" format:"int64"`
	CreatedAt            time.Time `json:"created_at" format:"date-time"`
}
