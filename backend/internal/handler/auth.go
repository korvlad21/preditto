package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"preditto/internal/auth"
	authdto "preditto/internal/dto/auth"
	"preditto/internal/repository"
	"preditto/internal/response"
	"preditto/internal/service"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(service *service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req authdto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		authError(c, service.ErrInvalidRequest)
		return
	}
	if err := req.Validate(); err != nil {
		authError(c, service.ErrInvalidRequest)
		return
	}
	result, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		authError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req authdto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		authError(c, service.ErrInvalidRequest)
		return
	}
	result, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		authError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req authdto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		authError(c, service.ErrInvalidRequest)
		return
	}
	result, err := h.service.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		authError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req authdto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		authError(c, service.ErrInvalidRequest)
		return
	}
	if err := h.service.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		authError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func authError(c *gin.Context, err error) {
	status, code, message := http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error"
	switch {
	case errors.Is(err, service.ErrInvalidRequest):
		status, code, message = http.StatusBadRequest, "INVALID_REQUEST", "Invalid request fields"
	case errors.Is(err, service.ErrInvalidCredentials):
		status, code, message = http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid login or password"
	case errors.Is(err, auth.ErrInvalidToken):
		status, code, message = http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired token"
	case errors.Is(err, service.ErrUserBlocked):
		status, code, message = http.StatusForbidden, "USER_BLOCKED", "User is blocked"
	case errors.Is(err, repository.ErrUsernameExists):
		status, code, message = http.StatusConflict, "USERNAME_ALREADY_EXISTS", "Username is already taken"
	case errors.Is(err, repository.ErrEmailExists):
		status, code, message = http.StatusConflict, "EMAIL_ALREADY_EXISTS", "Email is already taken"
	case errors.Is(err, repository.ErrFavoriteTeamNotFound):
		status, code, message = http.StatusBadRequest, "FAVORITE_TEAM_NOT_FOUND", "Favorite team does not exist"
	default:
		_ = c.Error(err)
	}
	response.AbortError(c, status, code, message)
}
