package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"preditto/internal/response"
)

// RequireRole must run after RequireAuth and LoadUser; it reuses the request snapshot.
func RequireRole(code string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := CurrentUser(c)
		if !ok {
			unauthorized(c)
			return
		}
		for _, role := range user.Roles {
			if role == code {
				c.Next()
				return
			}
		}
		response.AbortError(c, http.StatusForbidden, response.Forbidden)
	}
}
