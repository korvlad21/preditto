package router

import (
	"github.com/gin-gonic/gin"

	"preditto/internal/handler"
)

// RegisterTournamentRoutes expects a group with RequireAuth and LoadUser attached.
func RegisterTournamentRoutes(api *gin.RouterGroup, tournaments *handler.TournamentHandler) {
	routes := api.Group("/tournaments")
	routes.POST("/get_all_tournaments", tournaments.GetAllTournaments)
	routes.POST("/select_tournament", tournaments.SelectTournament)
}
