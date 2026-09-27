package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"preditto/internal/handler"
)

// RegisterAuthRoutes attaches public authentication endpoints to the API group.
func RegisterAuthRoutes(api *gin.RouterGroup, auth *handler.AuthHandler) {
	routes := api.Group("/auth")
	routes.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Header("Pragma", "no-cache")
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
		c.Next()
	})
	routes.POST("/register", auth.Register)
	routes.POST("/login", auth.Login)
	routes.POST("/refresh", auth.Refresh)
	routes.POST("/logout", auth.Logout)
}
