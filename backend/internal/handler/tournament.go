package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	tournamentsdto "preditto/internal/dto/tournaments"
	"preditto/internal/middleware"
	"preditto/internal/repository"
	"preditto/internal/response"
	"preditto/internal/service"
)

type TournamentHandler struct {
	service *service.TournamentService
}

func NewTournamentHandler(service *service.TournamentService) *TournamentHandler {
	return &TournamentHandler{service: service}
}

// SelectTournament select the current user tournament.
// @Summary Select the current user tournament
// @Tags Tournaments
// @Description Requires an existing profile and tournament, without membership or status restrictions. Repeated selection succeeds. Success omits data. Cache-Control: no-store.
// @Accept json
// @Param request body tournamentsdto.SelectTournamentRequest true "JSON request body"
// @Produce json
// @Success 200 {object} apidocs.EmptySuccess
// @Failure 401 {object} apidocs.Error "UNAUTHORIZED (missing, invalid or expired access token)"
// @Failure 400 {object} apidocs.Error "TOURNAMENT_ID_REQUIRED (missing/null id), INVALID_TOURNAMENT_ID (nonpositive, wrong type or overflow), INVALID_REQUEST (invalid body)"
// @Failure 404 {object} apidocs.Error "USER_NOT_FOUND, USER_INFO_NOT_FOUND or TOURNAMENT_NOT_FOUND"
// @Failure 500 {object} apidocs.Error "TOURNAMENT_UPDATE_FAILED or INTERNAL_ERROR"
// @Security BearerAuth
// @Router /api/tournaments/select_tournament [post]
func (h *TournamentHandler) SelectTournament(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		response.AbortError(c, http.StatusInternalServerError, response.InternalError)
		return
	}
	var req tournamentsdto.SelectTournamentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		var typeErr *json.UnmarshalTypeError
		if errors.As(err, &typeErr) && typeErr.Field == "id" {
			response.AbortError(c, http.StatusBadRequest, response.InvalidTournamentID)
		} else {
			response.AbortError(c, http.StatusBadRequest, response.InvalidRequest)
		}
		return
	}
	if req.ID == nil {
		response.AbortError(c, http.StatusBadRequest, response.TournamentIDRequired)
		return
	}
	if *req.ID <= 0 {
		response.AbortError(c, http.StatusBadRequest, response.InvalidTournamentID)
		return
	}
	if err := h.service.SelectTournament(c.Request.Context(), user.ID, *req.ID); err != nil {
		switch {
		case errors.Is(err, repository.ErrTournamentNotFound):
			response.AbortError(c, http.StatusNotFound, response.TournamentNotFound)
		case errors.Is(err, repository.ErrNotFound):
			response.AbortError(c, http.StatusNotFound, response.UserInfoNotFound)
		case errors.Is(err, service.ErrTournamentUpdateFailed):
			_ = c.Error(err)
			response.AbortError(c, http.StatusInternalServerError, response.TournamentUpdateFailed)
		default:
			_ = c.Error(err)
			response.AbortError(c, http.StatusInternalServerError, response.InternalError)
		}
		return
	}
	user.SelectedTournamentID = req.ID
	response.Success(c, http.StatusOK, nil)
}

// GetAllTournaments list tournaments.
// @Summary List tournaments
// @Tags Tournaments
// @Description Empty body or {} applies no filters. Omitted/empty status or lowercase all disables status filtering; other values match exactly (unknown values return []). is_participate considers only active participation; null/omission disables that filter. Filters combine with AND. Query parameters and user selectors are ignored. Ordered by created_at DESC then id DESC. Empty results are []. Cache-Control: no-store.
// @Accept json
// @Param request body tournamentsdto.GetAllTournamentsRequest false "JSON request body"
// @Produce json
// @Success 200 {object} apidocs.Success{data=apidocs.TournamentsData}
// @Failure 400 {object} apidocs.Error "INVALID_REQUEST"
// @Failure 401 {object} apidocs.Error "UNAUTHORIZED (missing, invalid or expired access token)"
// @Failure 404 {object} apidocs.Error "USER_NOT_FOUND"
// @Failure 500 {object} apidocs.Error "INTERNAL_ERROR"
// @Security BearerAuth
// @Router /api/tournaments/get_all_tournaments [post]
func (h *TournamentHandler) GetAllTournaments(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		response.AbortError(c, http.StatusInternalServerError, response.InternalError)
		return
	}
	var req tournamentsdto.GetAllTournamentsRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		response.AbortError(c, http.StatusBadRequest, response.InvalidRequest)
		return
	}
	tournaments, err := h.service.GetAllTournaments(c.Request.Context(), user.ID, req.Status, req.IsParticipate, user.SelectedTournamentID)
	if err != nil {
		_ = c.Error(err)
		response.AbortError(c, http.StatusInternalServerError, response.InternalError)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"tournaments": tournaments})
}
