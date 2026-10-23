package middleware_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"preditto/internal/auth"
	"preditto/internal/config"
	"preditto/internal/handler"
	"preditto/internal/middleware"
	"preditto/internal/model"
	"preditto/internal/repository"
)

type userLoader func(context.Context, int64) (model.UserInfo, error)

func (f userLoader) GetUserInfo(ctx context.Context, id int64) (model.UserInfo, error) {
	return f(ctx, id)
}

func TestLoadUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens, err := auth.NewTokenManager(config.AuthConfig{AccessSecret: strings.Repeat("a", 32), RefreshSecret: strings.Repeat("r", 32), AccessTTL: time.Minute, RefreshTTL: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	pair, _, err := tokens.GeneratePair(81)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, token   string
		err           error
		status, calls int
	}{
		{"valid", pair.AccessToken, nil, 200, 1},
		{"missing token", "", nil, 401, 0},
		{"invalid token", "invalid", nil, 401, 0},
		{"missing user", pair.AccessToken, repository.ErrNotFound, 404, 1},
		{"database error", pair.AccessToken, errors.New("private database error"), 500, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls, downstream := 0, false
			r := gin.New()
			r.GET("/user", middleware.RequireAuth(tokens), middleware.LoadUser(userLoader(func(ctx context.Context, id int64) (model.UserInfo, error) {
				calls++
				if id != 81 {
					t.Fatalf("user ID = %d", id)
				}
				return model.UserInfo{ID: id, CreatedAt: time.Date(2026, 1, 2, 3, 4, 5, 123456000, time.UTC)}, tc.err
			})), func(c *gin.Context) {
				downstream = true
				user, ok := middleware.CurrentUser(c)
				if !ok || user.ID != 81 {
					t.Fatal("missing context user")
				}
				// Simulate a successful write: all consumers share the updated object.
				id := int64(202)
				user.SelectedTournamentID = &id
				again, ok := middleware.CurrentUser(c)
				if !ok || again != user {
					t.Fatal("context snapshot was copied")
				}
				c.Next()
			}, handler.NewUserHandler().GetUserInfo)
			req := httptest.NewRequest("GET", "/user", nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.status || calls != tc.calls || downstream != (tc.status == 200) {
				t.Fatalf("status=%d calls=%d downstream=%v", w.Code, calls, downstream)
			}
			if strings.Contains(w.Body.String(), "private database error") {
				t.Fatal("error leaked")
			}
			var body map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || len(body) != 2 || body["success"] != (tc.status == 200) {
				t.Fatalf("response envelope: %s", w.Body.String())
			}
			if tc.status != 200 {
				want := map[int]string{401: "UNAUTHORIZED", 404: "USER_NOT_FOUND", 500: "INTERNAL_ERROR"}[tc.status]
				if body["error"] != want {
					t.Fatalf("error = %v, want %s", body["error"], want)
				}
			}
			if tc.status == 200 && !strings.Contains(w.Body.String(), `"selected_tournament_id":202`) {
				t.Fatal("handler did not use updated context")
			}
			if tc.status == 200 && body["data"].(map[string]any)["created_at"] != "2026-01-02T03:04:05.123456Z" {
				t.Fatal("handler did not preserve account creation timestamp")
			}
		})
	}
}

func TestUserContextMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/user", handler.NewUserHandler().GetUserInfo)
	r.GET("/load", middleware.LoadUser(userLoader(func(context.Context, int64) (model.UserInfo, error) {
		t.Fatal("database called without authentication")
		return model.UserInfo{}, nil
	})))
	for path, status := range map[string]int{"/user": 500, "/load": 401} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != status {
			t.Fatalf("%s: status=%d", path, w.Code)
		}
	}
}
