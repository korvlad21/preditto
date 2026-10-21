package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"preditto/internal/handler"
	"preditto/internal/middleware"
	"preditto/internal/model"
	"preditto/internal/repository"
	"preditto/internal/router"
)

func selectTournamentRequest(api http.Handler, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/tournaments/select_tournament", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	return w
}

type selectionUserLoader struct{}

func (selectionUserLoader) GetUserInfo(context.Context, int64) (model.UserInfo, error) {
	return model.UserInfo{ID: 81}, nil
}

func TestSelectTournamentValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := router.New()
	api.Use(func(c *gin.Context) { c.Set(middleware.UserIDKey, int64(81)) }, middleware.LoadUser(selectionUserLoader{}))
	// A nil service proves invalid bodies never reach persistence.
	router.RegisterTournamentRoutes(api.Group("/api"), handler.NewTournamentHandler(nil))
	for _, tc := range []struct{ name, body, code string }{
		{"missing", `{}`, "TOURNAMENT_ID_REQUIRED"},
		{"null id", `{"id":null}`, "TOURNAMENT_ID_REQUIRED"},
		{"null body", `null`, "TOURNAMENT_ID_REQUIRED"},
		{"zero", `{"id":0}`, "INVALID_TOURNAMENT_ID"},
		{"negative", `{"id":-1}`, "INVALID_TOURNAMENT_ID"},
		{"string", `{"id":"201"}`, "INVALID_TOURNAMENT_ID"},
		{"fraction", `{"id":1.5}`, "INVALID_TOURNAMENT_ID"},
		{"boolean", `{"id":true}`, "INVALID_TOURNAMENT_ID"},
		{"object", `{"id":{}}`, "INVALID_TOURNAMENT_ID"},
		{"array id", `{"id":[]}`, "INVALID_TOURNAMENT_ID"},
		{"overflow", `{"id":9223372036854775808}`, "INVALID_TOURNAMENT_ID"},
		{"array body", `[]`, "INVALID_REQUEST"},
		{"empty", ``, "INVALID_REQUEST"},
		{"malformed", `{"id":`, "INVALID_REQUEST"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertError(t, selectTournamentRequest(api, "", tc.body), http.StatusBadRequest, tc.code)
		})
	}
}

func TestSelectTournamentRequiresAuth(t *testing.T) {
	api, tokens, users := tournamentAPI(t, nil)
	pair, _, err := tokens.GeneratePair(81)
	if err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{"", "invalid", pair.RefreshToken} {
		assertError(t, selectTournamentRequest(api, token, `{"id":201}`), http.StatusUnauthorized, "UNAUTHORIZED")
	}
	if users.calls != 0 {
		t.Fatal("unauthenticated requests loaded user data")
	}
	w := httptest.NewRecorder()
	api.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/tournaments/select_tournament", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("GET status = %d", w.Code)
	}
}

func TestSelectTournamentMissingUserContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := router.New()
	router.RegisterTournamentRoutes(api.Group("/api"), handler.NewTournamentHandler(nil))
	assertError(t, selectTournamentRequest(api, "", `{"id":201}`), http.StatusInternalServerError, "INTERNAL_ERROR")
}

