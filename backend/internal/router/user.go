package router

import (
	"github.com/gin-gonic/gin"

	"preditto/internal/auth"
	"preditto/internal/handler"
	"preditto/internal/middleware"
)

func RegisterUserRoutes(api *gin.RouterGroup, user *handler.UserHandler, tokens *auth.TokenManager) {
	routes := api.Group("/user", middleware.RequireAuth(tokens))
	routes.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Next()
	})
	routes.POST("/get_user_info", user.GetUserInfo)
}
