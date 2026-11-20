package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"preditto/internal/model"
	"preditto/internal/repository"
	"preditto/internal/response"
)

const participantInfoKey = "participant_info"

type ParticipantInfoLoader interface {
	GetParticipantInfo(ctx context.Context, userID, tournamentID int64) (model.ParticipantInfo, error)
}

// LoadParticipant must run after RequireAuth and LoadUser.
func LoadParticipant(participants ParticipantInfoLoader) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := CurrentUser(c)
		if !ok {
			unauthorized(c)
			return
		}
		if user.SelectedTournamentID == nil {
			response.AbortError(c, http.StatusNotFound, response.TournamentNotSelected)
			return
		}
		participant, err := participants.GetParticipantInfo(c.Request.Context(), user.ID, *user.SelectedTournamentID)
		if errors.Is(err, repository.ErrNotFound) {
			response.AbortError(c, http.StatusNotFound, response.TournamentParticipantNotFound)
			return
		}
		if err != nil {
			_ = c.Error(err)
			response.AbortError(c, http.StatusInternalServerError, response.InternalError)
			return
		}
		c.Set(participantInfoKey, &participant)
		c.Next()
	}
}

func CurrentParticipant(c *gin.Context) (*model.ParticipantInfo, bool) {
	value, _ := c.Get(participantInfoKey)
	participant, ok := value.(*model.ParticipantInfo)
	return participant, ok && participant != nil
}
