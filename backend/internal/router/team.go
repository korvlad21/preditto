package router

import (
	"github.com/gin-gonic/gin"

	"preditto/internal/handler"
)

func RegisterTeamRoutes(api *gin.RouterGroup, teams *handler.TeamHandler, countries *handler.CountryHandler) {
	routes := api.Group("/teams")
	routes.POST("/get_team", teams.GetTeam)
	routes.POST("/get_all_teams", teams.GetAllTeams)
	routes.POST("/get_all_countries", countries.GetAllCountries)
}

// RegisterProtectedTeamRoutes requires RequireAuth followed by LoadUser.
func RegisterProtectedTeamRoutes(api *gin.RouterGroup, teams *handler.TeamHandler) {
	api.POST("/teams/get_favorite_team", teams.GetFavoriteTeam)
}
