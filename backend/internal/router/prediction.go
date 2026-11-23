package router

import (
	"github.com/gin-gonic/gin"

	"preditto/internal/handler"
	"preditto/internal/middleware"
)

// RegisterPredictionRoutes expects RequireAuth and LoadUser on the API group.
func RegisterPredictionRoutes(api *gin.RouterGroup, predictions *handler.PredictionHandler, participants middleware.ParticipantInfoLoader) {
	routes := api.Group("/predictions", middleware.LoadParticipant(participants))
	routes.POST("/get_stages", predictions.GetStages)
	routes.POST("/get_stage", predictions.GetStage)
	routes.POST("/get_user_predictions", predictions.GetUserPredictions)
}
