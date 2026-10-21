package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"preditto/internal/response"
)

// APIErrors handles errors left without a response. It never rewrites a body.
func APIErrors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if strings.HasPrefix(c.Request.URL.Path, "/api/") && !c.Writer.Written() &&
			(len(c.Errors) > 0 || c.Writer.Status() >= http.StatusInternalServerError) {
			response.AbortError(c, http.StatusInternalServerError, response.InternalError)
		}
	}
}

// Recover is used with Gin's recovery middleware, which retains panic diagnostics.
func Recover(c *gin.Context, _ any) {
	if c.Writer.Written() {
		c.Abort()
		return
	}
	if strings.HasPrefix(c.Request.URL.Path, "/api/") {
		response.AbortError(c, http.StatusInternalServerError, response.InternalError)
		return
	}
	c.AbortWithStatus(http.StatusInternalServerError)
}
