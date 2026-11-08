package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"preditto/internal/model"
)

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name   string
		user   *model.UserInfo
		status int
		body   string
	}{
		{"missing snapshot", nil, 401, `{"success":false,"error":"UNAUTHORIZED"}`},
		{"no roles", &model.UserInfo{ID: 81}, 403, `{"success":false,"error":"FORBIDDEN"}`},
		{"participant", &model.UserInfo{ID: 81, Roles: []string{model.RoleCodeParticipant}}, 403, `{"success":false,"error":"FORBIDDEN"}`},
		{"exact role code required", &model.UserInfo{ID: 81, Roles: []string{"Admin", "administrator"}}, 403, `{"success":false,"error":"FORBIDDEN"}`},
		{"admin", &model.UserInfo{ID: 81, Roles: []string{model.RoleCodeAdmin}}, 204, ""},
		{"multiple roles", &model.UserInfo{ID: 81, Roles: []string{model.RoleCodeParticipant, model.RoleCodeAdmin}}, 204, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			r := gin.New()
			r.POST("/register", func(c *gin.Context) {
				if tc.user != nil {
					c.Set(userInfoKey, tc.user)
				}
			}, RequireRole(model.RoleCodeAdmin), func(c *gin.Context) {
				called = true
				c.Status(http.StatusNoContent)
			})
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/register", nil))
			if w.Code != tc.status || w.Body.String() != tc.body || called != (tc.status == http.StatusNoContent) {
				t.Fatalf("status=%d body=%s downstream=%v", w.Code, w.Body.String(), called)
			}
		})
	}
}
