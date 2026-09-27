package auth

import (
	"errors"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin/binding"
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_.-]{3,50}$`)

type RegisterRequest struct {
	Username             string `json:"username" binding:"required,min=3,max=50"`
	FirstName            string `json:"first_name" binding:"required,max=100"`
	LastName             string `json:"last_name" binding:"required,max=100"`
	Email                string `json:"email" binding:"required,email,max=255"`
	Password             string `json:"password" binding:"required,min=8"`
	PasswordConfirmation string `json:"password_confirmation" binding:"required,eqfield=Password"`
	FavoriteTeamID       *int64 `json:"favorite_team_id" binding:"omitempty,gt=0"`
}

func (r RegisterRequest) Validate() error {
	if err := binding.Validator.ValidateStruct(r); err != nil {
		return err
	}
	if !usernamePattern.MatchString(r.Username) || strings.TrimSpace(r.FirstName) == "" || strings.TrimSpace(r.LastName) == "" || len(r.Password) > 72 {
		return errors.New("invalid registration fields")
	}
	return nil
}

type LoginRequest struct {
	Login    string `json:"login" binding:"required,max=255"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required,max=4096"`
}
