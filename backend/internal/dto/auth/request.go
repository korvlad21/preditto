package auth

import (
	"errors"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin/binding"
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_.-]{3,50}$`)

type RegisterRequest struct {
	// ASCII letters, digits, underscore, dot or hyphen only (3-50).
	Username string `json:"username" binding:"required,min=3,max=50" minLength:"3" maxLength:"50" example:"swagger_user"`
	// Required; must not consist only of whitespace.
	FirstName string `json:"first_name" binding:"required,max=100" maxLength:"100" example:"Alex"`
	// Required; must not consist only of whitespace.
	LastName string `json:"last_name" binding:"required,max=100" maxLength:"100" example:"Smith"`
	// Valid email address; uniqueness is case-sensitive.
	Email string `json:"email" binding:"required,email,max=255" maxLength:"255" format:"email" example:"alex@example.com"`
	// At least 8 characters and at most 72 UTF-8 bytes.
	Password string `json:"password" binding:"required,min=8" format:"password" example:"example-password" minLength:"8"`
	// Must exactly match password.
	PasswordConfirmation string `json:"password_confirmation" binding:"required,eqfield=Password" format:"password" example:"example-password"`
	// Optional positive existing team ID; null is accepted.
	FavoriteTeamID *int64 `json:"favorite_team_id" binding:"omitempty,gt=0" minimum:"1" format:"int64" extensions:"x-nullable" example:"1"`
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
	// Username or email, matched case-sensitively.
	Login string `json:"login" binding:"required,max=255" maxLength:"255" example:"swagger_user"`
	// Required nonempty password; no registration length constraints are applied here.
	Password string `json:"password" binding:"required" format:"password" example:"example-password"`
	// False: short session; true: long session; omitted/null: legacy long session.
	RememberMe *bool `json:"remember_me,omitempty" extensions:"x-nullable" example:"false"`
}

type RefreshRequest struct {
	// Current refresh JWT, not an access JWT.
	RefreshToken string `json:"refresh_token" binding:"required,max=4096" maxLength:"4096" example:"<refresh_token>"`
}
