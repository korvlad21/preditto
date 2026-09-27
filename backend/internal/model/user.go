package model

const UserStatusActive = "ACTIVE"

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
