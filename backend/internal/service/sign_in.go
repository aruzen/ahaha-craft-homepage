package service

import (
	"context"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"

	"backend/internal/domain"
	"backend/internal/repository"
)

// SignInService はサインイン処理を司る具体実装の雛形。
type SignInService struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.LoginSessionRepository
	logger      *log.Logger
}

func NewSignInService(userRepo *repository.UserRepository, sessionRepo *repository.LoginSessionRepository, logger *log.Logger) *SignInService {
	if logger == nil {
		logger = log.Default()
	}
	return &SignInService{userRepo: userRepo, sessionRepo: sessionRepo, logger: logger}
}

func (s *SignInService) SignIn(ctx context.Context, credential domain.SignInCredential) (domain.SessionData, error) {
	now := time.Now()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(credential.Password()), bcrypt.DefaultCost)
	if err != nil {
		s.logError("hash password", err)
		return domain.SessionData{}, err
	}

	password, err := domain.NewHashedPassword(string(hashedPassword))
	if err != nil {
		s.logError("build hashed password domain", err)
		return domain.SessionData{}, err
	}

	user, err := domain.NewUser(credential.Name(), credential.Email(), password, domain.UserRoleUser, now)
	if err != nil {
		s.logError("build user domain", err)
		return domain.SessionData{}, err
	}

	if err = s.userRepo.Create(ctx, user); err != nil {
		s.logError("create user", err)
		return domain.SessionData{}, err
	}

	token, err := domain.NewLoginSessionToken()
	if err != nil {
		s.logError("issue session token", err)
		return domain.SessionData{}, err
	}

	data, err := domain.NewSessionData(user.ID(), token)
	if err != nil {
		s.logError("build session data", err)
		return domain.SessionData{}, err
	}

	hashedToken, err := data.Token().Hash()
	if err != nil {
		s.logError("hash login token", err)
		return domain.SessionData{}, err
	}

	session, err := domain.NewLoginSession(data.UserID(), hashedToken, now)
	if err != nil {
		s.logError("build login session", err)
		return domain.SessionData{}, err
	}

	if err = s.sessionRepo.Create(ctx, session); err != nil {
		s.logError("persist login session", err)
		return domain.SessionData{}, err
	}
	return data, nil
}

func (s *SignInService) logError(action string, err error) {
	if err == nil {
		return
	}
	s.logger.Printf("[SignInService] %s: %v", action, err)
}
