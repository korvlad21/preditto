package router

import (
	"github.com/gin-gonic/gin"

	"preditto/internal/handler"
)

func RegisterTournamentRoutes(api *gin.RouterGroup, tournaments *handler.TournamentHandler) {
	routes := api.Group("/tournaments")
	routes.GET("/get_all_tournaments", tournaments.GetAllTournaments)
}
