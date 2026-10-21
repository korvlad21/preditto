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
	response.Success(c, http.StatusCreated, result)
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
	response.Success(c, http.StatusOK, result)
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
	response.Success(c, http.StatusOK, result)
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
	response.Success(c, http.StatusOK, nil)
}

func authError(c *gin.Context, err error) {
	status, code := http.StatusInternalServerError, response.InternalError
	switch {
	case errors.Is(err, service.ErrInvalidRequest):
		status, code = http.StatusBadRequest, response.InvalidRequest
	case errors.Is(err, service.ErrInvalidCredentials):
		status, code = http.StatusUnauthorized, response.InvalidCredentials
	case errors.Is(err, auth.ErrInvalidToken):
		status, code = http.StatusUnauthorized, response.InvalidToken
	case errors.Is(err, service.ErrUserBlocked):
		status, code = http.StatusForbidden, response.UserBlocked
	case errors.Is(err, repository.ErrUsernameExists):
		status, code = http.StatusConflict, response.UsernameAlreadyExists
	case errors.Is(err, repository.ErrEmailExists):
		status, code = http.StatusConflict, response.EmailAlreadyExists
	case errors.Is(err, repository.ErrFavoriteTeamNotFound):
		status, code = http.StatusBadRequest, response.FavoriteTeamNotFound
	default:
		_ = c.Error(err)
	}
	response.AbortError(c, status, code)
}
