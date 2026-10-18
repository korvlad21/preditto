package service

import (
	"context"

	"preditto/internal/model"
	"preditto/internal/repository"
)

type TournamentService struct {
	repo *repository.TournamentRepository
}

func NewTournamentService(repo *repository.TournamentRepository) *TournamentService {
	return &TournamentService{repo: repo}
}

func (s *TournamentService) GetAllTournaments(ctx context.Context, status string, selectedTournamentID *int64) ([]model.Tournament, error) {
	if status == "all" {
		status = ""
	}
	tournaments, err := s.repo.GetAllTournaments(ctx, status)
	if err != nil {
		return nil, err
	}
	for i := range tournaments {
		tournaments[i].IsSelectedTournament = selectedTournamentID != nil && tournaments[i].ID == *selectedTournamentID
	}
	return tournaments, nil
}
