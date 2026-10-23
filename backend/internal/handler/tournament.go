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
