package test

import (
	"context"
	"strings"
	"testing"

	"preditto/seeds/development"
)

func TestLeaguePhaseMatchesPostgres(t *testing.T) {
	db := seedPostgres(t)
	if _, err := db.Exec(`SET TIME ZONE 'Asia/Tbilisi'`); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := development.Run(ctx, db); err != nil {
		t.Fatal(err)
	}
	var rounds int
	err := db.QueryRow(`
		SELECT count(*) FROM (
			SELECT s.code FROM matches m JOIN tournament_stages s ON s.id = m.stage_id
			JOIN tournaments t ON t.id = s.tournament_id
			WHERE t.slug = 'uefa-champions-league-2026-2027' AND s.sequence_no BETWEEN 1 AND 8
			GROUP BY s.code, s.sequence_no HAVING count(*) = 18 AND bool_and(
				m.prediction_lock_at = m.scheduled_start_at - interval '1 hour' AND
				CASE WHEN s.sequence_no <= 5 THEN m.status = 'finished'
					AND m.home_score IS NOT NULL AND m.away_score IS NOT NULL
					AND m.actual_start_at IS NOT NULL AND m.finished_at IS NOT NULL
					AND m.finished_at > m.actual_start_at
				ELSE m.status = 'scheduled' AND m.home_score IS NULL AND m.away_score IS NULL
					AND m.actual_start_at IS NULL AND m.finished_at IS NULL END)
		) valid_rounds
	`).Scan(&rounds)
	if err != nil || rounds != 8 {
		t.Fatalf("valid matchdays = %d, want 8: %v", rounds, err)
	}
	var appearances int
	err = db.QueryRow(`SELECT count(*) FROM (
		SELECT stage_id, club FROM matches
		CROSS JOIN LATERAL (VALUES (home_team_id), (away_team_id)) sides(club)
		GROUP BY stage_id, club HAVING count(*) = 1
	) once_per_round`).Scan(&appearances)
	if err != nil || appearances != 8*36 {
		t.Fatalf("unique club appearances = %d, want 288: %v", appearances, err)
	}
	// Preserve a later live score and an identically paired fixture in the GEL tournament.
	// Corrupt a finished fixture after the snapshot to exercise in-place correction.
	_, err = db.Exec(`
		UPDATE matches SET status = 'live', home_score = 1, away_score = 0
		WHERE status = 'scheduled' AND id = (SELECT min(id) FROM matches WHERE status = 'scheduled');
		WITH other_stage AS (
			INSERT INTO tournament_stages (tournament_id, code, name, stage_kind, sequence_no,
				status, outcome_points, exact_score_points)
			SELECT id, 'league-1', 'Other stage', 'group', 1, 'planned', 1, 3
			FROM tournaments WHERE slug = 'uefa-champions-league-2026-2027-gel' RETURNING id
		)
		INSERT INTO matches (stage_id, home_team_id, away_team_id,
			scheduled_start_at, prediction_lock_at, status)
		SELECT other_stage.id, h.id, a.id, '2030-01-01', '2029-12-31 23:00', 'postponed'
		FROM other_stage, teams h, teams a WHERE h.slug = 'aek-athens' AND a.slug = 'lask';
		CREATE TEMP TABLE before_rerun AS SELECT * FROM matches;
		UPDATE matches SET home_score = NULL, away_score = NULL, status = 'scheduled',
			scheduled_start_at = '2000-01-01' WHERE home_team_id =
			(SELECT id FROM teams WHERE slug = 'aek-athens') AND away_team_id =
			(SELECT id FROM teams WHERE slug = 'lask') AND status = 'finished';
	`)
	if err != nil {
		t.Fatal(err)
	}
	if err := development.Run(ctx, db); err != nil {
		t.Fatal(err)
	}
	var changed int
	err = db.QueryRow(`SELECT count(*) FROM before_rerun b FULL JOIN matches m USING (id)
		WHERE (to_jsonb(b) - 'updated_at') IS DISTINCT FROM (to_jsonb(m) - 'updated_at')`).Scan(&changed)
	if err != nil || changed != 0 {
		t.Fatalf("unexpected changes on rerun = %d: %v", changed, err)
	}
	var kickoffCorrect bool
	err = db.QueryRow(`SELECT scheduled_start_at = '2026-09-08 16:45' AND home_score = 1 AND away_score = 0
		AND actual_start_at = '2026-09-08 16:46:32.404' AND finished_at = '2026-09-08 18:46:14.426'
		FROM matches WHERE status = 'finished' AND home_team_id = (SELECT id FROM teams WHERE slug = 'aek-athens')
		AND away_team_id = (SELECT id FROM teams WHERE slug = 'lask')`).Scan(&kickoffCorrect)
	if err != nil || !kickoffCorrect {
		t.Fatalf("incorrect persisted AEK–LASK score or UTC timestamps: %v", err)
	}
	t.Run("ambiguous existing fixture", func(t *testing.T) {
		_, err := db.Exec(`INSERT INTO matches (stage_id, home_team_id,
			away_team_id, scheduled_start_at, prediction_lock_at)
			SELECT stage_id, home_team_id, away_team_id, scheduled_start_at,
				prediction_lock_at FROM matches WHERE status = 'finished' LIMIT 1`)
		if err != nil {
			t.Fatal(err)
		}
		if err := development.Run(ctx, db); err == nil || !strings.Contains(err.Error(), "duplicate fixture") {
			t.Fatalf("expected an ambiguous fixture error, got %v", err)
		}
	})
}

func TestMissingMatchTeamRollsBackPostgres(t *testing.T) {
	db := seedPostgres(t)
	_, err := db.Exec(`
		CREATE FUNCTION skip_team() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RETURN NULL; END $$;
		CREATE TRIGGER skip_villarreal BEFORE INSERT ON teams FOR EACH ROW
		WHEN (NEW.slug = 'villarreal') EXECUTE FUNCTION skip_team()
	`)
	if err != nil {
		t.Fatal(err)
	}
	err = development.Run(context.Background(), db)
	if err == nil || !strings.Contains(err.Error(), "seed matches:") || !strings.Contains(err.Error(), "villarreal") {
		t.Fatalf("expected a match error identifying the missing club, got %v", err)
	}
	var rows int
	if err := db.QueryRow(`SELECT (SELECT count(*) FROM countries) + (SELECT count(*) FROM matches)`).Scan(&rows); err != nil || rows != 0 {
		t.Fatalf("rows after rollback = %d: %v", rows, err)
	}
}
