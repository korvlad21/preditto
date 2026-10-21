package service

import (
	"context"

	"preditto/internal/model"
	"preditto/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUserInfo(ctx context.Context, id int64) (model.UserInfo, error) {
	return s.repo.GetUserInfo(ctx, id)
}
