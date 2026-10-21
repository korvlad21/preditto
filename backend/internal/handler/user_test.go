package handler_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"preditto/internal/auth"
	"preditto/internal/handler"
	"preditto/internal/middleware"
	"preditto/internal/repository"
	"preditto/internal/router"
	"preditto/internal/service"
)

func userAPI(t *testing.T, db *sql.DB) (*gin.Engine, *auth.TokenManager) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	tokens, err := auth.NewTokenManager(testConfig())
	if err != nil {
		t.Fatal(err)
	}
	api := router.New()
	users := service.NewUserService(repository.NewUserRepository(db))
	router.RegisterUserRoutes(api.Group("/api", middleware.RequireAuth(tokens), middleware.LoadUser(users)), handler.NewUserHandler())
	return api, tokens
}

func userRequest(api http.Handler, token, query, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/user/get_user_info"+query, strings.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	return w
}

func TestGetUserInfoRequiresAuth(t *testing.T) {
	api, tokens := userAPI(t, nil)
	pair, _, err := tokens.GeneratePair(81)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, token string }{
		{"missing token", ""},
		{"invalid token", "invalid"},
		{"refresh token", pair.RefreshToken},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := userRequest(api, tc.token, "", "")
			assertError(t, w, http.StatusUnauthorized, "UNAUTHORIZED")
			if w.Header().Get("WWW-Authenticate") != "Bearer" {
				t.Fatal("missing bearer challenge")
			}
		})
	}
	w := httptest.NewRecorder()
	api.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/user/get_user_info", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("GET status = %d, want 404", w.Code)
	}
}

func TestGetUserInfoDatabaseError(t *testing.T) {
	db, err := sql.Open("postgres", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	api, tokens := userAPI(t, db)
	pair, _, err := tokens.GeneratePair(81)
	if err != nil {
		t.Fatal(err)
	}
	w := userRequest(api, pair.AccessToken, "", "")
	assertError(t, w, http.StatusInternalServerError, "INTERNAL_ERROR")
	if strings.Contains(w.Body.String(), "database is closed") {
		t.Fatalf("database error leaked: %s", w.Body.String())
	}
}

func TestGetUserInfoPostgres(t *testing.T) {
	db := postgres(t)
	execSQL(t, db, `
		INSERT INTO users (id, username, email, password_hash, status) VALUES
			(81, 'new_user', 'new_user@example.com', 'private-password-hash', 'ACTIVE'),
			(82, 'other_user', 'other_user@example.com', 'other-password-hash', 'ACTIVE');
		INSERT INTO user_info (user_id, first_name, last_name, avatar_url, favorite_team_id)
		VALUES (81, 'Vladislav', 'Korobkin', '/avatars/81.png', 5);
	`)
	api, tokens := userAPI(t, db)
	pair, _, err := tokens.GeneratePair(81)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"id": float64(81), "username": "new_user", "email": "new_user@example.com", "status": "ACTIVE",
		"first_name": "Vladislav", "last_name": "Korobkin", "avatar_url": "/avatars/81.png", "favorite_team_id": float64(5), "selected_tournament_id": nil,
	}
	check := func(t *testing.T, token, query, body string, want map[string]any) {
		t.Helper()
		w := userRequest(api, token, query, body)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d: %s", w.Code, w.Body.String())
		}
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("user response can be cached")
		}
		var got map[string]any
		if err := json.Unmarshal(successData(t, w, ""), &got); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("user = %v, want %v", got, want)
		}
	}
	for _, tc := range []struct{ name, query, body string }{
		{"no parameters", "", ""},
		{"query cannot select another user", "?user_id=82&id=82", ""},
		{"body cannot select another user", "", `{"user_id":82,"id":82}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			check(t, pair.AccessToken, tc.query, tc.body, want)
		})
	}

	execSQL(t, db, `INSERT INTO tournaments (id, name, slug) VALUES (201, 'First', 'first'), (202, 'Second', 'second')`)
	for _, id := range []int64{201, 202} {
		if _, err := db.Exec("UPDATE user_info SET selected_tournament_id = $1 WHERE user_id = 81", id); err != nil {
			t.Fatal(err)
		}
		want["selected_tournament_id"] = float64(id)
		check(t, pair.AccessToken, "", "", want)
	}
	execSQL(t, db, `DELETE FROM tournaments WHERE id = 202`)
	want["selected_tournament_id"] = nil
	check(t, pair.AccessToken, "", "", want)

	execSQL(t, db, `UPDATE user_info SET first_name = NULL, last_name = NULL, avatar_url = NULL, favorite_team_id = NULL WHERE user_id = 81`)
	for _, field := range []string{"first_name", "last_name", "avatar_url", "favorite_team_id"} {
		want[field] = nil
	}
	t.Run("nullable profile", func(t *testing.T) {
		check(t, pair.AccessToken, "", "", want)
	})

	execSQL(t, db, `UPDATE user_info SET first_name = '', last_name = '', avatar_url = '' WHERE user_id = 81`)
	for _, field := range []string{"first_name", "last_name", "avatar_url"} {
		want[field] = ""
	}
	t.Run("empty strings are preserved", func(t *testing.T) {
		check(t, pair.AccessToken, "", "", want)
	})

	execSQL(t, db, `DELETE FROM user_info WHERE user_id = 81`)
	for _, field := range []string{"first_name", "last_name", "avatar_url"} {
		want[field] = nil
	}
	t.Run("missing profile", func(t *testing.T) {
		check(t, pair.AccessToken, "", "", want)
	})

	otherPair, _, err := tokens.GeneratePair(82)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("another authenticated user", func(t *testing.T) {
		check(t, otherPair.AccessToken, "?user_id=81", "", map[string]any{
			"id": float64(82), "username": "other_user", "email": "other_user@example.com", "status": "ACTIVE",
			"first_name": nil, "last_name": nil, "avatar_url": nil, "favorite_team_id": nil, "selected_tournament_id": nil,
		})
	})

	execSQL(t, db, `DELETE FROM users WHERE id = 81`)
	t.Run("deleted user with valid token", func(t *testing.T) {
		assertError(t, userRequest(api, pair.AccessToken, "", ""), http.StatusNotFound, "USER_NOT_FOUND")
	})
}
