package handler

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"preditto/internal/auth"
	authdto "preditto/internal/dto/auth"
	"preditto/internal/mailer"
	"preditto/internal/repository"
	"preditto/internal/response"
	"preditto/internal/service"
)

// EmailSender is the email transport used for login notifications.
type EmailSender interface {
	Send(context.Context, mailer.Message) error
}

type AuthHandler struct {
	service *service.AuthService
	mailer  EmailSender
}

func NewAuthHandler(service *service.AuthService, sender EmailSender) *AuthHandler {
	return &AuthHandler{service: service, mailer: sender}
}

// Register register a user.
// @Summary Register a user
// @Tags Auth
// @Description Body limit: 16 KiB. Responses have Cache-Control: no-store and Pragma: no-cache. Always issues a short refresh session. Unknown fields, including remember_me, are ignored.
// @Accept json
// @Param request body authdto.RegisterRequest true "JSON request body"
// @Produce json
// @Success 201 {object} apidocs.Success{data=authdto.AuthResponse}
// @Failure 400 {object} apidocs.Error "INVALID_REQUEST or FAVORITE_TEAM_NOT_FOUND"
// @Failure 409 {object} apidocs.Error "USERNAME_ALREADY_EXISTS or EMAIL_ALREADY_EXISTS"
// @Failure 500 {object} apidocs.Error "INTERNAL_ERROR"
// @Router /api/auth/register [post]
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

// Login log in with username or email.
// @Summary Log in with username or email
// @Tags Auth
// @Description Body limit: 16 KiB. Responses have Cache-Control: no-store and Pragma: no-cache. Case-sensitive login. remember_me=false selects short refresh lifetime; true selects long lifetime; omitted/null retains long legacy lifetime. Successful login attempts an email notification when mail is enabled; delivery errors do not change the successful response.
// @Accept json
// @Param request body authdto.LoginRequest true "JSON request body"
// @Produce json
// @Success 200 {object} apidocs.Success{data=authdto.AuthResponse}
// @Failure 400 {object} apidocs.Error "INVALID_REQUEST"
// @Failure 401 {object} apidocs.Error "INVALID_CREDENTIALS"
// @Failure 403 {object} apidocs.Error "USER_BLOCKED"
// @Failure 500 {object} apidocs.Error "INTERNAL_ERROR"
// @Router /api/auth/login [post]
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

	if err := h.mailer.Send(c.Request.Context(), mailer.Message{
		To:      result.User.Email,
		Subject: "Вход в Preditto",
		Text:    "Вы вошли в свой аккаунт Preditto.",
	}); err != nil && !errors.Is(err, mailer.ErrDisabled) {
		log.Printf("login notification failed: %v", err)
	}
	response.Success(c, http.StatusOK, result)
}

// Refresh rotate a refresh token.
// @Summary Rotate a refresh token
// @Tags Auth
// @Description Body limit: 16 KiB. Responses have Cache-Control: no-store and Pragma: no-cache. Single-use rotation. Replace both tokens after success. The signed session mode is preserved; body fields cannot override it.
// @Accept json
// @Param request body authdto.RefreshRequest true "JSON request body"
// @Produce json
// @Success 200 {object} apidocs.Success{data=authdto.TokenResponse}
// @Failure 400 {object} apidocs.Error "INVALID_REQUEST"
// @Failure 401 {object} apidocs.Error "INVALID_TOKEN"
// @Failure 403 {object} apidocs.Error "USER_BLOCKED"
// @Failure 500 {object} apidocs.Error "INTERNAL_ERROR"
// @Router /api/auth/refresh [post]
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

// Logout revoke a refresh session.
// @Summary Revoke a refresh session
// @Tags Auth
// @Description Body limit: 16 KiB. Responses have Cache-Control: no-store and Pragma: no-cache. Repeated logout of an unexpired revoked session succeeds. Blocked users may log out. Access tokens remain valid until expiry. Success omits data.
// @Accept json
// @Param request body authdto.RefreshRequest true "JSON request body"
// @Produce json
// @Success 200 {object} apidocs.EmptySuccess
// @Failure 400 {object} apidocs.Error "INVALID_REQUEST"
// @Failure 401 {object} apidocs.Error "INVALID_TOKEN"
// @Failure 500 {object} apidocs.Error "INTERNAL_ERROR"
// @Router /api/auth/logout [post]
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
