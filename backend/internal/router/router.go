package router

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"preditto/internal/middleware"
	"preditto/internal/response"
)

func New() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), middleware.APIErrors(), gin.CustomRecovery(middleware.Recover))
	// Preserve 404 for unsupported methods, including OPTIONS and HEAD.
	router.HandleMethodNotAllowed = false
	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			response.AbortError(c, http.StatusNotFound, response.NotFound)
		}
	})
	return router
}
