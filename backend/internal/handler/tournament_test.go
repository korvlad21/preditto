package handler_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
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

type tournamentUserLoader struct {
	middleware.UserInfoLoader
	calls int
}

func (l *tournamentUserLoader) GetUserInfo(ctx context.Context, userID int64) (model.UserInfo, error) {
	l.calls++
	return l.UserInfoLoader.GetUserInfo(ctx, userID)
}

func tournamentAPI(t *testing.T, db *sql.DB, observers ...gin.HandlerFunc) (*gin.Engine, *auth.TokenManager, *tournamentUserLoader) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	tokens, err := auth.NewTokenManager(testConfig())
	if err != nil {
		t.Fatal(err)
	}
	api := gin.New()
	users := &tournamentUserLoader{UserInfoLoader: service.NewUserService(repository.NewAuthRepository(db))}
	authMiddleware := append([]gin.HandlerFunc{middleware.RequireAuth(tokens), middleware.LoadUser(users)}, observers...)
	router.RegisterTournamentRoutes(api.Group("/api"), handler.NewTournamentHandler(service.NewTournamentService(repository.NewTournamentRepository(db), repository.NewAuthRepository(db))), authMiddleware...)
	return api, tokens, users
}

func tournamentRequest(api http.Handler, token, query string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/tournaments/get_all_tournaments"+query, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)
	return w
}

func TestGetAllTournamentsRequiresAuth(t *testing.T) {
	api, tokens, users := tournamentAPI(t, nil)
	pair, _, err := tokens.GeneratePair(81)
	if err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{"", "invalid", pair.RefreshToken} {
		w := tournamentRequest(api, token, "")
		assertError(t, w, http.StatusUnauthorized, "UNAUTHORIZED")
		if w.Header().Get("WWW-Authenticate") != "Bearer" {
			t.Fatal("missing bearer challenge")
		}
	}
	if users.calls != 0 {
		t.Fatal("unauthenticated requests loaded user data")
	}
	w := httptest.NewRecorder()
	api.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/tournaments/get_all_tournaments", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("GET status = %d, want 404", w.Code)
	}
}

func TestGetAllTournamentsMissingUserContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := gin.New()
	router.RegisterTournamentRoutes(api.Group("/api"), handler.NewTournamentHandler(nil))
	assertError(t, tournamentRequest(api, "", ""), http.StatusInternalServerError, "INTERNAL_ERROR")
}

