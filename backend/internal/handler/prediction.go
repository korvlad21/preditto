package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"preditto/internal/middleware"
	"preditto/internal/response"
	"preditto/internal/service"
)

type PredictionHandler struct {
	service *service.PredictionService
}

func NewPredictionHandler(service *service.PredictionService) *PredictionHandler {
	return &PredictionHandler{service: service}
}

// GetStages lists the selected tournament's stages with the active participant's progress.
// @Summary Get tournament stages for predictions
// @Tags Predictions
// @Description Uses the authenticated user's tournament selection and active participation. Stages are ordered by sequence_no then id. Stages without matches or predictions have zero counts. An empty tournament returns data: []. Cache-Control: no-store.
// @Produce json
// @Success 200 {object} apidocs.Success{data=[]model.TournamentStage}
// @Failure 401 {object} apidocs.Error "UNAUTHORIZED"
// @Failure 404 {object} apidocs.Error "USER_NOT_FOUND, TOURNAMENT_NOT_SELECTED or TOURNAMENT_PARTICIPANT_NOT_FOUND"
// @Failure 500 {object} apidocs.Error "INTERNAL_ERROR"
// @Security BearerAuth
// @Router /api/predictions/get_stages [post]
func (h *PredictionHandler) GetStages(c *gin.Context) {
	user, hasUser := middleware.CurrentUser(c)
	participant, hasParticipant := middleware.CurrentParticipant(c)
	if !hasUser || user.SelectedTournamentID == nil || !hasParticipant {
		response.AbortError(c, http.StatusInternalServerError, response.InternalError)
		return
	}
	stages, err := h.service.GetStages(c.Request.Context(), *user.SelectedTournamentID, participant.ID)
	if err != nil {
		_ = c.Error(err)
		response.AbortError(c, http.StatusInternalServerError, response.InternalError)
		return
	}
	response.Success(c, http.StatusOK, stages)
}
