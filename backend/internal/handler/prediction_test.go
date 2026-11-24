package handler_test

import (
	"database/sql"
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

func predictionRequest(t *testing.T, db *sql.DB) func(string, string) *httptest.ResponseRecorder {
	t.Helper()
	tokens, err := auth.NewTokenManager(testConfig())
	if err != nil {
		t.Fatal(err)
	}
	api := router.New()
	users := service.NewUserService(repository.NewUserRepository(db), nil)
	participants := service.NewTournamentParticipantService(repository.NewTournamentParticipantRepository(db))
	predictions := service.NewPredictionService(repository.NewTournamentStageRepository(db), repository.NewPredictionRepository(db))
	router.RegisterPredictionRoutes(api.Group("/api", middleware.RequireAuth(tokens), middleware.LoadUser(users)),
		handler.NewPredictionHandler(predictions), participants)
	pair, _, err := tokens.GeneratePair(81, nil)
	if err != nil {
		t.Fatal(err)
	}
	return func(path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/predictions/"+path, strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		return w
	}
}

func TestPredictionStagesPostgres(t *testing.T) {
	db := postgres(t)
	request := predictionRequest(t, db)
	assertError(t, request("get_stages", ""), 404, "USER_NOT_FOUND")
	assertError(t, request("get_stage", `{"id":21}`), 404, "USER_NOT_FOUND")
	assertError(t, request("get_user_predictions", `{"id":21}`), 404, "USER_NOT_FOUND")
	execSQL(t, db, `
		INSERT INTO users (id, username, email, password_hash, status) VALUES
			(81, 'first', 'first@example.com', 'unused', 'ACTIVE'),
			(82, 'other', 'other@example.com', 'unused', 'ACTIVE');
		INSERT INTO tournaments (id, name, slug) VALUES (1, 'First', 'first'), (2, 'Other', 'other');
		INSERT INTO user_info (user_id) VALUES (81);`)
	assertError(t, request("get_stages", ""), 404, "TOURNAMENT_NOT_SELECTED")
	assertError(t, request("get_stage", `{"id":21}`), 404, "TOURNAMENT_NOT_SELECTED")
	assertError(t, request("get_user_predictions", `{"id":21}`), 404, "TOURNAMENT_NOT_SELECTED")
	execSQL(t, db, "UPDATE user_info SET selected_tournament_id = 1 WHERE user_id = 81")
	assertError(t, request("get_stages", ""), 404, "TOURNAMENT_PARTICIPANT_NOT_FOUND")
	assertError(t, request("get_stage", `{"id":21}`), 404, "TOURNAMENT_PARTICIPANT_NOT_FOUND")
	execSQL(t, db, `INSERT INTO tournament_participants (id, tournament_id, user_id, status)
		VALUES (11, 1, 81, 'blocked'), (12, 1, 82, 'active')`)
	assertError(t, request("get_stages", ""), 404, "TOURNAMENT_PARTICIPANT_NOT_FOUND")
	assertError(t, request("get_stage", `{"id":21}`), 404, "TOURNAMENT_PARTICIPANT_NOT_FOUND")
	assertError(t, request("get_user_predictions", `{"id":21}`), 404, "TOURNAMENT_PARTICIPANT_NOT_FOUND")
	execSQL(t, db, "UPDATE tournament_participants SET status = 'active' WHERE id = 11")
	for _, body := range []string{"", `{}`, `{"id":null}`, `{"id":0}`, `{"id":-1}`, `{"id":"21"}`,
		`{"id":1.5}`, `{"id":9223372036854775808}`, `{"id":`} {
		t.Run("invalid stage request "+body, func(t *testing.T) {
			assertError(t, request("get_stage?id=21", body), 400, "INVALID_REQUEST")
			assertError(t, request("get_user_predictions?id=21", body), 400, "INVALID_REQUEST")
		})
	}
	checkData := func(t *testing.T, path, body, expected string) {
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
	checkData(t, "get_stages", "", "[]")
	assertError(t, request("get_stage", `{"id":21}`), 404, "TOURNAMENT_STAGE_NOT_FOUND")
	assertError(t, request("get_user_predictions", `{"id":21}`), 404, "TOURNAMENT_STAGE_NOT_FOUND")
	execSQL(t, db, `
		INSERT INTO tournament_stages (id, tournament_id, code, name, stage_kind, sequence_no, status,
			predictions_open_at, outcome_points, exact_score_points) VALUES
			(21, 1, 'final', 'Final', 'play-off', 2, 'planned', NULL, 2, 6),
			(22, 1, 'first', 'First', 'group', 1, 'finished', '2026-09-01 12:00:00.123456', 1, 3),
			(23, 2, 'other', 'Other', 'group', 1, 'planned', NULL, 1, 3);
		INSERT INTO teams (id, name, slug, country, short_name, logo_url)
			VALUES (6, 'Away', 'away', 'TST', 'AWY', '/away.png');
		INSERT INTO matches (id, stage_id, home_team_id, away_team_id,
			scheduled_start_at, prediction_lock_at) VALUES
			(34, 22, 5, 6, '2026-11-25 21:00', '2026-11-25 20:55'),
			(31, 22, 5, 6, '2026-11-25 21:00', '2026-11-25 20:55'),
			(32, 22, 6, 5, '2026-11-24 21:00', '2026-11-24 20:55'),
			(33, 23, 5, 6, now(), now());
		UPDATE matches SET home_score = 1, away_score = 0, status = 'finished',
			actual_start_at = '2026-11-25 21:01', finished_at = '2026-11-25 22:52' WHERE id = 31;
		INSERT INTO predictions (id, participant_id, match_id, home_score, away_score, result_type, awarded_points) VALUES
			(41, 11, 31, 1, 0, 'exact_score', 3), (42, 12, 31, 0, 0, 'missed', 0),
			(43, 12, 32, 1, 1, NULL, NULL), (44, 11, 33, 0, 1, NULL, NULL),
			(45, 11, 34, 0, 0, NULL, NULL);`)
	const firstStage = `{"id":22,"tournament_id":1,"code":"first","name":"First","stage_kind":"group","sequence_no":1,
		 "status":"finished","predictions_open_at":"2026-09-01T12:00:00.123456Z","outcome_points":1,
		 "exact_score_points":3,"count_matches":3,"count_predictions":2}`
	const finalStage = `{"id":21,"tournament_id":1,"code":"final","name":"Final","stage_kind":"play-off","sequence_no":2,
		 "status":"planned","predictions_open_at":null,"outcome_points":2,"exact_score_points":6,
		 "count_matches":0,"count_predictions":0}`
	checkData(t, "get_stages?tournament_id=2&participant_id=12", `{"tournament_id":2,"participant_id":12}`, "["+firstStage+","+finalStage+"]")
	checkData(t, "get_stage?id=21&tournament_id=2&participant_id=12", `{"id":22,"tournament_id":2,"participant_id":12}`, firstStage)
	checkData(t, "get_stage", `{"id":21}`, finalStage)
	assertError(t, request("get_stage", `{"id":23}`), 404, "TOURNAMENT_STAGE_NOT_FOUND")
	t.Run("user predictions", func(t *testing.T) {
		const firstTeam = `{"id":5,"name":"Test team","short_name":null,"logo_url":null}`
		const secondTeam = `{"id":6,"name":"Away","short_name":"AWY","logo_url":"/away.png"}`
		expected := `[
			{"match":{"id":32,"home_score":null,"away_score":null,"scheduled_start_at":"2026-11-24T21:00:00Z",
			          "prediction_lock_at":"2026-11-24T20:55:00Z","actual_start_at":null,"finished_at":null,"status":"scheduled"},
			 "home_team":` + secondTeam + `,"away_team":` + firstTeam + `,"prediction":null},
			{"match":{"id":31,"home_score":1,"away_score":0,"scheduled_start_at":"2026-11-25T21:00:00Z",
			          "prediction_lock_at":"2026-11-25T20:55:00Z","actual_start_at":"2026-11-25T21:01:00Z",
			          "finished_at":"2026-11-25T22:52:00Z","status":"finished"},
			 "home_team":` + firstTeam + `,"away_team":` + secondTeam + `,
			 "prediction":{"id":41,"home_score":1,"away_score":0,"result_type":"exact_score","awarded_points":3}},
			{"match":{"id":34,"home_score":null,"away_score":null,"scheduled_start_at":"2026-11-25T21:00:00Z",
			          "prediction_lock_at":"2026-11-25T20:55:00Z","actual_start_at":null,"finished_at":null,"status":"scheduled"},
			 "home_team":` + firstTeam + `,"away_team":` + secondTeam + `,
			 "prediction":{"id":45,"home_score":0,"away_score":0,"result_type":null,"awarded_points":null}}
		]`
		checkData(t, "get_user_predictions?id=23&participant_id=12", `{"id":22,"tournament_id":2,"participant_id":12}`, expected)
		checkData(t, "get_user_predictions", `{"id":21}`, "[]")
		assertError(t, request("get_user_predictions", `{"id":23,"tournament_id":2}`), 404, "TOURNAMENT_STAGE_NOT_FOUND")
	})
	execSQL(t, db, "DELETE FROM predictions WHERE participant_id = 11 AND match_id IN (31, 34)")
	checkData(t, "get_stage", `{"id":22}`, strings.Replace(firstStage, `"count_predictions":2`, `"count_predictions":0`, 1))
	execSQL(t, db, "ALTER TABLE predictions RENAME TO unavailable_predictions")
	assertError(t, request("get_stages", ""), 500, "INTERNAL_ERROR")
	assertError(t, request("get_stage", `{"id":22}`), 500, "INTERNAL_ERROR")
	assertError(t, request("get_user_predictions", `{"id":22}`), 500, "INTERNAL_ERROR")
}
