package handler

import (
	"encoding/json"
	"errors"
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
		response.AbortError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутренняя ошибка сервера")
		return
	}
	var req tournamentsdto.SelectTournamentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		var typeErr *json.UnmarshalTypeError
		if errors.As(err, &typeErr) && typeErr.Field == "id" {
			response.AbortError(c, http.StatusBadRequest, "INVALID_TOURNAMENT_ID", "ID турнира должен быть положительным целым числом")
		} else {
			response.AbortError(c, http.StatusBadRequest, "INVALID_REQUEST", "Некорректный JSON запроса")
		}
		return
	}
	if req.ID == nil {
		response.AbortError(c, http.StatusBadRequest, "TOURNAMENT_ID_REQUIRED", "ID турнира обязателен и не может быть null")
		return
	}
	if *req.ID <= 0 {
		response.AbortError(c, http.StatusBadRequest, "INVALID_TOURNAMENT_ID", "ID турнира должен быть положительным целым числом")
		return
	}
	if err := h.service.SelectTournament(c.Request.Context(), user.ID, *req.ID); err != nil {
		switch {
		case errors.Is(err, repository.ErrTournamentNotFound):
			response.AbortError(c, http.StatusNotFound, "TOURNAMENT_NOT_FOUND", "Турнир не найден")
		case errors.Is(err, repository.ErrNotFound):
			response.AbortError(c, http.StatusNotFound, "USER_INFO_NOT_FOUND", "Профиль пользователя не найден")
		case errors.Is(err, service.ErrTournamentUpdateFailed):
			_ = c.Error(err)
			response.AbortError(c, http.StatusInternalServerError, "TOURNAMENT_UPDATE_FAILED", "Не удалось сохранить выбранный турнир")
		default:
			_ = c.Error(err)
			response.AbortError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутренняя ошибка сервера")
		}
		return
	}
	user.SelectedTournamentID = req.ID
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *TournamentHandler) GetAllTournaments(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		response.AbortError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}
	var req tournamentsdto.GetAllTournamentsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.AbortError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request fields")
		return
	}
	tournaments, err := h.service.GetAllTournaments(c.Request.Context(), req.Status, user.SelectedTournamentID)
	if err != nil {
		_ = c.Error(err)
		response.AbortError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}
	c.JSON(http.StatusOK, tournaments)
}
