package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"preditto/internal/handler"
	"preditto/internal/middleware"
	"preditto/internal/model"
)

// RegisterAuthRoutes protects registration; registrationAuth must run RequireAuth then LoadUser.
// Login, refresh and logout remain public and validate their own credentials.
func RegisterAuthRoutes(api *gin.RouterGroup, auth *handler.AuthHandler, registrationAuth []gin.HandlerFunc) {
	routes := api.Group("/auth")
	routes.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Header("Pragma", "no-cache")
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
		c.Next()
	})
	registration := routes.Group("", registrationAuth...)
	registration.POST("/register", middleware.RequireRole(model.RoleCodeAdmin), auth.Register)
	routes.POST("/login", auth.Login)
	routes.POST("/refresh", auth.Refresh)
	routes.POST("/logout", auth.Logout)
}