func TestGetAllTournamentsPostgres(t *testing.T) {
	db := postgres(t)
	execSQL(t, db, `
		INSERT INTO users (id, username, email, password_hash, status) VALUES
			(81, 'first_user', 'first@example.com', 'private-hash', 'ACTIVE'),
			(82, 'other_user', 'other@example.com', 'private-hash', 'ACTIVE'),
			(83, 'third_user', 'third@example.com', 'private-hash', 'ACTIVE'),
			(84, 'fourth_user', 'fourth@example.com', 'private-hash', 'ACTIVE');
		INSERT INTO tournaments (id, name, slug, type_of_tournament, status, starts_at, ends_at, cost_amount, amount_currency) VALUES
			(201, 'A active', 'active', 'Champions league', 'active', '2026-09-08 12:34:56.123456', '2027-06-05', 10.25, 'USD'),
			(202, 'B planned', 'planned', NULL, 'planned', NULL, NULL, 0, 'USD'),
			(203, 'C finished', 'finished', NULL, 'finished', NULL, NULL, NULL, NULL),
			(204, 'D cancelled', 'cancelled', NULL, 'cancelled', NULL, NULL, NULL, NULL),
			(205, 'E nullable', 'nullable', NULL, NULL, NULL, NULL, NULL, NULL);
		INSERT INTO user_info (user_id, selected_tournament_id) VALUES (81, 201);
		INSERT INTO tournament_prize_rules (tournament_id, place, percentage) VALUES
			(201, 1, 50.25), (201, 2, 30.50), (201, 3, 14.25), (201, 4, 5),
			(202, 2, 0);
		INSERT INTO tournament_participants (tournament_id, user_id, status) VALUES
			(201, 81, 'active'), (201, 82, 'active'), (201, 83, 'not_paid'), (201, 84, 'blocked'),
			(202, 81, 'blocked'), (203, 82, 'active');
	`)
	api, tokens, users := tournamentAPI(t, db)
	pair, _, err := tokens.GeneratePair(81)
	if err != nil {
		t.Fatal(err)
	}
	all := []map[string]any{}
	for i, name := range []string{"A active", "B planned", "C finished", "D cancelled", "E nullable"} {
		all = append(all, map[string]any{
			"id": float64(201 + i), "name": name, "type_of_tournament": nil, "status": nil,
			"starts_at": nil, "ends_at": nil, "cost_amount": nil, "amount_currency": nil,
			"prize_first_place": nil, "prize_second_place": nil, "prize_third_place": nil,
			"count_participant": float64(0), "is_selected_tournament": false,
		})
	}
	for i, status := range []string{"active", "planned", "finished", "cancelled"} {
		all[i]["status"] = status
	}
	all[0]["type_of_tournament"] = "Champions league"
	all[0]["starts_at"] = "2026-09-08T12:34:56.123456Z"
	all[0]["ends_at"] = "2027-06-05T00:00:00Z"
	all[0]["cost_amount"] = 10.25
	all[0]["amount_currency"] = "USD"
	all[0]["prize_first_place"] = 50.25
	all[0]["prize_second_place"] = 30.50
	all[0]["prize_third_place"] = 14.25
	all[0]["count_participant"] = float64(2)
	all[0]["is_selected_tournament"] = true
	all[1]["cost_amount"] = float64(0)
	all[1]["amount_currency"] = "USD"
	all[1]["prize_second_place"] = float64(0)
	all[2]["count_participant"] = float64(1)

	check := func(t *testing.T, token, query string, want []map[string]any) {
		t.Helper()
		before := users.calls
		w := tournamentRequest(api, token, query)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d: %s", w.Code, w.Body.String())
		}
		if users.calls != before+1 {
			t.Fatalf("user loads = %d, want 1", users.calls-before)
		}
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("personalized response can be cached")
		}
		var got []map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("tournaments = %v, want %v", got, want)
		}
	}
	empty := []map[string]any{}
	for _, tc := range []struct {
		name, query string
		want        []map[string]any
	}{
		{"omitted status", "", all},
		{"empty status", "?status=", all},
		{"all", "?status=all", all},
		{"active", "?status=active", all[:1]},
		{"planned", "?status=planned", all[1:2]},
		{"finished", "?status=finished", all[2:3]},
		{"cancelled", "?status=cancelled", all[3:4]},
		{"unknown status", "?status=unknown", empty},
		{"case sensitive", "?status=ALL", empty},
		{"whitespace preserved", "?status=" + url.QueryEscape(" active "), empty},
		{"SQL injection", "?status=" + url.QueryEscape("active' OR 1=1 --"), empty},
		{"selection cannot be overridden", "?user_id=82&selected_tournament_id=202", all},
	} {
		t.Run(tc.name, func(t *testing.T) { check(t, pair.AccessToken, tc.query, tc.want) })
	}

	// Relax uniqueness only in this disposable schema to exercise DISTINCT.
	execSQL(t, db, `ALTER TABLE tournament_participants DROP CONSTRAINT uq_tournament_participants_tournament_user;
		INSERT INTO tournament_participants (tournament_id, user_id, status) VALUES (201, 81, 'active')`)
	check(t, pair.AccessToken, "", all)

	execSQL(t, db, `UPDATE user_info SET selected_tournament_id = 202 WHERE user_id = 81`)
	all[0]["is_selected_tournament"] = false
	all[1]["is_selected_tournament"] = true
	check(t, pair.AccessToken, "", all)
	check(t, pair.AccessToken, "?status=active", all[:1])
	execSQL(t, db, `UPDATE user_info SET selected_tournament_id = NULL WHERE user_id = 81`)
	all[1]["is_selected_tournament"] = false
	check(t, pair.AccessToken, "", all)

	execSQL(t, db, `UPDATE user_info SET selected_tournament_id = 201 WHERE user_id = 81`)
	otherPair, _, err := tokens.GeneratePair(82)
	if err != nil {
		t.Fatal(err)
	}
	check(t, otherPair.AccessToken, "?user_id=81", all)

	execSQL(t, db, `DELETE FROM tournament_participants; DELETE FROM tournament_prize_rules; DELETE FROM tournaments`)
	check(t, pair.AccessToken, "", empty)
	execSQL(t, db, `DROP TABLE tournament_prize_rules`)
	assertError(t, tournamentRequest(api, pair.AccessToken, ""), http.StatusInternalServerError, "INTERNAL_ERROR")
	execSQL(t, db, `DELETE FROM users WHERE id = 81`)
	assertError(t, tournamentRequest(api, pair.AccessToken, ""), http.StatusNotFound, "USER_NOT_FOUND")
}
