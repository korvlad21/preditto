package users

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin/binding"
)

type UpdateUserRequest struct {
	UserID int64 `json:"user_id" binding:"required,gt=0" minimum:"1" format:"int64"`
	// Uniqueness follows the database's case-sensitive email constraint.
	Email string `json:"email" binding:"required,email,max=255" maxLength:"255" format:"email"`
	// Omitted/null optional fields are unchanged. Zero clears the favorite team.
	FavoriteTeamID *int64  `json:"favorite_team_id" binding:"omitempty,gte=0" minimum:"0" format:"int64"`
	FirstName      *string `json:"first_name" binding:"omitempty,max=100" maxLength:"100"`
	LastName       *string `json:"last_name" binding:"omitempty,max=100" maxLength:"100"`
	// Empty clears the avatar; a nonempty value must equal this user's stored URL.
	AvatarURL *string `json:"avatar_url" binding:"omitempty,max=500" maxLength:"500"`
}

func (r UpdateUserRequest) Validate() error {
	if err := binding.Validator.ValidateStruct(r); err != nil {
		return err
	}
	for _, value := range []*string{&r.Email, r.FirstName, r.LastName, r.AvatarURL} {
		if value != nil && strings.ContainsRune(*value, '\x00') {
			return errors.New("NUL is not allowed in PostgreSQL text")
		}
	}
	return nil
}
