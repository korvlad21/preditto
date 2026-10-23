package model

import "time"

const (
	UserStatusActive  = "ACTIVE"
	UserStatusBlocked = "BLOCKED"
)

type User struct {
	ID             int64  `json:"id"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	PasswordHash   string `json:"-"`
	Status         string `json:"status"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	FavoriteTeamID *int64 `json:"favorite_team_id"`
}

type UserInfo struct {
	ID                   int64     `json:"id"`
	Username             string    `json:"username"`
	Email                string    `json:"email"`
	Status               string    `json:"status"`
	FirstName            *string   `json:"first_name"`
	LastName             *string   `json:"last_name"`
	AvatarURL            *string   `json:"avatar_url"`
	FavoriteTeamID       *int64    `json:"favorite_team_id"`
	SelectedTournamentID *int64    `json:"selected_tournament_id"`
	CreatedAt            time.Time `json:"created_at"`
}
