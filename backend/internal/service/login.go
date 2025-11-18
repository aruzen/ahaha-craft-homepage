package service

import (
	"context"
	"errors"
	"log"
	"time"

	"backend/internal/domain"
	"backend/internal/repository"

	"github.com/jackc/pgx/v5"
)

// LoginService はログイン処理の具象実装を提供する雛形。
type LoginService struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.LoginSessionRepository
	logger      *log.Logger
}

func NewLoginService(userRepo *repository.UserRepository, sessionRepo *repository.LoginSessionRepository, logger *log.Logger) *LoginService {
	if logger == nil {
		logger = log.Default()
	}
	return &LoginService{userRepo: userRepo, sessionRepo: sessionRepo, logger: logger}
}

func (s *LoginService) Login(ctx context.Context, credential domain.AdminCredential) (domain.SessionData, domain.UserRole, error) {
	user, err := s.userRepo.FindByName(ctx, credential.Name())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logError("user not found", err)
			return domain.SessionData{}, "", domain.ErrInvalidCredential
		}
		s.logError("find user by name", err)
		return domain.SessionData{}, "", err
	}

	/*
		// 管理者以外は認証失敗とする
		if user.Role() != domain.UserRoleAdmin {
			return domain.SessionData{}, domain.ErrInvalidCredential
		}
	*/

	if err := user.HashedPassword().Verify(credential.Password()); err != nil {
		s.logError("password verification failed", err)
		return domain.SessionData{}, "", domain.ErrInvalidCredential
	}

	token, err := domain.NewLoginSessionToken()
	if err != nil {
		s.logError("issue login token", err)
		return domain.SessionData{}, "", err
	}

	sessionData, err := domain.NewSessionData(user.ID(), token)
	if err != nil {
		s.logError("build session data", err)
		return domain.SessionData{}, "", err
	}

	hashedToken, err := token.Hash()
	if err != nil {
		s.logError("hash login token", err)
		return domain.SessionData{}, "", err
	}

	session, err := domain.NewLoginSession(user.ID(), hashedToken, time.Now())
	if err != nil {
		s.logError("build login session", err)
		return domain.SessionData{}, "", err
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		s.logError("persist login session", err)
		return domain.SessionData{}, "", err
	}

	return sessionData, user.Role(), nil
}

func (s *LoginService) logError(action string, err error) {
	if err == nil {
		return
	}
	s.logger.Printf("[LoginService] %s: %v", action, err)
}
