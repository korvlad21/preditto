package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	teamsdto "preditto/internal/dto/teams"
	"preditto/internal/middleware"
	"preditto/internal/repository"
	"preditto/internal/response"
	"preditto/internal/service"
)

type TeamHandler struct {
	service *service.TeamService
}

func NewTeamHandler(service *service.TeamService) *TeamHandler {
	return &TeamHandler{service: service}
}

func (h *TeamHandler) GetTeam(c *gin.Context) {
	var req teamsdto.GetTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AbortError(c, http.StatusBadRequest, response.InvalidRequest)
		return
	}
	h.getTeam(c, req.ID)
}

func (h *TeamHandler) GetFavoriteTeam(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		response.AbortError(c, http.StatusInternalServerError, response.InternalError)
		return
	}
	if user.FavoriteTeamID == nil {
		response.AbortError(c, http.StatusNotFound, response.FavoriteTeamNotFound)
		return
	}
	h.getTeam(c, *user.FavoriteTeamID)
}

func (h *TeamHandler) getTeam(c *gin.Context, id int64) {
	team, err := h.service.GetTeam(c.Request.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		response.AbortError(c, http.StatusNotFound, response.TeamNotFound)
		return
	}
	if err != nil {
		_ = c.Error(err)
		response.AbortError(c, http.StatusInternalServerError, response.InternalError)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"team": team})
}

func (h *TeamHandler) GetAllTeams(c *gin.Context) {
	var req teamsdto.GetAllTeamsRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		response.AbortError(c, http.StatusBadRequest, response.InvalidRequest)
		return
	}
	teams, err := h.service.GetAllTeams(c.Request.Context(), req.Country)
	if err != nil {
		_ = c.Error(err)
		response.AbortError(c, http.StatusInternalServerError, response.InternalError)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"teams": teams})
}
