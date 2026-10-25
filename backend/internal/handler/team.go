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

// GetTeam get a team by id.
// @Summary Get a team by ID
// @Tags Teams
// @Description ID is read only from JSON; query parameters are ignored.
// @Accept json
// @Param request body teamsdto.GetTeamRequest true "JSON request body"
// @Produce json
// @Success 200 {object} apidocs.Success{data=apidocs.TeamData}
// @Failure 400 {object} apidocs.Error "INVALID_REQUEST"
// @Failure 404 {object} apidocs.Error "TEAM_NOT_FOUND"
// @Failure 500 {object} apidocs.Error "INTERNAL_ERROR"
// @Router /api/teams/get_team [post]
func (h *TeamHandler) GetTeam(c *gin.Context) {
	var req teamsdto.GetTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.AbortError(c, http.StatusBadRequest, response.InvalidRequest)
		return
	}
	h.getTeam(c, req.ID)
}

// GetFavoriteTeam get the current user favorite team.
// @Summary Get the current user favorite team
// @Tags Teams
// @Description No body or parameters required. Selectors are ignored. data.team is null when the profile or favorite is absent. Cache-Control: no-store.
// @Produce json
// @Success 200 {object} apidocs.Success{data=apidocs.FavoriteTeamData}
// @Failure 401 {object} apidocs.Error "UNAUTHORIZED (missing, invalid or expired access token)"
// @Failure 404 {object} apidocs.Error "USER_NOT_FOUND or TEAM_NOT_FOUND"
// @Failure 500 {object} apidocs.Error "INTERNAL_ERROR"
// @Security BearerAuth
// @Router /api/teams/get_favorite_team [post]
func (h *TeamHandler) GetFavoriteTeam(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		response.AbortError(c, http.StatusInternalServerError, response.InternalError)
		return
	}
	if user.FavoriteTeamID == nil {
		response.Success(c, http.StatusOK, gin.H{"team": nil})
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

// GetAllTeams list teams.
// @Summary List teams
// @Tags Teams
// @Description Empty body or {} lists all teams. Omitted/empty country or uppercase ALL disables filtering. Other values match exactly. Query parameters are ignored. Ordered by name; empty results are [].
// @Accept json
// @Param request body teamsdto.GetAllTeamsRequest false "JSON request body"
// @Produce json
// @Success 200 {object} apidocs.Success{data=apidocs.TeamsData}
// @Failure 400 {object} apidocs.Error "INVALID_REQUEST"
// @Failure 500 {object} apidocs.Error "INTERNAL_ERROR"
// @Router /api/teams/get_all_teams [post]
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
