package router

import (
	"github.com/gin-gonic/gin"

	"preditto/internal/handler"
)

func RegisterTeamRoutes(api *gin.RouterGroup, teams *handler.TeamHandler) {
	routes := api.Group("/teams")
	routes.POST("/get_all_teams", teams.GetAllTeams)
}
