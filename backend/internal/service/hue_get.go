package service

import (
	"context"
	"log"

	"backend/internal/domain"
	"backend/internal/repository"
)

type HueGetService struct {
	hueRepo     *repository.HueRepository
	sessionRepo *repository.LoginSessionRepository
	logger      *log.Logger
}

func NewHueGetService(hueRepo *repository.HueRepository, sessionRepo *repository.LoginSessionRepository, logger *log.Logger) *HueGetService {
	if logger == nil {
		logger = log.Default()
	}
	return &HueGetService{hueRepo: hueRepo, sessionRepo: sessionRepo, logger: logger}
}

func (s *HueGetService) GetData(ctx context.Context, session domain.SessionData, recordRange domain.RecordRange) ([]domain.HueRecord, error) {
	return nil, nil
}
