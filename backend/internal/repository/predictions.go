package repository

import (
	"context"
	"database/sql"
	"fmt"

	"preditto/internal/model"
)

type PredictionRepository struct {
	db *sql.DB
}

func NewPredictionRepository(db *sql.DB) *PredictionRepository {
	return &PredictionRepository{db: db}
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
