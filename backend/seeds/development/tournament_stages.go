package development

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"preditto/internal/model"
)

// All end times are estimates, not observed final whistles. Three hours allows
// for breaks, stoppage time, extra time and penalties, but not postponements.
const estimatedStageMatchDuration = 3 * time.Hour

type tournamentStage struct {
	code, name, kind, status string
	sequenceNo               int
	firstKickoff, endsAt     time.Time
	predictionsOpenAt        time.Time
	predictionsCloseAt       time.Time
	outcomePoints            int
	exactScorePoints         int
}

// UEFA calendar/fixtures checked on 2026-11-17. Explicit CET/CEST offsets are
// converted to UTC before arithmetic and storage in TIMESTAMP WITHOUT TIME ZONE.
// Knockout times are provisional regulation slots (earliest/latest for rounds
// of 32/16), not confirmed match assignments. See ../README.md#tournament-stages
// for official sources, the final's 18:00 exception and remaining uncertainties.
func championsLeagueStages(now time.Time) ([]tournamentStage, error) {
	schedule := []struct {
		code, name, kind           string
		firstKickoff, lastKickoff  string
		outcomePoints, exactPoints int
	}{
		{"league-1", "Общий этап — 1-й тур", "group", "2026-09-08T18:45:00+02:00", "2026-09-10T21:00:00+02:00", 1, 3},
		{"league-2", "Общий этап — 2-й тур", "group", "2026-10-13T18:45:00+02:00", "2026-10-14T21:00:00+02:00", 1, 3},
		{"league-3", "Общий этап — 3-й тур", "group", "2026-10-20T18:45:00+02:00", "2026-10-21T21:00:00+02:00", 1, 3},
		{"league-4", "Общий этап — 4-й тур", "group", "2026-11-03T18:45:00+01:00", "2026-11-04T21:00:00+01:00", 1, 3},
		{"league-5", "Общий этап — 5-й тур", "group", "2026-11-24T18:45:00+01:00", "2026-11-25T21:00:00+01:00", 1, 3},
		{"league-6", "Общий этап — 6-й тур", "group", "2026-12-08T18:45:00+01:00", "2026-12-09T21:00:00+01:00", 1, 3},
		{"league-7", "Общий этап — 7-й тур", "group", "2027-01-19T18:45:00+01:00", "2027-01-20T21:00:00+01:00", 1, 3},
		{"league-8", "Общий этап — 8-й тур", "group", "2027-01-27T21:00:00+01:00", "2027-01-27T21:00:00+01:00", 1, 3},
		{"round-32-1", "1/16 финала — первые матчи", "play-off", "2027-02-16T18:45:00+01:00", "2027-02-17T21:00:00+01:00", 2, 6},
		{"round-32-2", "1/16 финала — ответные матчи", "play-off", "2027-02-23T18:45:00+01:00", "2027-02-24T21:00:00+01:00", 2, 6},
		{"round-16-1", "1/8 финала — первые матчи", "play-off", "2027-03-09T18:45:00+01:00", "2027-03-10T21:00:00+01:00", 3, 9},
		{"round-16-2", "1/8 финала — ответные матчи", "play-off", "2027-03-16T18:45:00+01:00", "2027-03-17T21:00:00+01:00", 3, 9},
		{"quarter-final-1", "1/4 финала — первые матчи", "play-off", "2027-04-06T21:00:00+02:00", "2027-04-07T21:00:00+02:00", 4, 12},
		{"quarter-final-2", "1/4 финала — ответные матчи", "play-off", "2027-04-13T21:00:00+02:00", "2027-04-14T21:00:00+02:00", 4, 12},
		{"semi-final-1", "1/2 финала — первые матчи", "play-off", "2027-04-27T21:00:00+02:00", "2027-04-28T21:00:00+02:00", 5, 15},
		{"semi-final-2", "1/2 финала — ответные матчи", "play-off", "2027-05-04T21:00:00+02:00", "2027-05-05T21:00:00+02:00", 5, 15},
		{"final", "Финал", "play-off", "2027-06-05T18:00:00+02:00", "2027-06-05T18:00:00+02:00", 6, 18},
	}
	stages := make([]tournamentStage, 0, len(schedule))
	openAt := time.Date(2026, time.August, 20, 0, 0, 0, 0, time.UTC)
	activeAssigned := false
	for i, item := range schedule {
		first, err := time.Parse(time.RFC3339, item.firstKickoff)
		if err != nil {
			return nil, fmt.Errorf("stage %q first kickoff: %w", item.code, err)
		}
		last, err := time.Parse(time.RFC3339, item.lastKickoff)
		if err != nil {
			return nil, fmt.Errorf("stage %q last kickoff: %w", item.code, err)
		}
		first, last = first.UTC(), last.UTC()
		endsAt := last.Add(estimatedStageMatchDuration)
		closeAt := first.Add(-2 * time.Hour)
		if last.Before(first) || closeAt.Before(openAt) {
			return nil, fmt.Errorf("stage %q has an invalid match or prediction window", item.code)
		}
		status := model.TournamentStatusPlanned
		if !now.Before(endsAt) {
			status = model.TournamentStatusFinished
		} else if !activeAssigned {
			status = model.TournamentStatusActive
			activeAssigned = true
		}
		stages = append(stages, tournamentStage{
			code:               item.code,
			name:               item.name,
			kind:               item.kind,
			status:             status,
			sequenceNo:         i + 1,
			firstKickoff:       first,
			endsAt:             endsAt,
			predictionsOpenAt:  openAt,
			predictionsCloseAt: closeAt,
			outcomePoints:      item.outcomePoints, exactScorePoints: item.exactPoints,
		})
		openAt = last.Add(8 * time.Hour)
	}
	return stages, nil
}

func seedTournamentStages(ctx context.Context, tx *sql.Tx) error {
	var tournamentID int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM tournaments WHERE slug = $1`,
		championsLeague20262027Slug).Scan(&tournamentID); err != nil {
		return fmt.Errorf("find tournament for stages %q: %w", championsLeague20262027Slug, err)
	}
	stages, err := championsLeagueStages(time.Now())
	if err != nil {
		return err
	}
	for _, stage := range stages {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO tournament_stages (tournament_id, code, name, stage_kind, sequence_no,
				status, predictions_open_at, predictions_close_at, outcome_points, exact_score_points,
				created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
				CURRENT_TIMESTAMP AT TIME ZONE 'UTC', CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
			ON CONFLICT (tournament_id, code) DO UPDATE SET
				name = EXCLUDED.name,
				stage_kind = EXCLUDED.stage_kind,
				sequence_no = EXCLUDED.sequence_no,
				status = EXCLUDED.status,
				predictions_open_at = EXCLUDED.predictions_open_at,
				predictions_close_at = EXCLUDED.predictions_close_at,
				outcome_points = EXCLUDED.outcome_points,
				exact_score_points = EXCLUDED.exact_score_points,
				updated_at = CURRENT_TIMESTAMP AT TIME ZONE 'UTC'
		`, tournamentID, stage.code, stage.name, stage.kind, stage.sequenceNo, stage.status,
			stage.predictionsOpenAt, stage.predictionsCloseAt, stage.outcomePoints, stage.exactScorePoints)
		if err != nil {
			return fmt.Errorf("upsert tournament %q stage %q: %w", championsLeague20262027Slug, stage.code, err)
		}
	}
	return nil
}