func TestSelectTournamentPostgres(t *testing.T) {
	db := postgres(t)
	execSQL(t, db, `
		INSERT INTO users (id, username, email, password_hash, status) VALUES
			(81, 'first', 'first@example.com', 'private', 'ACTIVE'),
			(82, 'second', 'second@example.com', 'private', 'ACTIVE');
		INSERT INTO tournaments (id, name, slug) VALUES (201, 'First', 'first'), (202, 'Second', 'second');
		INSERT INTO user_info (user_id, first_name, selected_tournament_id) VALUES (81, 'Name', NULL), (82, 'Other', 202);
	`)
	var before, after *model.UserInfo
	var oldID *int64
	api, tokens, users := tournamentAPI(t, db, func(c *gin.Context) {
		before, _ = middleware.CurrentUser(c)
		oldID = before.SelectedTournamentID
		c.Next()
		after, _ = middleware.CurrentUser(c)
	})
	pair, _, err := tokens.GeneratePair(81)
	if err != nil {
		t.Fatal(err)
	}
	repo := repository.NewAuthRepository(db)
	checkStored := func(want *int64) {
		t.Helper()
		user, err := repo.GetUserInfo(context.Background(), 81)
		if err != nil {
			t.Fatal(err)
		}
		if (user.SelectedTournamentID == nil) != (want == nil) || (want != nil && *user.SelectedTournamentID != *want) {
			t.Fatalf("stored selection = %v, want %v", user.SelectedTournamentID, want)
		}
		var otherID int64
		var name string
		if err := db.QueryRow(`SELECT selected_tournament_id, first_name FROM user_info WHERE user_id = 82`).Scan(&otherID, &name); err != nil || otherID != 202 || name != "Other" {
			t.Fatalf("another user's profile changed: %d %s %v", otherID, name, err)
		}
	}
	for _, tc := range []struct {
		body string
		id   int64
	}{
		{`{"id":201,"user_id":82}`, 201},
		{`{"id":202}`, 202},
		{`{"id":202}`, 202},
	} {
		calls := users.calls
		w := selectTournamentRequest(api, pair.AccessToken, tc.body)
		if w.Code != http.StatusOK || string(successData(t, w, "")) != "null" {
			t.Fatalf("selection: %d %s", w.Code, w.Body.String())
		}
		if before != after || after.SelectedTournamentID == nil || *after.SelectedTournamentID != tc.id || users.calls != calls+1 {
			t.Fatal("request snapshot was not updated in place with one user load")
		}
		if w.Header().Get("Cache-Control") != "no-store" || after.FirstName == nil || *after.FirstName != "Name" {
			t.Fatal("cache policy or unrelated profile field changed")
		}
		checkStored(&tc.id)
	}
	id := int64(202)
	assertError(t, selectTournamentRequest(api, pair.AccessToken, `{"id":999}`), http.StatusNotFound, "TOURNAMENT_NOT_FOUND")
	if after.SelectedTournamentID != oldID {
		t.Fatal("failed lookup changed request snapshot")
	}
	checkStored(&id)

	// Reuse the same access token to verify the next request sees the new selection.
	w := tournamentRequest(api, pair.AccessToken, "")
	var tournaments []model.Tournament
	if err := json.Unmarshal(successData(t, w, "tournaments"), &tournaments); err != nil || w.Code != http.StatusOK || len(tournaments) != 2 {
		t.Fatalf("list after selection: %d %s", w.Code, w.Body.String())
	}
	for _, tournament := range tournaments {
		if tournament.IsSelectedTournament != (tournament.ID == id) {
			t.Fatal("listing uses a stale selection")
		}
	}

	// The repository supports NULL, while the HTTP endpoint rejects it.
	if err := repo.UpdateSelectedTournament(context.Background(), 81, nil); err != nil {
		t.Fatal(err)
	}
	checkStored(nil)
	assertError(t, selectTournamentRequest(api, pair.AccessToken, `{"id":null}`), http.StatusBadRequest, "TOURNAMENT_ID_REQUIRED")
	checkStored(nil)
	missing := int64(999)
	if err := repo.UpdateSelectedTournament(context.Background(), 81, &missing); !errors.Is(err, repository.ErrTournamentNotFound) {
		t.Fatalf("foreign key error = %v", err)
	}
	checkStored(nil)

	// Force a write failure after the existence check succeeds.
	execSQL(t, db, `ALTER TABLE user_info ADD CONSTRAINT reject_selection CHECK (selected_tournament_id IS NULL OR user_id = 82)`)
	assertError(t, selectTournamentRequest(api, pair.AccessToken, `{"id":201}`), http.StatusInternalServerError, "TOURNAMENT_UPDATE_FAILED")
	if after.SelectedTournamentID != oldID {
		t.Fatal("failed write changed request snapshot")
	}
	checkStored(nil)
	execSQL(t, db, `ALTER TABLE user_info DROP CONSTRAINT reject_selection; DELETE FROM user_info WHERE user_id = 81`)
	assertError(t, selectTournamentRequest(api, pair.AccessToken, `{"id":201}`), http.StatusNotFound, "USER_INFO_NOT_FOUND")
	if err := repo.UpdateSelectedTournament(context.Background(), 81, nil); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("missing profile error = %v", err)
	}

	// Keep profile loading possible while making only the tournament lookup fail.
	execSQL(t, db, `ALTER TABLE tournaments RENAME TO unavailable_tournaments`)
	assertError(t, selectTournamentRequest(api, pair.AccessToken, `{"id":201}`), http.StatusInternalServerError, "INTERNAL_ERROR")
	execSQL(t, db, `DELETE FROM users WHERE id = 81`)
	assertError(t, selectTournamentRequest(api, pair.AccessToken, `{"id":201}`), http.StatusNotFound, "USER_NOT_FOUND")
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	assertError(t, selectTournamentRequest(api, pair.AccessToken, `{"id":201}`), http.StatusInternalServerError, "INTERNAL_ERROR")
}
