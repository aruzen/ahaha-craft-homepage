package service

import (
	"context"

	"backend/internal/domain"
	"backend/internal/repository"
)

// SignInService はサインイン処理を司る具体実装の雛形。
type SignInService struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.LoginSessionRepository
}

func NewSignInService(userRepo *repository.UserRepository, sessionRepo *repository.LoginSessionRepository) *SignInService {
	return &SignInService{userRepo: userRepo, sessionRepo: sessionRepo}
}

func (s *SignInService) SignIn(ctx context.Context, credential domain.SignInCredential) (domain.SessionData, error) {
	return domain.SessionData{}, nil
}
