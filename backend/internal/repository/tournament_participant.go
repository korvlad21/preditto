package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"preditto/internal/model"
)

type TournamentParticipantRepository struct {
	db *sql.DB
}

func NewTournamentParticipantRepository(db *sql.DB) *TournamentParticipantRepository {
	return &TournamentParticipantRepository{db: db}
}

func (r *TournamentParticipantRepository) GetParticipantInfo(ctx context.Context, userID, tournamentID int64) (model.ParticipantInfo, error) {
	var participant model.ParticipantInfo
	err := r.db.QueryRowContext(ctx, `
		SELECT id, joined_at
		FROM tournament_participants
		WHERE user_id = $1 AND tournament_id = $2 AND status = $3
	`, userID, tournamentID, model.TournamentParticipantStatusActive).Scan(&participant.ID, &participant.JoinedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.ParticipantInfo{}, ErrNotFound
	}
	if err != nil {
		return model.ParticipantInfo{}, fmt.Errorf("get tournament participant: %w", err)
	}
	return participant, nil
}
