package service

import (
	"context"

	"backend/internal/domain"
	"backend/internal/repository"
)

type HueGetService struct {
	hueRepo     *repository.HueRepository
	sessionRepo *repository.LoginSessionRepository
}

func NewHueGetService(hueRepo *repository.HueRepository, sessionRepo *repository.LoginSessionRepository) *HueGetService {
	return &HueGetService{hueRepo: hueRepo, sessionRepo: sessionRepo}
}

func (s *HueGetService) GetData(ctx context.Context, session domain.SessionData, recordRange domain.RecordRange) ([]domain.HueRecord, error) {
	return nil, nil
}
