package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	tournamentsdto "preditto/internal/dto/tournaments"
	"preditto/internal/middleware"
	"preditto/internal/response"
	"preditto/internal/service"
)

type TournamentHandler struct {
	service *service.TournamentService
}

func NewTournamentHandler(service *service.TournamentService) *TournamentHandler {
	return &TournamentHandler{service: service}
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
