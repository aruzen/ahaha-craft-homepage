package service

import (
	"context"
	"log"

	"backend/internal/domain"
	"backend/internal/repository"
)

type HueSaveService struct {
	hueRepo *repository.HueRepository
}

func NewHueSaveService(hueRepo *repository.HueRepository) *HueSaveService {
	return &HueSaveService{hueRepo: hueRepo}
}

func (s *HueSaveService) SaveResult(ctx context.Context, record domain.HueRecord) error {
	log.Print("SaveResult")
	return s.hueRepo.Save(ctx, record)
}
