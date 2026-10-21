package response

import "github.com/gin-gonic/gin"

// Public error codes are stable translation keys for API clients.
const (
	InvalidRequest         = "INVALID_REQUEST"
	InvalidCredentials     = "INVALID_CREDENTIALS"
	InvalidToken           = "INVALID_TOKEN"
	Unauthorized           = "UNAUTHORIZED"
	UserBlocked            = "USER_BLOCKED"
	UsernameAlreadyExists  = "USERNAME_ALREADY_EXISTS"
	EmailAlreadyExists     = "EMAIL_ALREADY_EXISTS"
	FavoriteTeamNotFound   = "FAVORITE_TEAM_NOT_FOUND"
	UserNotFound           = "USER_NOT_FOUND"
	UserInfoNotFound       = "USER_INFO_NOT_FOUND"
	TournamentIDRequired   = "TOURNAMENT_ID_REQUIRED"
	InvalidTournamentID    = "INVALID_TOURNAMENT_ID"
	TournamentNotFound     = "TOURNAMENT_NOT_FOUND"
	TournamentUpdateFailed = "TOURNAMENT_UPDATE_FAILED"
	InternalError          = "INTERNAL_ERROR"
	NotFound               = "NOT_FOUND"
)

func AbortError(c *gin.Context, status int, code string) {
	c.AbortWithStatusJSON(status, struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}{Success: false, Error: code})
}
