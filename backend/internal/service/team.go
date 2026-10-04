package service

import (
	"context"

	"preditto/internal/model"
	"preditto/internal/repository"
)

type TeamService struct {
	repo *repository.TeamRepository
}

func NewTeamService(repo *repository.TeamRepository) *TeamService {
	return &TeamService{repo: repo}
}

func (s *TeamService) GetAllTeams(ctx context.Context, country string) ([]model.Team, error) {
	if country == "ALL" {
		country = ""
	}
	return s.repo.GetAllTeams(ctx, country)
}
