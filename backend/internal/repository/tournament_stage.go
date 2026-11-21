package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"preditto/internal/model"
)

type TournamentStageRepository struct {
	db *sql.DB
}

func NewTournamentStageRepository(db *sql.DB) *TournamentStageRepository {
	return &TournamentStageRepository{db: db}
}

const tournamentStageSelect = `
	SELECT s.id, s.tournament_id, s.code, s.name, s.stage_kind, s.sequence_no,
	       s.status, s.predictions_open_at, s.outcome_points, s.exact_score_points,
	       COUNT(m.id) AS count_matches, COUNT(p.match_id) AS count_predictions
	FROM tournament_stages s
	LEFT JOIN matches m ON m.stage_id = s.id
	LEFT JOIN predictions p ON p.match_id = m.id AND p.participant_id = $2
	WHERE s.tournament_id = $1
`

func (r *TournamentStageRepository) GetStages(ctx context.Context, tournamentID, participantID int64) ([]model.TournamentStage, error) {
	rows, err := r.db.QueryContext(ctx, tournamentStageSelect+`
		GROUP BY s.id
		ORDER BY s.sequence_no ASC, s.id ASC
	`, tournamentID, participantID)
	if err != nil {
		return nil, fmt.Errorf("query tournament stages: %w", err)
	}
	defer rows.Close()

	stages := make([]model.TournamentStage, 0)
	for rows.Next() {
		stage, err := scanTournamentStage(rows)
		if err != nil {
			return nil, fmt.Errorf("scan tournament stage: %w", err)
		}
		stages = append(stages, stage)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read tournament stages: %w", err)
	}
	return stages, nil
}

func (r *TournamentStageRepository) GetStage(ctx context.Context, tournamentID, participantID, stageID int64) (model.TournamentStage, error) {
	row := r.db.QueryRowContext(ctx, tournamentStageSelect+`
		AND s.id = $3
		GROUP BY s.id
	`, tournamentID, participantID, stageID)
	stage, err := scanTournamentStage(row)
	if errors.Is(err, sql.ErrNoRows) {
		return model.TournamentStage{}, ErrNotFound
	}
	if err != nil {
		return model.TournamentStage{}, fmt.Errorf("get tournament stage: %w", err)
	}
	return stage, nil
}

func scanTournamentStage(row interface{ Scan(...any) error }) (model.TournamentStage, error) {
	var stage model.TournamentStage
	err := row.Scan(&stage.ID, &stage.TournamentID, &stage.Code, &stage.Name, &stage.StageKind,
		&stage.SequenceNo, &stage.Status, &stage.PredictionsOpenAt, &stage.OutcomePoints,
		&stage.ExactScorePoints, &stage.CountMatches, &stage.CountPredictions)
	return stage, err
}
