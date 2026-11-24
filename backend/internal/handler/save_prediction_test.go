package handler_test

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSavePredictionPostgres(t *testing.T) {
	db := postgres(t)
	request := predictionRequest(t, db)
	save := func(body string) *httptest.ResponseRecorder {
		return request("save_prediction", body)
	}
	const draw = `{"match_id":31,"home_score":0,"away_score":0}`
	assertError(t, save(draw), 404, "USER_NOT_FOUND")
	execSQL(t, db, `
		INSERT INTO users (id, username, email, password_hash, status)
		VALUES (81, 'first', 'first@example.com', 'unused', 'ACTIVE');
		INSERT INTO tournaments (id, name, slug) VALUES (1, 'First', 'first'), (2, 'Other', 'other');
		INSERT INTO user_info (user_id) VALUES (81);
		INSERT INTO tournament_stages (id, tournament_id, code, name, stage_kind, sequence_no, status,
			predictions_open_at, outcome_points, exact_score_points)
		SELECT 20+id, id, 'final', 'Final', 'play-off', 1, 'active', timestamp '2000-01-01', 1, 3 FROM tournaments;
		INSERT INTO teams (id, name, slug, country) VALUES (6, 'Away', 'away', 'TST');
		INSERT INTO matches (id, stage_id, home_team_id, away_team_id, scheduled_start_at, prediction_lock_at)
		VALUES (31, 21, 5, 6, now(), '2100-01-01'), (32, 22, 5, 6, now(), '2100-01-01')`)
	assertError(t, save(draw), 404, "TOURNAMENT_NOT_SELECTED")
	execSQL(t, db, "UPDATE user_info SET selected_tournament_id = 1")
	assertError(t, save(draw), 404, "TOURNAMENT_PARTICIPANT_NOT_FOUND")
	execSQL(t, db, `INSERT INTO tournament_participants (id, tournament_id, user_id, status)
		VALUES (11, 1, 81, 'not_paid'), (12, 2, 81, 'active')`)
	assertError(t, save(draw), 404, "TOURNAMENT_PARTICIPANT_NOT_FOUND")
	execSQL(t, db, "UPDATE tournament_participants SET status = 'active' WHERE id = 11")

	for _, body := range []string{"", "{", "{}", `{"home_score":0,"away_score":0}`,
		`{"match_id":0,"home_score":0,"away_score":0}`, `{"match_id":-1,"home_score":0,"away_score":0}`,
		`{"match_id":1.5,"home_score":0,"away_score":0}`, `{"match_id":9223372036854775808,"home_score":0,"away_score":0}`,
		`{"match_id":31,"home_score":0}`, `{"match_id":31,"away_score":0}`,
		`{"match_id":31,"home_score":null,"away_score":0}`, `{"match_id":31,"home_score":0,"away_score":null}`,
		`{"match_id":31,"home_score":-1,"away_score":0}`, `{"match_id":31,"home_score":0,"away_score":32768}`,
		`{"match_id":31,"home_score":"1","away_score":0}`, `{"match_id":31,"home_score":0,"away_score":1.5}`} {
		assertError(t, save(body), 400, "INVALID_REQUEST")
	}
	assertError(t, save(`{"match_id":99,"home_score":0,"away_score":0}`), 404, "MATCH_NOT_FOUND")
	assertError(t, save(`{"match_id":32,"home_score":0,"away_score":0,"participant_id":12,"tournament_id":2}`), 404, "MATCH_NOT_FOUND")
	for _, opening := range []string{"NULL", "'2100-01-01'"} {
		execSQL(t, db, "UPDATE tournament_stages SET predictions_open_at = "+opening+" WHERE id = 21")
		assertError(t, save(draw), 409, "PREDICTIONS_NOT_OPEN")
	}
	execSQL(t, db, "UPDATE tournament_stages SET predictions_open_at = clock_timestamp() AT TIME ZONE 'UTC' WHERE id = 21")
	for _, status := range []string{"live", "finished"} {
		execSQL(t, db, "UPDATE matches SET status = $1 WHERE id = 31", status)
		assertError(t, save(draw), 409, "PREDICTION_LOCKED")
	}
	execSQL(t, db, "UPDATE matches SET status = 'scheduled', prediction_lock_at = clock_timestamp() AT TIME ZONE 'UTC' WHERE id = 31")
	assertError(t, save(draw), 409, "PREDICTION_LOCKED")
	assertCount := func(query string, want int) {
		t.Helper()
		var n int
		if err := db.QueryRow("SELECT count(*) FROM predictions " + query).Scan(&n); err != nil || n != want {
			t.Fatalf("count %s = %d, want %d, err=%v", query, n, want, err)
		}
	}
	assertCount("", 0)
	execSQL(t, db, "UPDATE matches SET prediction_lock_at = '2100-01-01' WHERE id = 31")
	assertSaved := func(t *testing.T, w *httptest.ResponseRecorder) {
		t.Helper()
		if w.Code != 200 || w.Body.String() != `{"success":true}` {
			t.Fatalf("save: %d %s", w.Code, w.Body.String())
		}
	}
	assertSaved(t, save(draw))
	assertCount("WHERE participant_id=11 AND match_id=31 AND home_score=0 AND away_score=0 AND result_type IS NULL AND awarded_points IS NULL AND calculated_at IS NULL", 1)
	execSQL(t, db, "UPDATE predictions SET updated_at='2000-01-01', result_type='missed', awarded_points=0")
	for _, status := range []string{"postponed", "cancelled"} {
		execSQL(t, db, "UPDATE matches SET status=$1 WHERE id=31", status)
		assertSaved(t, save(`{"match_id":31,"home_score":32767,"away_score":2}`))
	}
	assertCount("WHERE home_score=32767 AND away_score=2 AND updated_at>'2000-01-01' AND result_type='missed' AND awarded_points=0 AND calculated_at IS NULL", 1)
	execSQL(t, db, "UPDATE matches SET status='scheduled' WHERE id=31")
	execSQL(t, db, "DELETE FROM predictions")

	t.Run("parallel first submissions", func(t *testing.T) {
		const n = 12
		start := make(chan struct{})
		results := make(chan *httptest.ResponseRecorder, n)
		for i := 1; i <= n; i++ {
			go func(score int) {
				<-start
				results <- save(fmt.Sprintf(`{"match_id":31,"home_score":%d,"away_score":%d}`, score, n-score))
			}(i)
		}
		close(start)
		for i := 0; i < n; i++ {
			assertSaved(t, <-results)
		}
	})
	assertCount("", 1)
	assertCount("WHERE participant_id=11 AND match_id=31 AND home_score BETWEEN 1 AND 12 AND home_score+away_score=12", 1)

	// Hold a real database lock and wait until the save is blocked on our transaction.
	for _, tc := range []struct {
		name, lockSQL, code string
		status              int
	}{
		{"deadline during update", "SELECT id FROM predictions FOR UPDATE", "PREDICTION_LOCKED", 409},
		{"concurrent calculation", "UPDATE predictions SET calculated_at=now(), result_type='exact_score', awarded_points=3", "PREDICTION_ALREADY_CALCULATED", 409},
		{"concurrent revocation", "UPDATE tournament_participants SET status='blocked' WHERE id=11", "TOURNAMENT_PARTICIPANT_NOT_FOUND", 404},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.code == "PREDICTION_LOCKED" {
				execSQL(t, db, "UPDATE matches SET prediction_lock_at = (clock_timestamp() AT TIME ZONE 'UTC') + interval '3 seconds' WHERE id=31")
			} else {
				execSQL(t, db, "UPDATE matches SET prediction_lock_at = '2100-01-01' WHERE id=31")
			}
			tx, err := db.Begin()
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()
			if _, err := tx.Exec(tc.lockSQL); err != nil {
				t.Fatal(err)
			}
			var pid int
			if err := tx.QueryRow("SELECT pg_backend_pid()").Scan(&pid); err != nil {
				t.Fatal(err)
			}
			done := make(chan *httptest.ResponseRecorder, 1)
			go func() { done <- save(draw) }()
			until := time.Now().Add(5 * time.Second)
			for {
				var waiting bool
				if err := db.QueryRow("SELECT EXISTS (SELECT FROM pg_stat_activity WHERE $1 = ANY(pg_blocking_pids(pid)))", pid).Scan(&waiting); err != nil {
					t.Fatal(err)
				}
				if waiting {
					break
				}
				if time.Now().After(until) {
					t.Fatal("save did not wait for lock")
				}
				time.Sleep(10 * time.Millisecond)
			}
			if tc.code == "PREDICTION_LOCKED" {
				execSQL(t, db, `SELECT pg_sleep(GREATEST(0, extract(epoch FROM prediction_lock_at - (clock_timestamp() AT TIME ZONE 'UTC')))) FROM matches WHERE id=31`)
			}
			if err := tx.Commit(); err != nil {
				t.Fatal(err)
			}
			assertError(t, <-done, tc.status, tc.code)
			assertError(t, save(draw), tc.status, tc.code)
		})
	}
	assertCount("WHERE home_score+away_score=12 AND calculated_at IS NOT NULL AND result_type='exact_score' AND awarded_points=3", 1)
}
