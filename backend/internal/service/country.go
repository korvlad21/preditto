package service

import (
	"context"

	"preditto/internal/model"
	"preditto/internal/repository"
)

type CountryService struct {
	repo *repository.CountryRepository
}

func NewCountryService(repo *repository.CountryRepository) *CountryService {
	return &CountryService{repo: repo}
}

func (s *CountryService) GetAllCountries(ctx context.Context) ([]model.Country, error) {
	return s.repo.GetAllCountries(ctx)
}
