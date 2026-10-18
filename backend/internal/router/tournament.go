package router

import (
	"github.com/gin-gonic/gin"

	"preditto/internal/handler"
	"preditto/internal/response"
)

// Pass authentication middleware separately so selection errors share one envelope.
func RegisterTournamentRoutes(api *gin.RouterGroup, tournaments *handler.TournamentHandler, auth ...gin.HandlerFunc) {
	routes := api.Group("/tournaments")
	routes.Group("", auth...).POST("/get_all_tournaments", tournaments.GetAllTournaments)
	selection := routes.Group("", response.WithFailureEnvelope())
	selection.Use(auth...)
	selection.POST("/select_tournament", tournaments.SelectTournament)
}
