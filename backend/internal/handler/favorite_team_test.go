package handler_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"preditto/internal/auth"
	"preditto/internal/handler"
	"preditto/internal/middleware"
	"preditto/internal/model"
	"preditto/internal/repository"
	"preditto/internal/router"
	"preditto/internal/service"
)

type favoriteTeamLoader struct {
	user  model.UserInfo
	calls int
}

func (l *favoriteTeamLoader) GetUserInfo(context.Context, int64) (model.UserInfo, error) {
	l.calls++
	return l.user, nil
}

func favoriteTeamAPI(t *testing.T, db *sql.DB, users middleware.UserInfoLoader) (*gin.Engine, *auth.TokenManager) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	tokens, err := auth.NewTokenManager(testConfig())
	if err != nil {
		t.Fatal(err)
	}
	api := router.New()
	teams := handler.NewTeamHandler(service.NewTeamService(repository.NewTeamRepository(db)))
	router.RegisterTeamRoutes(api.Group("/api"), teams, handler.NewCountryHandler(nil))
	router.RegisterProtectedTeamRoutes(api.Group("/api", middleware.RequireAuth(tokens), middleware.LoadUser(users)), teams)
	return api, tokens
}

func favoriteTeamRequest(api http.Handler, token, query, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/teams/get_favorite_team"+query, strings.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	return w
}

func assertNoFavoriteTeam(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	if w.Code != http.StatusOK || string(successData(t, w, "team")) != "null" {
		t.Fatalf("expected successful null team: %d %s", w.Code, w.Body.String())
	}
}

func TestGetFavoriteTeamAuthAndMissingFavorite(t *testing.T) {
	users := &favoriteTeamLoader{user: model.UserInfo{ID: 81}}
	api, tokens := favoriteTeamAPI(t, nil, users)
	pair, _, err := tokens.GeneratePair(81)
	if err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{"", "invalid", pair.RefreshToken} {
		assertError(t, favoriteTeamRequest(api, token, "", ""), http.StatusUnauthorized, "UNAUTHORIZED")
	}
	if users.calls != 0 {
		t.Fatal("unauthenticated request loaded user")
	}
	assertNoFavoriteTeam(t, favoriteTeamRequest(api, pair.AccessToken, "?id=5", `{"favorite_team_id":5}`))
	if users.calls != 1 {
		t.Fatalf("user loads = %d, want 1", users.calls)
	}
	w := httptest.NewRecorder()
	api.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/teams/get_favorite_team", nil))
	assertError(t, w, http.StatusNotFound, "NOT_FOUND")
}

func TestGetFavoriteTeamMissingContext(t *testing.T) {
	api := router.New()
	router.RegisterProtectedTeamRoutes(api.Group("/api"), handler.NewTeamHandler(nil))
	assertError(t, favoriteTeamRequest(api, "", "", ""), http.StatusInternalServerError, "INTERNAL_ERROR")
}

func TestGetFavoriteTeamPostgres(t *testing.T) {
	db := postgres(t)
	execSQL(t, db, `
		INSERT INTO users (id, username, email, password_hash, status) VALUES
			(81, 'favorite_user', 'favorite@example.com', 'hash', 'ACTIVE'),
			(82, 'other_user', 'other@example.com', 'hash', 'ACTIVE');
		INSERT INTO user_info (user_id, favorite_team_id) VALUES (81, 5);
	`)
	api, tokens := favoriteTeamAPI(t, db, service.NewUserService(repository.NewUserRepository(db)))
	pair, _, err := tokens.GeneratePair(81)
	if err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{"", `{}`, `{"id":999,"favorite_team_id":999,"user_id":82}`} {
		w := favoriteTeamRequest(api, pair.AccessToken, "?id=999&user_id=82", body)
		if w.Code != http.StatusOK || w.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("status = %d: %s", w.Code, w.Body.String())
		}
		want := getTeamRequest(api, "", `{"id":5}`)
		if string(successData(t, w, "team")) != string(successData(t, want, "team")) {
			t.Fatalf("favorite response differs from get_team: %s", w.Body.String())
		}
	}
	other, _, err := tokens.GeneratePair(82)
	if err != nil {
		t.Fatal(err)
	}
	assertNoFavoriteTeam(t, favoriteTeamRequest(api, other.AccessToken, "?user_id=81", ""))
	execSQL(t, db, `UPDATE user_info SET favorite_team_id = NULL WHERE user_id = 81`)
	assertNoFavoriteTeam(t, favoriteTeamRequest(api, pair.AccessToken, "", ""))

	// A team may disappear after LoadUser has captured its ID.
	id := int64(999)
	users := &favoriteTeamLoader{user: model.UserInfo{ID: 81, FavoriteTeamID: &id}}
	staleAPI, _ := favoriteTeamAPI(t, db, users)
	assertError(t, favoriteTeamRequest(staleAPI, pair.AccessToken, "", ""), http.StatusNotFound, "TEAM_NOT_FOUND")
	if users.calls != 1 {
		t.Fatalf("user loads = %d, want 1", users.calls)
	}
	execSQL(t, db, `DROP TABLE teams CASCADE`)
	assertError(t, favoriteTeamRequest(staleAPI, pair.AccessToken, "", ""), http.StatusInternalServerError, "INTERNAL_ERROR")
	execSQL(t, db, `DELETE FROM users WHERE id = 81`)
	assertError(t, favoriteTeamRequest(api, pair.AccessToken, "", ""), http.StatusNotFound, "USER_NOT_FOUND")
}
