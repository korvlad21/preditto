package service

import (
	"context"

	"preditto/internal/model"
	"preditto/internal/repository"
)

type PredictionService struct {
	stages      *repository.TournamentStageRepository
	predictions *repository.PredictionRepository
}

func NewPredictionService(stages *repository.TournamentStageRepository, predictions *repository.PredictionRepository) *PredictionService {
	return &PredictionService{stages: stages, predictions: predictions}
}

func (s *PredictionService) GetStages(ctx context.Context, tournamentID, participantID int64) ([]model.TournamentStage, error) {
	return s.stages.GetStages(ctx, tournamentID, participantID)
}

func (s *PredictionService) GetStage(ctx context.Context, tournamentID, participantID, stageID int64) (model.TournamentStage, error) {
	return s.stages.GetStage(ctx, tournamentID, participantID, stageID)
}

func (s *PredictionService) GetUserPredictions(ctx context.Context, tournamentID, participantID, stageID int64) ([]model.MatchWithPrediction, error) {
	return s.predictions.GetUserPredictions(ctx, tournamentID, participantID, stageID)
}

func (s *PredictionService) SavePrediction(ctx context.Context, tournamentID, participantID, userID, matchID int64, homeScore, awayScore int16) error {
	return s.predictions.SavePrediction(ctx, tournamentID, participantID, userID, matchID, homeScore, awayScore)
}
