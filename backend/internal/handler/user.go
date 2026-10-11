package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"preditto/internal/middleware"
	"preditto/internal/repository"
	"preditto/internal/response"
	"preditto/internal/service"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) GetUserInfo(c *gin.Context) {
	userID := c.GetInt64(middleware.UserIDKey)
	user, err := h.service.GetUserInfo(c.Request.Context(), userID)
	if errors.Is(err, repository.ErrNotFound) {
		response.AbortError(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		return
	}
	if err != nil {
		_ = c.Error(err)
		response.AbortError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}
	c.JSON(http.StatusOK, user)
}
