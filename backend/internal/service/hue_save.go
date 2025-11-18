package service

import (
	"context"
	"log"

	"backend/internal/domain"
	"backend/internal/repository"
)

type HueSaveService struct {
	hueRepo *repository.HueRepository
	logger  *log.Logger
}

func NewHueSaveService(hueRepo *repository.HueRepository, logger *log.Logger) *HueSaveService {
	if logger == nil {
		logger = log.Default()
	}
	return &HueSaveService{hueRepo: hueRepo, logger: logger}
}

func (s *HueSaveService) SaveResult(ctx context.Context, record domain.HueRecord) error {
	if err := s.hueRepo.Save(ctx, record); err != nil {
		s.logger.Printf("[HueSaveService] save hue record: %v", err)
		return err
	}
	return nil
}
