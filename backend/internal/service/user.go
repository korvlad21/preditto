package service

import (
	"context"

	"preditto/internal/model"
	"preditto/internal/repository"
)

type UserService struct {
	repo *repository.AuthRepository
}

func NewUserService(repo *repository.AuthRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUserInfo(ctx context.Context, id int64) (model.UserInfo, error) {
	return s.repo.GetUserInfo(ctx, id)
}
