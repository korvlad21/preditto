package service

import (
	"context"
	"errors"
	"fmt"

	"preditto/internal/model"
	"preditto/internal/repository"
)

var ErrTournamentUpdateFailed = errors.New("tournament selection update failed")

type TournamentService struct {
	repo  *repository.TournamentRepository
	users *repository.UserRepository
}

func NewTournamentService(repo *repository.TournamentRepository, users *repository.UserRepository) *TournamentService {
	return &TournamentService{repo: repo, users: users}
}

func (s *TournamentService) SelectTournament(ctx context.Context, userID, tournamentID int64) error {
	exists, err := s.repo.Exists(ctx, tournamentID)
	if err != nil {
		return err
	}
	if !exists {
		return repository.ErrTournamentNotFound
	}
	if err := s.users.UpdateSelectedTournament(ctx, userID, &tournamentID); err != nil {
		return fmt.Errorf("%w: %w", ErrTournamentUpdateFailed, err)
	}
	return nil
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
