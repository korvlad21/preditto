package service

import (
	"context"
	"errors"
	"fmt"
	"io"

	usersdto "preditto/internal/dto/users"
	"preditto/internal/model"
	"preditto/internal/repository"
)

type UserService struct {
	repo    *repository.UserRepository
	avatars *AvatarStore
}

func NewUserService(repo *repository.UserRepository, avatars *AvatarStore) *UserService {
	return &UserService{repo: repo, avatars: avatars}
}

var ErrAvatarTooLarge = errors.New("avatar exceeds size limit")

func (s *UserService) UpdateUser(ctx context.Context, req usersdto.UpdateUserRequest, avatar io.Reader) (model.UserInfo, error) {
	if err := req.Validate(); err != nil {
		return model.UserInfo{}, ErrInvalidRequest
	}
	var uploadedURL *string
	var cleanup func() error
	if avatar != nil {
		if req.AvatarURL != nil && *req.AvatarURL != "" {
			return model.UserInfo{}, ErrInvalidRequest
		}
		url, remove, err := s.avatars.save(avatar)
		if err != nil {
			return model.UserInfo{}, err
		}
		uploadedURL, cleanup = &url, remove
	}
	user, err := s.repo.UpdateUser(ctx, req, uploadedURL)
	if err != nil && cleanup != nil && !errors.Is(err, repository.ErrUserCommitUncertain) {
		if cleanupErr := cleanup(); cleanupErr != nil {
			return model.UserInfo{}, fmt.Errorf("avatar cleanup after user update (%v): %w", err, cleanupErr)
		}
	}
	return user, err
}

func (s *UserService) GetUserInfo(ctx context.Context, id int64) (model.UserInfo, error) {
	return s.repo.GetUserInfo(ctx, id)
}
