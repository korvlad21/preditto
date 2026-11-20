package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"preditto/internal/auth"
	"preditto/internal/handler"
	"preditto/internal/middleware"
	"preditto/internal/repository"
	"preditto/internal/router"
	"preditto/internal/service"
)

func TestGetStagesPostgres(t *testing.T) {
	db := postgres(t)
	tokens, err := auth.NewTokenManager(testConfig())
	if err != nil {
		t.Fatal(err)
	}
	api := router.New()
	users := service.NewUserService(repository.NewUserRepository(db), nil)
	participants := service.NewTournamentParticipantService(repository.NewTournamentParticipantRepository(db))
	predictions := service.NewPredictionService(repository.NewTournamentStageRepository(db))
	router.RegisterPredictionRoutes(api.Group("/api", middleware.RequireAuth(tokens), middleware.LoadUser(users)),
		handler.NewPredictionHandler(predictions), middleware.LoadParticipant(participants))
	pair, _, err := tokens.GeneratePair(81, nil)
	if err != nil {
		t.Fatal(err)
	}
	request := func(path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/api/predictions/get_stages"+path, strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		return w
	}
	assertError(t, request("", ""), 404, "USER_NOT_FOUND")
	execSQL(t, db, `
		INSERT INTO users (id, username, email, password_hash, status) VALUES
			(81, 'first', 'first@example.com', 'unused', 'ACTIVE'),
			(82, 'other', 'other@example.com', 'unused', 'ACTIVE');
		INSERT INTO tournaments (id, name, slug) VALUES (1, 'First', 'first'), (2, 'Other', 'other');
		INSERT INTO user_info (user_id) VALUES (81);`)
	assertError(t, request("", ""), 404, "TOURNAMENT_NOT_SELECTED")
	execSQL(t, db, "UPDATE user_info SET selected_tournament_id = 1 WHERE user_id = 81")
	assertError(t, request("", ""), 404, "TOURNAMENT_PARTICIPANT_NOT_FOUND")
	execSQL(t, db, `INSERT INTO tournament_participants (id, tournament_id, user_id, status)
		VALUES (11, 1, 81, 'blocked'), (12, 1, 82, 'active')`)
	assertError(t, request("", ""), 404, "TOURNAMENT_PARTICIPANT_NOT_FOUND")
	execSQL(t, db, "UPDATE tournament_participants SET status = 'active' WHERE id = 11")
	checkStages := func(path, body, expected string) {
		t.Helper()
		w := request(path, body)
		var got, want any
		if w.Code != 200 {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
		if err := json.Unmarshal(successData(t, w, ""), &got); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal([]byte(expected), &want); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("data = %#v, want %#v", got, want)
		}
	}
	checkStages("", "", "[]")
	execSQL(t, db, `
		INSERT INTO tournament_stages (id, tournament_id, code, name, stage_kind, sequence_no, status,
			predictions_open_at, outcome_points, exact_score_points) VALUES
			(21, 1, 'final', 'Final', 'play-off', 2, 'planned', NULL, 2, 6),
			(22, 1, 'first', 'First', 'group', 1, 'finished', '2026-09-01 12:00:00.123456', 1, 3),
			(23, 2, 'other', 'Other', 'group', 1, 'planned', NULL, 1, 3);
		INSERT INTO teams (id, name, slug, country) VALUES (6, 'Away', 'away', 'TST');
		INSERT INTO matches (id, stage_id, home_tournament_team_id, away_tournament_team_id,
			scheduled_start_at, prediction_lock_at) VALUES
			(31, 22, 5, 6, now(), now()), (32, 22, 6, 5, now(), now()), (33, 23, 5, 6, now(), now());
		INSERT INTO predictions (participant_id, match_id, home_score, away_score) VALUES
			(11, 31, 1, 0), (12, 31, 0, 0), (12, 32, 1, 1), (11, 33, 0, 1);`)
	checkStages("?tournament_id=2&participant_id=12", `{"tournament_id":2,"participant_id":12}`, `[
		{"id":22,"tournament_id":1,"code":"first","name":"First","stage_kind":"group","sequence_no":1,
		 "status":"finished","predictions_open_at":"2026-09-01T12:00:00.123456Z","outcome_points":1,
		 "exact_score_points":3,"count_matches":2,"count_predictions":1},
		{"id":21,"tournament_id":1,"code":"final","name":"Final","stage_kind":"play-off","sequence_no":2,
		 "status":"planned","predictions_open_at":null,"outcome_points":2,"exact_score_points":6,
		 "count_matches":0,"count_predictions":0}]`)
	execSQL(t, db, "ALTER TABLE predictions RENAME TO unavailable_predictions")
	assertError(t, request("", ""), 500, "INTERNAL_ERROR")
}
