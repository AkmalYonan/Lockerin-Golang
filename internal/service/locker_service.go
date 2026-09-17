package service

import (
	"context"

	"lockerin-backend/internal/domain"
	"lockerin-backend/internal/repository"

	"github.com/google/uuid"
)

type LockerService struct {
	lockerRepo repository.LockerRepository
	slotRepo   repository.SlotRepository
}

func NewLockerService(lockerRepo repository.LockerRepository, slotRepo repository.SlotRepository) *LockerService {
	return &LockerService{
		lockerRepo: lockerRepo,
		slotRepo:   slotRepo,
	}
}

func (s *LockerService) GetLocker(ctx context.Context, id uuid.UUID) (*domain.Locker, error) {
	locker, err := s.lockerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	slots, err := s.slotRepo.ListByLocker(ctx, id)
	if err == nil {
		locker.Slots = slots
	}
	return locker, nil
}

func (s *LockerService) GetLockerSlots(ctx context.Context, lockerID uuid.UUID) ([]domain.LockerSlot, error) {
	return s.slotRepo.ListByLocker(ctx, lockerID)
}
