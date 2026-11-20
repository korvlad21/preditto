package service

import (
	"context"

	"preditto/internal/model"
	"preditto/internal/repository"
)

type TournamentParticipantService struct {
	repo *repository.TournamentParticipantRepository
}

func NewTournamentParticipantService(repo *repository.TournamentParticipantRepository) *TournamentParticipantService {
	return &TournamentParticipantService{repo: repo}
}

func (s *TournamentParticipantService) GetParticipantInfo(ctx context.Context, userID, tournamentID int64) (model.ParticipantInfo, error) {
	return s.repo.GetParticipantInfo(ctx, userID, tournamentID)
}
