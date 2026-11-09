package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	usersdto "preditto/internal/dto/users"
	"preditto/internal/middleware"
	"preditto/internal/model"
	"preditto/internal/repository"
	"preditto/internal/response"
	"preditto/internal/service"
)

const maxUserJSONBytes = 16 << 10

// UpdateUser updates an account and its profile.
// @Summary Update a user
// @Tags Users
// @Description Owner or admin only, using CurrentUser from RequireAuth then LoadUser. JSON: user_id/email required; omitted/null optional fields are unchanged, favorite_team_id=0 clears the favorite, avatar_url="" clears the avatar, nonempty avatar_url must equal the target's stored URL. For upload use multipart/form-data with this JSON in the data field and one avatar file; omit avatar_url or set it empty. JPEG/PNG only, at most 5 MiB and 4096x4096 pixels, decoded and re-encoded as JPEG with a UUID name. JSON limit 16 KiB; multipart limit 5 MiB + 20 KiB. Atomic account/profile update; a missing profile is created. Other fields are preserved. Returns updated UserInfo. Cache-Control: no-store. Swagger 2 represents the JSON mode; the multipart alternative is described here and in docs/user.md.
// @Accept json,mpfd
// @Param request body usersdto.UpdateUserRequest true "JSON body, or JSON data field with an avatar file in multipart mode"
// @Produce json
// @Success 200 {object} apidocs.Success{data=model.UserInfo}
// @Failure 400 {object} apidocs.Error "INVALID_REQUEST (including invalid image or avatar path) or FAVORITE_TEAM_NOT_FOUND"
// @Failure 401 {object} apidocs.Error "UNAUTHORIZED"
// @Failure 403 {object} apidocs.Error "FORBIDDEN"
// @Failure 404 {object} apidocs.Error "USER_NOT_FOUND"
// @Failure 409 {object} apidocs.Error "EMAIL_ALREADY_EXISTS"
// @Failure 413 {object} apidocs.Error "INVALID_REQUEST (request or avatar exceeds size limit)"
// @Failure 415 {object} apidocs.Error "INVALID_REQUEST (unsupported request Content-Type)"
// @Failure 500 {object} apidocs.Error "INTERNAL_ERROR"
// @Security BearerAuth
// @Router /api/user/update_user [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	current, ok := middleware.CurrentUser(c)
	if !ok || current.ID <= 0 {
		c.Header("WWW-Authenticate", "Bearer")
		response.AbortError(c, http.StatusUnauthorized, response.Unauthorized)
		return
	}
	req, file, err := bindUserUpdate(c)
	if c.Request.MultipartForm != nil {
		defer c.Request.MultipartForm.RemoveAll()
	}
	if err != nil {
		updateUserError(c, err)
		return
	}
	if current.ID != req.UserID {
		admin := false
		for _, role := range current.Roles {
			if role == model.RoleCodeAdmin {
				admin = true
				break
			}
		}
		if !admin {
			response.AbortError(c, http.StatusForbidden, response.Forbidden)
			return
		}
	}
	var avatar io.Reader
	if file != nil {
		if file.Size > service.MaxAvatarBytes {
			updateUserError(c, service.ErrAvatarTooLarge)
			return
		}
		source, err := file.Open()
		if err != nil {
			updateUserError(c, err)
			return
		}
		defer source.Close()
		avatar = source
	}
	updated, err := h.service.UpdateUser(c.Request.Context(), req, avatar)
	if err != nil {
		updateUserError(c, err)
		return
	}
	if current.ID == updated.ID {
		*current = updated
	}
	response.Success(c, http.StatusOK, updated)
}

var errUserMediaType = errors.New("unsupported user update media type")

func bindUserUpdate(c *gin.Context) (usersdto.UpdateUserRequest, *multipart.FileHeader, error) {
	var req usersdto.UpdateUserRequest
	mediaType, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
	if err != nil {
		return req, nil, errUserMediaType
	}
	var source io.Reader
	var file *multipart.FileHeader
	switch mediaType {
	case "application/json":
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUserJSONBytes)
		source = c.Request.Body
	case "multipart/form-data":
		limit := int64(service.MaxAvatarBytes + maxUserJSONBytes + 4096)
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		if err := c.Request.ParseMultipartForm(limit); err != nil {
			var sizeErr *http.MaxBytesError
			if errors.As(err, &sizeErr) || errors.Is(err, multipart.ErrMessageTooLarge) {
				return req, nil, service.ErrAvatarTooLarge
			}
			return req, nil, service.ErrInvalidRequest
		}
		form := c.Request.MultipartForm
		if len(form.Value) != 1 || len(form.Value["data"]) != 1 || len(form.File) != 1 || len(form.File["avatar"]) != 1 {
			return req, nil, service.ErrInvalidRequest
		}
		if len(form.Value["data"][0]) > maxUserJSONBytes {
			return req, nil, service.ErrAvatarTooLarge
		}
		source, file = strings.NewReader(form.Value["data"][0]), form.File["avatar"][0]
	default:
		return req, nil, errUserMediaType
	}
	data, err := io.ReadAll(source)
	if err != nil {
		var sizeErr *http.MaxBytesError
		if errors.As(err, &sizeErr) {
			return req, nil, err
		}
		return req, nil, service.ErrInvalidRequest
	}
	if !utf8.Valid(data) {
		return req, nil, service.ErrInvalidRequest
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return req, nil, service.ErrInvalidRequest
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return req, nil, service.ErrInvalidRequest
	}
	if err := req.Validate(); err != nil {
		return req, nil, service.ErrInvalidRequest
	}
	return req, file, nil
}

func updateUserError(c *gin.Context, err error) {
	status, code := http.StatusBadRequest, response.InvalidRequest
	var sizeErr *http.MaxBytesError
	switch {
	case errors.Is(err, errUserMediaType):
		status = http.StatusUnsupportedMediaType
	case errors.Is(err, service.ErrAvatarTooLarge), errors.As(err, &sizeErr):
		status = http.StatusRequestEntityTooLarge
	case errors.Is(err, repository.ErrNotFound):
		status, code = http.StatusNotFound, response.UserNotFound
	case errors.Is(err, repository.ErrEmailExists):
		status, code = http.StatusConflict, response.EmailAlreadyExists
	case errors.Is(err, repository.ErrFavoriteTeamNotFound):
		code = response.FavoriteTeamNotFound
	case errors.Is(err, service.ErrInvalidRequest), errors.Is(err, repository.ErrInvalidAvatarURL):
	default:
		status, code = http.StatusInternalServerError, response.InternalError
		_ = c.Error(err)
	}
	response.AbortError(c, status, code)
}
