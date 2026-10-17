package router

import (
	"github.com/gin-gonic/gin"

	"preditto/internal/handler"
)

// RegisterUserRoutes expects a group with RequireAuth and LoadUser attached.
func RegisterUserRoutes(api *gin.RouterGroup, user *handler.UserHandler) {
	routes := api.Group("/user")
	routes.POST("/get_user_info", user.GetUserInfo)
}
