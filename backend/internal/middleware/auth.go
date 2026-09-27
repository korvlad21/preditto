package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"preditto/internal/auth"
	"preditto/internal/response"
)

const UserIDKey = "user_id"

func RequireAuth(tokens *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		values := c.Request.Header.Values("Authorization")
		if len(values) != 1 {
			unauthorized(c)
			return
		}
		parts := strings.Fields(values[0])
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			unauthorized(c)
			return
		}
		claims, err := tokens.ParseAccessToken(parts[1])
		if err != nil {
			unauthorized(c)
			return
		}
		c.Set(UserIDKey, claims.UserID())
		c.Next()
	}
}

func unauthorized(c *gin.Context) {
	c.Header("WWW-Authenticate", "Bearer")
	response.AbortError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Valid access token required")
}
