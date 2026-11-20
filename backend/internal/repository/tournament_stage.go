package repository

import (
	"context"
	"database/sql"
	"fmt"

	"preditto/internal/model"
)

type TournamentStageRepository struct {
	db *sql.DB
}

func NewTournamentStageRepository(db *sql.DB) *TournamentStageRepository {
	return &TournamentStageRepository{db: db}
}

func (r *TournamentStageRepository) GetStages(ctx context.Context, tournamentID, participantID int64) ([]model.TournamentStage, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT s.id, s.tournament_id, s.code, s.name, s.stage_kind, s.sequence_no,
		       s.status, s.predictions_open_at, s.outcome_points, s.exact_score_points,
		       COUNT(m.id), COUNT(p.match_id)
		FROM tournament_stages s
		LEFT JOIN matches m ON m.stage_id = s.id
		LEFT JOIN predictions p ON p.match_id = m.id AND p.participant_id = $2
		WHERE s.tournament_id = $1
		GROUP BY s.id
		ORDER BY s.sequence_no ASC, s.id ASC
	`, tournamentID, participantID)
	if err != nil {
		return nil, fmt.Errorf("query tournament stages: %w", err)
	}
	defer rows.Close()

	stages := make([]model.TournamentStage, 0)
	for rows.Next() {
		var s model.TournamentStage
		if err := rows.Scan(&s.ID, &s.TournamentID, &s.Code, &s.Name, &s.StageKind, &s.SequenceNo,
			&s.Status, &s.PredictionsOpenAt, &s.OutcomePoints, &s.ExactScorePoints,
			&s.CountMatches, &s.CountPredictions); err != nil {
			return nil, fmt.Errorf("scan tournament stage: %w", err)
		}
		stages = append(stages, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read tournament stages: %w", err)
	}
	return stages, nil
}
