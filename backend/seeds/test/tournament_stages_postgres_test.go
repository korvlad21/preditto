package test

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"preditto/seeds/development"
)

func TestTournamentStagesPostgres(t *testing.T) {
	db := seedPostgres(t)
	// Neither the tournament nor an existing stage has an assumed numeric ID.
	// A non-UTC database session must not shift the stored UTC wall-clock values.
	if _, err := db.Exec(`
		SET TIME ZONE 'Asia/Tbilisi';
		INSERT INTO tournaments (id, name, slug) VALUES
			(1, 'Other', 'other'), (81, 'Old', 'uefa-champions-league-2026-2027');
		ALTER TABLE tournaments ALTER COLUMN id RESTART WITH 101;
		INSERT INTO tournament_stages (id, tournament_id, code, name, stage_kind, sequence_no,
			status, outcome_points, exact_score_points, created_at) VALUES
			(91, 81, 'league-1', 'Old', 'round', 99, 'cancelled', 0, 0, '2000-01-01'),
			(1, 1, 'league-1', 'Other', 'round', 1, 'cancelled', 0, 0, '2000-01-01');
		ALTER TABLE tournament_stages ALTER COLUMN id RESTART WITH 101;
	`); err != nil {
		t.Fatal(err)
	}
	type snapshot struct {
		ID                       int64
		Code, Name, Kind, Status string
		Sequence, Outcome, Exact int
		Open, Close, Created     time.Time
	}
	var before []snapshot
	for run := 0; run < 2; run++ {
		var started time.Time
		if err := db.QueryRow(`SELECT CURRENT_TIMESTAMP AT TIME ZONE 'UTC'`).Scan(&started); err != nil {
			t.Fatal(err)
		}
		if err := development.Run(context.Background(), db); err != nil {
			t.Fatal(err)
		}
		rows, err := db.Query(`SELECT id, code, name, stage_kind, status, sequence_no,
			outcome_points, exact_score_points, predictions_open_at, predictions_close_at, created_at
			FROM tournament_stages WHERE tournament_id = 81 ORDER BY sequence_no`)
		if err != nil {
			t.Fatal(err)
		}
		var got []snapshot
		for rows.Next() {
			var s snapshot
			if err := rows.Scan(&s.ID, &s.Code, &s.Name, &s.Kind, &s.Status, &s.Sequence,
				&s.Outcome, &s.Exact, &s.Open, &s.Close, &s.Created); err != nil {
				rows.Close()
				t.Fatal(err)
			}
			got = append(got, s)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		if len(got) != 17 {
			t.Fatalf("got %d stages, want 17", len(got))
		}
		for i, s := range got {
			points, kind := 1, "group"
			if i >= 8 {
				points, kind = 2+(i-8)/2, "play-off"
			}
			if s.Sequence != i+1 || s.Kind != kind || s.Outcome != points || s.Exact != 3*points ||
				s.Name == "" || !s.Open.Before(s.Close) || s.Status == "cancelled" {
				t.Fatalf("invalid persisted stage: %+v", s)
			}
		}
		if got[0].ID != 91 || got[0].Created.Format("2006-01-02") != "2000-01-01" ||
			got[0].Open.Format("2006-01-02 15:04") != "2026-08-20 00:00" ||
			got[0].Close.Format("2006-01-02 15:04") != "2026-09-08 14:45" ||
			got[4].Open.Format("2006-01-02 15:04") != "2026-11-05 04:00" ||
			got[16].Close.Format("2006-01-02 15:04") != "2027-06-05 14:00" {
			t.Fatalf("incorrect identity or UTC timestamps: first=%+v fifth=%+v final=%+v", got[0], got[4], got[16])
		}
		var unchanged, utcTimestamps bool
		var count int
		if err := db.QueryRow(`SELECT
			(SELECT count(*) FROM tournament_stages),
			(SELECT name = 'Other' AND stage_kind = 'round' AND status = 'cancelled'
				FROM tournament_stages WHERE tournament_id = 1),
			(SELECT bool_and(updated_at BETWEEN $1::timestamp
				AND CURRENT_TIMESTAMP AT TIME ZONE 'UTC') FROM tournament_stages WHERE tournament_id = 81)
		`, started).Scan(&count, &unchanged, &utcTimestamps); err != nil {
			t.Fatal(err)
		}
		if count != 18 || !unchanged || !utcTimestamps {
			t.Fatalf("count=%d unrelated unchanged=%v UTC timestamps=%v", count, unchanged, utcTimestamps)
		}
		if run == 1 && !reflect.DeepEqual(before, got) {
			t.Fatal("rerun did not preserve identities/created_at or restore stage fields")
		}
		before = got
		if run == 0 {
			if _, err := db.Exec(`UPDATE tournament_stages SET name = 'Modified', stage_kind = 'round',
				sequence_no = 99, status = 'cancelled', predictions_open_at = NULL,
				predictions_close_at = NULL, outcome_points = 0, exact_score_points = 0,
				updated_at = '2000-01-01' WHERE tournament_id = 81`); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestTournamentStageFailureRollsBackPostgres(t *testing.T) {
	db := seedPostgres(t)
	if _, err := db.Exec(`ALTER TABLE tournament_stages ADD CONSTRAINT reject_final CHECK (sequence_no <> 17)`); err != nil {
		t.Fatal(err)
	}
	err := development.Run(context.Background(), db)
	if err == nil || !strings.Contains(err.Error(), "seed tournament_stages:") {
		t.Fatalf("expected stage write failure, got %v", err)
	}
	var count int
	if err := db.QueryRow(`SELECT (SELECT count(*) FROM countries) + (SELECT count(*) FROM teams)
		+ (SELECT count(*) FROM users) + (SELECT count(*) FROM user_info) + (SELECT count(*) FROM roles)
		+ (SELECT count(*) FROM user_roles) + (SELECT count(*) FROM tournaments)
		+ (SELECT count(*) FROM tournament_participants) + (SELECT count(*) FROM tournament_prize_rules)
		+ (SELECT count(*) FROM tournament_stages)`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("rollback left %d seeded rows", count)
	}
}
