package router

import (
	"github.com/gin-gonic/gin"

	"preditto/internal/handler"
	"preditto/internal/middleware"
)

// RegisterPredictionRoutes expects RequireAuth and LoadUser on the API group.
func RegisterPredictionRoutes(api *gin.RouterGroup, predictions *handler.PredictionHandler, participants middleware.ParticipantInfoLoader) {
	routes := api.Group("/predictions")
	routes.POST("/get_stages", middleware.LoadParticipant(participants), predictions.GetStages)
	routes.POST("/get_stage", middleware.LoadParticipant(participants), predictions.GetStage)
}
