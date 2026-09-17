package service

import (
	"context"

	"lockerin-backend/internal/domain"
	"lockerin-backend/internal/repository"

	"github.com/google/uuid"
)

type LocationService struct {
	locRepo    repository.LocationRepository
	lockerRepo repository.LockerRepository
	slotRepo   repository.SlotRepository
}

func NewLocationService(locRepo repository.LocationRepository, lockerRepo repository.LockerRepository, slotRepo repository.SlotRepository) *LocationService {
	return &LocationService{
		locRepo:    locRepo,
		lockerRepo: lockerRepo,
		slotRepo:   slotRepo,
	}
}

func (s *LocationService) ListLocations(ctx context.Context, limit, offset int) ([]domain.Location, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.locRepo.List(ctx, limit, offset)
}

func (s *LocationService) GetLocation(ctx context.Context, id uuid.UUID) (*domain.Location, error) {
	return s.locRepo.GetByID(ctx, id)
}

func (s *LocationService) GetLocationAvailability(ctx context.Context, id uuid.UUID) (*domain.LocationAvailability, error) {
	loc, err := s.locRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	lockers, err := s.lockerRepo.ListByLocation(ctx, id)
	if err != nil {
		return nil, err
	}

	totalSlots := 0
	availSlots := 0
	occupiedSlots := 0
	offlineSlots := 0
	smallAvail := 0
	mediumAvail := 0
	largeAvail := 0

	for _, lk := range lockers {
		slots, err := s.slotRepo.ListByLocker(ctx, lk.ID)
		if err != nil {
			continue
		}
		for _, sl := range slots {
			totalSlots++
			if lk.Status != domain.LockerOnline || sl.Status == domain.SlotOffline || sl.Status == domain.SlotMaintenance {
				offlineSlots++
			} else if sl.Status == domain.SlotAvailable {
				availSlots++
				switch sl.Size {
				case domain.SizeSmall:
					smallAvail++
				case domain.SizeMedium:
					mediumAvail++
				case domain.SizeLarge, domain.SizeXL:
					largeAvail++
				}
			} else {
				occupiedSlots++
			}
		}
	}

	return &domain.LocationAvailability{
		LocationID:     loc.ID,
		LocationName:   loc.Name,
		TotalSlots:     totalSlots,
		AvailableSlots: availSlots,
		OccupiedSlots:  occupiedSlots,
		OfflineSlots:   offlineSlots,
		SmallSlots:     smallAvail,
		MediumSlots:    mediumAvail,
		LargeSlots:     largeAvail,
		IsOpenNow:      true,
	}, nil
}
