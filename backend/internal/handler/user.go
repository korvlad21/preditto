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

func (h *UserHandler) GetUserInfo(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		response.AbortError(c, http.StatusInternalServerError, response.InternalError)
		return
	}
	response.Success(c, http.StatusOK, user)
}