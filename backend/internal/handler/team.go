package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	teamsdto "preditto/internal/dto/teams"
	"preditto/internal/response"
	"preditto/internal/service"
)

type TeamHandler struct {
	service *service.TeamService
}

func NewTeamHandler(service *service.TeamService) *TeamHandler {
	return &TeamHandler{service: service}
}

func (h *TeamHandler) GetAllTeams(c *gin.Context) {
	var req teamsdto.GetAllTeamsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.AbortError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request fields")
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		response.AbortError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request fields")
		return
	}
	teams, err := h.service.GetAllTeams(c.Request.Context(), req.Country)
	if err != nil {
		_ = c.Error(err)
		response.AbortError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}
	c.JSON(http.StatusOK, teams)
}
