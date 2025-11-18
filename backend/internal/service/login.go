package service

import (
	"context"

	"backend/internal/domain"
	"backend/internal/repository"
)

// LoginService はログイン処理の具象実装を提供する雛形。
type LoginService struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.LoginSessionRepository
}

func NewLoginService(userRepo *repository.UserRepository, sessionRepo *repository.LoginSessionRepository) *LoginService {
	return &LoginService{userRepo: userRepo, sessionRepo: sessionRepo}
}

func (s *LoginService) Login(ctx context.Context, credential domain.AdminCredential) (domain.SessionData, error) {
	return domain.SessionData{}, nil
}
