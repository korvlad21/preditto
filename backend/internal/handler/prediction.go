package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	predictionsdto "preditto/internal/dto/predictions"
	"preditto/internal/middleware"
	"preditto/internal/repository"
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

// GetStage returns one stage in the selected tournament with the active participant's progress.
// @Summary Get a tournament stage for predictions
// @Tags Predictions
// @Description Uses the authenticated user's tournament selection and active participation. The stage ID is read from the JSON body. Stages without matches or predictions have zero counts. Cache-Control: no-store.
// @Accept json
// @Param request body predictionsdto.GetStageRequest true "JSON request body"
// @Produce json
// @Success 200 {object} apidocs.Success{data=model.TournamentStage}
// @Failure 400 {object} apidocs.Error "INVALID_REQUEST (invalid body or missing/invalid id)"
// @Failure 401 {object} apidocs.Error "UNAUTHORIZED"
// @Failure 404 {object} apidocs.Error "USER_NOT_FOUND, TOURNAMENT_NOT_SELECTED, TOURNAMENT_PARTICIPANT_NOT_FOUND or TOURNAMENT_STAGE_NOT_FOUND"
// @Failure 500 {object} apidocs.Error "INTERNAL_ERROR"
// @Security BearerAuth
// @Router /api/predictions/get_stage [post]
func (h *PredictionHandler) GetStage(c *gin.Context) {
	user, hasUser := middleware.CurrentUser(c)
	participant, hasParticipant := middleware.CurrentParticipant(c)
	if !hasUser || user.SelectedTournamentID == nil || !hasParticipant {
		response.AbortError(c, http.StatusInternalServerError, response.InternalError)
		return
	}
	var req predictionsdto.GetStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AbortError(c, http.StatusBadRequest, response.InvalidRequest)
		return
	}
	stage, err := h.service.GetStage(c.Request.Context(), *user.SelectedTournamentID, participant.ID, req.ID)
	if errors.Is(err, repository.ErrNotFound) {
		response.AbortError(c, http.StatusNotFound, response.TournamentStageNotFound)
		return
	}
	if err != nil {
		_ = c.Error(err)
		response.AbortError(c, http.StatusInternalServerError, response.InternalError)
		return
	}
	response.Success(c, http.StatusOK, stage)
}

// GetUserPredictions returns the stage's matches and the current participant's predictions.
// @Summary Get stage matches with my predictions
// @Tags Predictions
// @Description Uses the authenticated user's selected tournament and active participant. Matches are ordered by scheduled_start_at then id. A missing prediction is null. An empty stage returns data: []. Cache-Control: no-store.
// @Accept json
// @Param request body predictionsdto.GetStageRequest true "JSON request body"
// @Produce json
// @Success 200 {object} apidocs.Success{data=[]model.MatchWithPrediction}
// @Failure 400 {object} apidocs.Error "INVALID_REQUEST (invalid body or missing/invalid id)"
// @Failure 401 {object} apidocs.Error "UNAUTHORIZED"
// @Failure 404 {object} apidocs.Error "USER_NOT_FOUND, TOURNAMENT_NOT_SELECTED, TOURNAMENT_PARTICIPANT_NOT_FOUND or TOURNAMENT_STAGE_NOT_FOUND"
// @Failure 500 {object} apidocs.Error "INTERNAL_ERROR"
// @Security BearerAuth
// @Router /api/predictions/get_user_predictions [post]
func (h *PredictionHandler) GetUserPredictions(c *gin.Context) {
	user, hasUser := middleware.CurrentUser(c)
	participant, hasParticipant := middleware.CurrentParticipant(c)
	if !hasUser || user.SelectedTournamentID == nil || !hasParticipant {
		response.AbortError(c, http.StatusInternalServerError, response.InternalError)
		return
	}
	var req predictionsdto.GetStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AbortError(c, http.StatusBadRequest, response.InvalidRequest)
		return
	}
	matches, err := h.service.GetUserPredictions(c.Request.Context(), *user.SelectedTournamentID, participant.ID, req.ID)
	if errors.Is(err, repository.ErrNotFound) {
		response.AbortError(c, http.StatusNotFound, response.TournamentStageNotFound)
		return
	}
	if err != nil {
		_ = c.Error(err)
		response.AbortError(c, http.StatusInternalServerError, response.InternalError)
		return
	}
	response.Success(c, http.StatusOK, matches)
}

// SavePrediction creates or updates the active participant's score prediction.
// @Summary Save my prediction for one match
// @Tags Predictions
// @Description Uses the selected tournament and authenticated participant. The stage must have opened (a null opening time is closed), the match deadline must be in the future, and live or finished matches are closed. Calculated predictions are final. Returns success after an insert or update. Cache-Control: no-store.
// @Accept json
// @Param request body predictionsdto.SavePredictionRequest true "Match ID and predicted home and away scores (zero is valid)"
// @Produce json
// @Success 200 {object} apidocs.EmptySuccess
// @Failure 400 {object} apidocs.Error "INVALID_REQUEST (invalid body, missing, null or out-of-range field)"
// @Failure 401 {object} apidocs.Error "UNAUTHORIZED"
// @Failure 404 {object} apidocs.Error "USER_NOT_FOUND, TOURNAMENT_NOT_SELECTED, TOURNAMENT_PARTICIPANT_NOT_FOUND or MATCH_NOT_FOUND"
// @Failure 409 {object} apidocs.Error "PREDICTIONS_NOT_OPEN, PREDICTION_LOCKED or PREDICTION_ALREADY_CALCULATED"
// @Failure 500 {object} apidocs.Error "INTERNAL_ERROR"
// @Security BearerAuth
// @Router /api/predictions/save_prediction [post]
func (h *PredictionHandler) SavePrediction(c *gin.Context) {
	user, hasUser := middleware.CurrentUser(c)
	participant, hasParticipant := middleware.CurrentParticipant(c)
	if !hasUser || user.SelectedTournamentID == nil || !hasParticipant {
		response.AbortError(c, http.StatusInternalServerError, response.InternalError)
		return
	}
	var req predictionsdto.SavePredictionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AbortError(c, http.StatusBadRequest, response.InvalidRequest)
		return
	}
	err := h.service.SavePrediction(c.Request.Context(), *user.SelectedTournamentID, participant.ID, user.ID, req.MatchID, *req.HomeScore, *req.AwayScore)
	switch {
	case err == nil:
		response.Success(c, http.StatusOK, nil)
	case errors.Is(err, repository.ErrParticipantNotFound):
		response.AbortError(c, http.StatusNotFound, response.TournamentParticipantNotFound)
	case errors.Is(err, repository.ErrNotFound):
		response.AbortError(c, http.StatusNotFound, response.MatchNotFound)
	case errors.Is(err, repository.ErrPredictionsNotOpen):
		response.AbortError(c, http.StatusConflict, response.PredictionsNotOpen)
	case errors.Is(err, repository.ErrPredictionLocked):
		response.AbortError(c, http.StatusConflict, response.PredictionLocked)
	case errors.Is(err, repository.ErrPredictionCalculated):
		response.AbortError(c, http.StatusConflict, response.PredictionAlreadyCalculated)
	default:
		_ = c.Error(err)
		response.AbortError(c, http.StatusInternalServerError, response.InternalError)
	}
}
