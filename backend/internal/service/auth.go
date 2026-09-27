package service

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"preditto/internal/auth"
	authdto "preditto/internal/dto/auth"
	"preditto/internal/model"
	"preditto/internal/repository"
)

var (
	ErrInvalidRequest     = errors.New("invalid request")
	ErrInvalidCredentials = errors.New("invalid login or password")
	ErrUserBlocked        = errors.New("user is blocked")
)

type AuthService struct {
	repo      *repository.AuthRepository
	tokens    *auth.TokenManager
	dummyHash []byte
}

func NewAuthService(repo *repository.AuthRepository, tokens *auth.TokenManager) (*AuthService, error) {
	// Unknown logins still perform the same expensive bcrypt comparison.
	hash, err := bcrypt.GenerateFromPassword([]byte("unused timing equalization password"), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("initialize password verifier: %w", err)
	}
	return &AuthService{repo: repo, tokens: tokens, dummyHash: hash}, nil
}

func (s *AuthService) Register(ctx context.Context, req authdto.RegisterRequest) (authdto.AuthResponse, error) {
	if err := req.Validate(); err != nil {
		return authdto.AuthResponse{}, ErrInvalidRequest
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return authdto.AuthResponse{}, fmt.Errorf("hash password: %w", err)
	}
	user := model.User{
		Username: req.Username, Email: req.Email, PasswordHash: string(hash), Status: model.UserStatusActive,
		FirstName: req.FirstName, LastName: req.LastName, FavoriteTeamID: req.FavoriteTeamID,
	}
	var pair authdto.TokenResponse
	err = s.repo.InTransaction(ctx, func(tx *repository.AuthRepository) error {
		if err := tx.CreateUser(ctx, &user); err != nil {
			return err
		}
		if err := tx.AssignRole(ctx, user.ID, "participant"); err != nil {
			return err
		}
		var err error
		pair, err = s.issue(ctx, tx, user.ID)
		return err
	})
	if err != nil {
		return authdto.AuthResponse{}, err
	}
	return authdto.AuthResponse{TokenResponse: pair, User: user}, nil
}

func (s *AuthService) Login(ctx context.Context, req authdto.LoginRequest) (authdto.AuthResponse, error) {
	user, err := s.repo.FindUserByLogin(ctx, req.Login)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return authdto.AuthResponse{}, err
	}
	hash := s.dummyHash
	if err == nil {
		hash = []byte(user.PasswordHash)
	}
	passwordErr := bcrypt.CompareHashAndPassword(hash, []byte(req.Password))
	if err != nil || passwordErr != nil || len(req.Password) > 72 {
		return authdto.AuthResponse{}, ErrInvalidCredentials
	}
	var pair authdto.TokenResponse
	err = s.repo.InTransaction(ctx, func(tx *repository.AuthRepository) error {
		current, err := tx.LockUser(ctx, user.ID)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrInvalidCredentials
		}
		if err != nil {
			return err
		}
		if current.PasswordHash != user.PasswordHash {
			return ErrInvalidCredentials
		}
		if current.Status != model.UserStatusActive {
			return ErrUserBlocked
		}
		user = current
		pair, err = s.issue(ctx, tx, user.ID)
		return err
	})
	if err != nil {
		return authdto.AuthResponse{}, err
	}
	return authdto.AuthResponse{TokenResponse: pair, User: user}, nil
}

func (s *AuthService) Refresh(ctx context.Context, raw string) (authdto.TokenResponse, error) {
	claims, err := s.tokens.ParseRefreshToken(raw)
	if err != nil {
		return authdto.TokenResponse{}, auth.ErrInvalidToken
	}
	var pair authdto.TokenResponse
	err = s.repo.InTransaction(ctx, func(tx *repository.AuthRepository) error {
		// Lock the parent first, matching PostgreSQL's cascading user deletion.
		user, err := tx.LockUser(ctx, claims.UserID())
		if errors.Is(err, repository.ErrNotFound) {
			return auth.ErrInvalidToken
		}
		if err != nil {
			return err
		}
		if user.Status != model.UserStatusActive {
			return ErrUserBlocked
		}
		// Check expiration after both locks have been acquired, including any wait.
		session, err := s.session(ctx, tx, raw, claims)
		if err != nil {
			return err
		}
		if session.RevokedAt != nil {
			return auth.ErrInvalidToken
		}
		if err := tx.RevokeRefreshToken(ctx, session.ID); err != nil {
			return err
		}
		pair, err = s.issue(ctx, tx, user.ID)
		return err
	})
	if err != nil {
		return authdto.TokenResponse{}, err
	}
	return pair, nil
}

// Logout is idempotent for an already revoked, otherwise valid session.
func (s *AuthService) Logout(ctx context.Context, raw string) error {
	claims, err := s.tokens.ParseRefreshToken(raw)
	if err != nil {
		return auth.ErrInvalidToken
	}
	return s.repo.InTransaction(ctx, func(tx *repository.AuthRepository) error {
		session, err := s.session(ctx, tx, raw, claims)
		if err != nil {
			return err
		}
		return tx.RevokeRefreshToken(ctx, session.ID)
	})
}

func (s *AuthService) session(ctx context.Context, tx *repository.AuthRepository, raw string, claims *auth.Claims) (model.RefreshToken, error) {
	session, err := tx.LockRefreshToken(ctx, claims.ID)
	if errors.Is(err, repository.ErrNotFound) {
		return model.RefreshToken{}, auth.ErrInvalidToken
	}
	if err != nil {
		return model.RefreshToken{}, err
	}
	now := time.Now()
	if session.UserID != claims.UserID() || !session.ExpiresAt.After(now) || !claims.ExpiresAt.After(now) || subtle.ConstantTimeCompare(session.TokenHash, auth.HashRefreshToken(raw)) != 1 {
		return model.RefreshToken{}, auth.ErrInvalidToken
	}
	return session, nil
}

func (s *AuthService) issue(ctx context.Context, tx *repository.AuthRepository, userID int64) (authdto.TokenResponse, error) {
	pair, session, err := s.tokens.GeneratePair(userID)
	if err != nil {
		return authdto.TokenResponse{}, err
	}
	if err := tx.CreateRefreshToken(ctx, session); err != nil {
		return authdto.TokenResponse{}, err
	}
	return pair, nil
}
