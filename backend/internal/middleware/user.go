package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"preditto/internal/model"
	"preditto/internal/repository"
	"preditto/internal/response"
)

const userInfoKey = "user_info"

type UserInfoLoader interface {
	GetUserInfo(context.Context, int64) (model.UserInfo, error)
}

// LoadUser must run after RequireAuth. It loads one snapshot per request.
func LoadUser(users UserInfoLoader) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		userID := c.GetInt64(UserIDKey)
		if userID <= 0 {
			unauthorized(c)
			return
		}
		user, err := users.GetUserInfo(c.Request.Context(), userID)
		if errors.Is(err, repository.ErrNotFound) {
			response.AbortError(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
			return
		}
		if err != nil {
			_ = c.Error(err)
			response.AbortError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
			return
		}
		c.Set(userInfoKey, &user)
		c.Next()
	}
}

// CurrentUser returns the request's shared user snapshot. After a successful
// database write, update this object if later handlers need the changed fields.
func CurrentUser(c *gin.Context) (*model.UserInfo, bool) {
	value, exists := c.Get(userInfoKey)
	user, ok := value.(*model.UserInfo)
	return user, exists && ok && user != nil
}
