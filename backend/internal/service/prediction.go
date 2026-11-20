package service

import (
	"context"

	"preditto/internal/model"
	"preditto/internal/repository"
)

type PredictionService struct {
	stages *repository.TournamentStageRepository
}

func NewPredictionService(stages *repository.TournamentStageRepository) *PredictionService {
	return &PredictionService{stages: stages}
}

func (s *PredictionService) GetStages(ctx context.Context, tournamentID, participantID int64) ([]model.TournamentStage, error) {
	return s.stages.GetStages(ctx, tournamentID, participantID)
}
