package router

import (
	"github.com/gin-gonic/gin"

	"preditto/internal/handler"
)

// RegisterPredictionRoutes expects RequireAuth and LoadUser on the API group.
func RegisterPredictionRoutes(api *gin.RouterGroup, predictions *handler.PredictionHandler, loadParticipant gin.HandlerFunc) {
	routes := api.Group("/predictions")
	routes.POST("/get_stages", loadParticipant, predictions.GetStages)
}
