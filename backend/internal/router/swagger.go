package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "preditto/docs/swagger"
)

// RegisterSwaggerRoutes exposes both the UI and specification only in development.
// Gin's default debug mode is deliberately not an environment signal.
func RegisterSwaggerRoutes(router *gin.Engine, environment string) {
	if environment != "development" {
		return
	}
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.URL("doc.json"),
		ginSwagger.PersistAuthorization(false),
	))
}
