package development

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed champions_league_2026_27_matches.json
var championsLeagueMatches string

// Each timestamp in the UEFA snapshot carries its own UTC offset. Convert to
// UTC wall time before writing the schema's timestamp-without-time-zone fields.
const matchFixturesSQL = `
	WITH fixtures AS (
		SELECT f.stage, f.home, f.away, s.id AS stage_id, h.id AS home_id, a.id AS away_id,
			f.home_score, f.away_score, f.kickoff AT TIME ZONE 'UTC' AS scheduled_start_at,
			(f.kickoff - interval '1 hour') AT TIME ZONE 'UTC' AS prediction_lock_at,
			f.actual_start AT TIME ZONE 'UTC' AS actual_start_at,
			(f.full_time + interval '10 minutes') AT TIME ZONE 'UTC' AS finished_at, f.status
		FROM jsonb_to_recordset($1::jsonb) AS f(stage text, home text, away text,
			kickoff timestamptz, home_score smallint, away_score smallint,
			actual_start timestamptz, full_time timestamptz, status text)
		LEFT JOIN tournaments t ON t.slug = $2
		LEFT JOIN tournament_stages s ON s.tournament_id = t.id AND s.code = f.stage
		LEFT JOIN teams h ON h.slug = f.home
		LEFT JOIN teams a ON a.slug = f.away
	)
`

func seedMatches(ctx context.Context, tx *sql.Tx) error {
	// The schema has no unique key for a fixture. Keep the identity check and
	// merge together even if another transaction is writing matches.
	if _, err := tx.ExecContext(ctx, "LOCK TABLE matches IN SHARE ROW EXCLUSIVE MODE"); err != nil {
		return fmt.Errorf("lock matches: %w", err)
	}
	var stage, home, away, problem string
	err := tx.QueryRowContext(ctx, matchFixturesSQL+`
		SELECT f.stage, f.home, f.away, CASE
			WHEN f.stage_id IS NULL THEN 'missing stage'
			WHEN f.home_id IS NULL THEN 'missing home team'
			WHEN f.away_id IS NULL THEN 'missing away team'
			ELSE 'duplicate fixture' END
		FROM fixtures f
		LEFT JOIN matches m ON m.stage_id = f.stage_id AND m.home_team_id = f.home_id
			AND m.away_team_id = f.away_id
		GROUP BY f.stage, f.home, f.away, f.stage_id, f.home_id, f.away_id
		HAVING f.stage_id IS NULL OR f.home_id IS NULL OR f.away_id IS NULL OR count(*) > 1
		LIMIT 1
	`, championsLeagueMatches, championsLeague20262027Slug).Scan(&stage, &home, &away, &problem)
	if err == nil {
		return fmt.Errorf("%s %s, %s vs %s: %s", championsLeague20262027Slug, stage, home, away, problem)
	}
	if err != sql.ErrNoRows {
		return fmt.Errorf("resolve match fixtures: %w", err)
	}
	_, err = tx.ExecContext(ctx, matchFixturesSQL+`
		MERGE INTO matches m USING fixtures f
		ON m.stage_id = f.stage_id AND m.home_team_id = f.home_id
			AND m.away_team_id = f.away_id
		WHEN MATCHED AND (f.status <> 'scheduled' OR m.status = 'scheduled')
			AND (m.scheduled_start_at, m.prediction_lock_at, m.actual_start_at, m.finished_at,
				 m.home_score, m.away_score, m.status) IS DISTINCT FROM
				(f.scheduled_start_at, f.prediction_lock_at, f.actual_start_at, f.finished_at,
				 f.home_score, f.away_score, f.status)
		THEN UPDATE SET scheduled_start_at = f.scheduled_start_at, prediction_lock_at = f.prediction_lock_at,
			actual_start_at = f.actual_start_at, finished_at = f.finished_at, home_score = f.home_score,
			away_score = f.away_score, status = f.status, updated_at = CURRENT_TIMESTAMP AT TIME ZONE 'UTC'
		WHEN NOT MATCHED THEN INSERT (stage_id, home_team_id, away_team_id,
			scheduled_start_at, prediction_lock_at, actual_start_at, finished_at, home_score,
			away_score, status, created_at, updated_at)
		VALUES (f.stage_id, f.home_id, f.away_id, f.scheduled_start_at, f.prediction_lock_at,
			f.actual_start_at, f.finished_at, f.home_score, f.away_score, f.status,
			CURRENT_TIMESTAMP AT TIME ZONE 'UTC', CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
	`, championsLeagueMatches, championsLeague20262027Slug)
	if err != nil {
		return fmt.Errorf("merge fixtures for %s: %w", championsLeague20262027Slug, err)
	}
	return nil
}
