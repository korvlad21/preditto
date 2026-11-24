package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"preditto/internal/model"
)

type PredictionRepository struct {
	db *sql.DB
}

func NewPredictionRepository(db *sql.DB) *PredictionRepository {
	return &PredictionRepository{db: db}
}

var (
	ErrParticipantNotFound  = errors.New("active tournament participant not found")
	ErrPredictionsNotOpen   = errors.New("stage predictions have not opened")
	ErrPredictionLocked     = errors.New("match closed for predictions")
	ErrPredictionCalculated = errors.New("prediction already calculated")
)

func (r *PredictionRepository) SavePrediction(ctx context.Context, tournamentID, participantID, userID, matchID int64, homeScore, awayScore int16) error {
	// Serialize this participant's submissions even when the prediction row does
	// not exist yet. Shared fixture locks allow other participants to save concurrently.
	// Materialize the locked inputs before testing the UTC wall-clock window.
	var outcome string
	err := r.db.QueryRowContext(ctx, `
		WITH participant AS MATERIALIZED (
			SELECT id FROM tournament_participants
			WHERE tournament_id = $1 AND id = $2 AND user_id = $3 AND status = $7
			FOR NO KEY UPDATE
		), fixture AS MATERIALIZED (
			SELECT m.id, m.status, m.prediction_lock_at, s.predictions_open_at
			FROM matches m JOIN tournament_stages s ON s.id = m.stage_id
			WHERE s.tournament_id = $1 AND m.id = $4
			FOR SHARE OF m, s
		), eligibility AS MATERIALIZED (
			SELECT f.id, f.prediction_lock_at, CASE
				WHEN NOT EXISTS (SELECT FROM participant) THEN 'no_participant'
				WHEN f.id IS NULL THEN 'no_match'
				WHEN status IN ('live', 'finished') THEN 'locked'
				WHEN predictions_open_at IS NULL OR predictions_open_at > (clock_timestamp() AT TIME ZONE 'UTC') THEN 'not_open'
			END AS refusal
			FROM (SELECT 1) anchor LEFT JOIN fixture f ON true
		), saved AS (
			INSERT INTO predictions (participant_id, match_id, home_score, away_score)
			SELECT $2, id, $5, $6 FROM eligibility
			WHERE refusal IS NULL
			  AND (clock_timestamp() AT TIME ZONE 'UTC') < prediction_lock_at
			ON CONFLICT ON CONSTRAINT uq_predictions_participant_match DO UPDATE
			SET home_score = EXCLUDED.home_score, away_score = EXCLUDED.away_score,
			    updated_at = clock_timestamp()
			WHERE predictions.calculated_at IS NULL
			  AND (clock_timestamp() AT TIME ZONE 'UTC') < (SELECT prediction_lock_at FROM eligibility)
			RETURNING id
		)
		SELECT CASE
			WHEN EXISTS (SELECT FROM saved) THEN 'saved'
			WHEN refusal IS NOT NULL THEN refusal
			WHEN prediction_lock_at <= (clock_timestamp() AT TIME ZONE 'UTC') THEN 'locked'
			ELSE 'calculated'
		END
		FROM eligibility
	`, tournamentID, participantID, userID, matchID, homeScore, awayScore, model.TournamentParticipantStatusActive).Scan(&outcome)
	if err != nil {
		return fmt.Errorf("save prediction: %w", err)
	}
	switch outcome {
	case "saved":
		return nil
	case "no_participant":
		return ErrParticipantNotFound
	case "no_match":
		return ErrNotFound
	case "locked":
		return ErrPredictionLocked
	case "not_open":
		return ErrPredictionsNotOpen
	case "calculated":
		return ErrPredictionCalculated
	default:
		return fmt.Errorf("unexpected save prediction outcome %q", outcome)
	}
}

func (r *PredictionRepository) GetUserPredictions(ctx context.Context, tournamentID, participantID, stageID int64) ([]model.MatchWithPrediction, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT m.id, m.home_score, m.away_score, m.scheduled_start_at, m.prediction_lock_at,
		       m.actual_start_at, m.finished_at, m.status,
		       home.id, home.name, home.short_name, home.logo_url,
		       away.id, away.name, away.short_name, away.logo_url,
		       p.id, p.home_score, p.away_score, p.result_type, p.awarded_points
		FROM tournament_stages s
		LEFT JOIN matches m ON m.stage_id = s.id
		LEFT JOIN teams home ON home.id = m.home_team_id
		LEFT JOIN teams away ON away.id = m.away_team_id
		LEFT JOIN predictions p ON p.match_id = m.id AND p.participant_id = $2
		WHERE s.tournament_id = $1 AND s.id = $3
		ORDER BY m.scheduled_start_at ASC, m.id ASC
	`, tournamentID, participantID, stageID)
	if err != nil {
		return nil, fmt.Errorf("query stage matches with predictions: %w", err)
	}
	defer rows.Close()

	var matches []model.MatchWithPrediction
	for rows.Next() {
		var item model.MatchWithPrediction
		var prediction model.Prediction
		var matchID, homeID, awayID, predictionID sql.NullInt64
		var scheduledStart, predictionLock sql.NullTime
		var status, homeName, awayName sql.NullString
		var predictedHome, predictedAway sql.NullInt16
		if err := rows.Scan(
			&matchID, &item.Match.HomeScore, &item.Match.AwayScore, &scheduledStart, &predictionLock,
			&item.Match.ActualStartAt, &item.Match.FinishedAt, &status,
			&homeID, &homeName, &item.HomeTeam.ShortName, &item.HomeTeam.LogoURL,
			&awayID, &awayName, &item.AwayTeam.ShortName, &item.AwayTeam.LogoURL,
			&predictionID, &predictedHome, &predictedAway, &prediction.ResultType, &prediction.AwardedPoints,
		); err != nil {
			return nil, fmt.Errorf("scan stage match with prediction: %w", err)
		}
		// An existing empty stage produces one row with a NULL match ID.
		if !matchID.Valid {
			matches = []model.MatchWithPrediction{}
			continue
		}
		item.Match.ID = matchID.Int64
		item.Match.ScheduledStartAt = scheduledStart.Time
		item.Match.PredictionLockAt = predictionLock.Time
		item.Match.Status = status.String
		item.HomeTeam.ID, item.HomeTeam.Name = homeID.Int64, homeName.String
		item.AwayTeam.ID, item.AwayTeam.Name = awayID.Int64, awayName.String
		if predictionID.Valid {
			prediction.ID = predictionID.Int64
			prediction.HomeScore, prediction.AwayScore = predictedHome.Int16, predictedAway.Int16
			item.Prediction = &prediction
		}
		matches = append(matches, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read stage matches with predictions: %w", err)
	}
	if matches == nil {
		return nil, ErrNotFound
	}
	return matches, nil
}
