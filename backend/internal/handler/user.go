package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"preditto/internal/middleware"
	"preditto/internal/response"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// GetUserInfo get the current user.
// @Summary Get the current user
// @Tags Users
// @Description No body or parameters required. Selectors are ignored. A missing profile leaves profile fields null. Cache-Control: no-store.
// @Produce json
// @Success 200 {object} apidocs.Success{data=model.UserInfo}
// @Failure 401 {object} apidocs.Error "UNAUTHORIZED (missing, invalid or expired access token)"
// @Failure 404 {object} apidocs.Error "USER_NOT_FOUND"
// @Failure 500 {object} apidocs.Error "INTERNAL_ERROR"
// @Security BearerAuth
// @Router /api/user/get_user_info [post]
func (h *UserHandler) GetUserInfo(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		response.AbortError(c, http.StatusInternalServerError, response.InternalError)
		return
	}
	response.Success(c, http.StatusOK, user)
}
