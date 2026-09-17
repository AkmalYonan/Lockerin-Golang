package service

import (
	"context"
	"fmt"
	"time"

	"lockerin-backend/internal/config"
	"lockerin-backend/internal/domain"
	"lockerin-backend/internal/integration/locker_device"
	"lockerin-backend/internal/repository"

	"github.com/google/uuid"
)

type RentalService struct {
	rentalRepo   repository.RentalRepository
	slotRepo     repository.SlotRepository
	lockerRepo   repository.LockerRepository
	promoRepo    repository.PromoRepository
	pinService   *PINService
	deviceDisp   *locker_device.Dispatcher
	cfg          *config.Config
}

func NewRentalService(
	rentalRepo repository.RentalRepository,
	slotRepo repository.SlotRepository,
	lockerRepo repository.LockerRepository,
	promoRepo repository.PromoRepository,
	pinService *PINService,
	deviceDisp *locker_device.Dispatcher,
	cfg *config.Config,
) *RentalService {
	return &RentalService{
		rentalRepo: rentalRepo,
		slotRepo:   slotRepo,
		lockerRepo: lockerRepo,
		promoRepo:  promoRepo,
		pinService: pinService,
		deviceDisp: deviceDisp,
		cfg:        cfg,
	}
}

func (s *RentalService) Quote(ctx context.Context, req *domain.RentalQuoteRequest) (*domain.RentalQuoteResponse, error) {
	if req.DurationHours <= 0 {
		return nil, domain.NewAppError(400, "INVALID_DURATION", "Duration must be greater than 0", nil)
	}

	slot, err := s.slotRepo.GetByID(ctx, req.SlotID)
	if err != nil {
		return nil, domain.NewAppError(404, "SLOT_NOT_FOUND", "Locker slot not found", err)
	}

	subtotal := slot.BasePricePerHour * float64(req.DurationHours)
	discount := 0.0
	promoApplied := ""

	if req.PromoCode != "" {
		promo, err := s.promoRepo.GetByCode(ctx, req.PromoCode)
		if err == nil && promo.Status == "active" && time.Now().Before(promo.EndsAt) {
			if subtotal >= promo.MinOrderAmount {
				if promo.DiscountType == "percentage" {
					discount = (subtotal * promo.DiscountValue) / 100.0
					if promo.MaxDiscountAmount != nil && discount > *promo.MaxDiscountAmount {
						discount = *promo.MaxDiscountAmount
					}
				} else {
					discount = promo.DiscountValue
				}
				promoApplied = promo.Code
			}
		}
	}

	total := subtotal - discount
	if total < 0 {
		total = 0
	}

	return &domain.RentalQuoteResponse{
		SlotID:           slot.ID,
		SlotCode:         slot.SlotCode,
		Size:             string(slot.Size),
		DurationHours:    req.DurationHours,
		BasePricePerHour: slot.BasePricePerHour,
		Subtotal:         subtotal,
		DiscountAmount:   discount,
		TotalAmount:      total,
		PromoApplied:     promoApplied,
	}, nil
}

func (s *RentalService) Reserve(ctx context.Context, userID uuid.UUID, req *domain.ReserveRentalRequest) (*domain.ReserveRentalResponse, error) {
	if req.DurationHours <= 0 {
		return nil, domain.NewAppError(400, "INVALID_DURATION", "Duration must be greater than 0", nil)
	}

	slot, err := s.slotRepo.GetByID(ctx, req.SlotID)
	if err != nil {
		return nil, domain.NewAppError(404, "SLOT_NOT_FOUND", "Locker slot not found", err)
	}

	locker, err := s.lockerRepo.GetByID(ctx, slot.LockerID)
	if err != nil {
		return nil, domain.NewAppError(404, "LOCKER_NOT_FOUND", "Locker not found", err)
	}

	if locker.Status != domain.LockerOnline {
		return nil, domain.NewAppError(400, "LOCKER_OFFLINE", "Locker unit is currently offline or under maintenance", domain.ErrDeviceOffline)
	}

	if slot.Status != domain.SlotAvailable {
		return nil, domain.NewAppError(409, "SLOT_UNAVAILABLE", "Selected slot is already reserved or occupied", domain.ErrSlotUnavailable)
	}

	// Calculate quote
	quote, err := s.Quote(ctx, &domain.RentalQuoteRequest{
		SlotID:        req.SlotID,
		DurationHours: req.DurationHours,
		PromoCode:     req.PromoCode,
	})
	if err != nil {
		return nil, err
	}

	ttlMinutes := s.cfg.ReservationTTLMin
	if ttlMinutes <= 0 {
		ttlMinutes = 15
	}
	resExpiry := time.Now().Add(time.Duration(ttlMinutes) * time.Minute)

	rentalID := uuid.New()
	rental := &domain.Rental{
		ID:                   rentalID,
		UserID:               userID,
		LocationID:           locker.LocationID,
		LockerID:             locker.ID,
		SlotID:               slot.ID,
		Status:               domain.RentalReserved,
		DurationHours:        req.DurationHours,
		BasePriceSnapshot:    slot.BasePricePerHour,
		DiscountAmount:       quote.DiscountAmount,
		TotalAmount:          quote.TotalAmount,
		ReservationExpiresAt: &resExpiry,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Create rental (will fail if another active rental exists for this slot due to index)
	if err := s.rentalRepo.Create(ctx, rental); err != nil {
		return nil, domain.NewAppError(409, "SLOT_UNAVAILABLE", "Slot was just reserved by another user", domain.ErrSlotUnavailable)
	}

	// Update slot status
	_ = s.slotRepo.UpdateStatus(ctx, slot.ID, domain.SlotReserved, &rentalID)

	return &domain.ReserveRentalResponse{
		Rental:               rental,
		ReservationExpiresAt: resExpiry,
	}, nil
}

func (s *RentalService) StartRental(ctx context.Context, rentalID uuid.UUID, userID uuid.UUID) (*domain.Rental, error) {
	rent, err := s.rentalRepo.GetByID(ctx, rentalID)
	if err != nil {
		return nil, domain.NewAppError(404, "RENTAL_NOT_FOUND", "Rental not found", err)
	}

	if rent.UserID != userID {
		return nil, domain.NewAppError(403, "FORBIDDEN", "Unauthorized access to this rental", domain.ErrForbidden)
	}

	if rent.Status != domain.RentalPaid && rent.Status != domain.RentalReserved {
		return nil, domain.NewAppError(400, "INVALID_STATE", fmt.Sprintf("Cannot start rental in status '%s'", rent.Status), domain.ErrInvalidRentalState)
	}

	now := time.Now()
	expiresAt := now.Add(time.Duration(rent.DurationHours) * time.Hour)

	rent.Status = domain.RentalActive
	rent.StartedAt = &now
	rent.ExpiresAt = &expiresAt

	if err := s.rentalRepo.Update(ctx, rent); err != nil {
		return nil, err
	}

	_ = s.slotRepo.UpdateStatus(ctx, rent.SlotID, domain.SlotOccupied, &rentalID)

	return rent, nil
}

func (s *RentalService) OpenLocker(ctx context.Context, rentalID uuid.UUID, userID uuid.UUID, pin string) (*domain.OpenLockerResponse, error) {
	rent, err := s.rentalRepo.GetByID(ctx, rentalID)
	if err != nil {
		return nil, domain.NewAppError(404, "RENTAL_NOT_FOUND", "Rental not found", err)
	}

	if rent.UserID != userID {
		return nil, domain.NewAppError(403, "FORBIDDEN", "Unauthorized access to this rental", domain.ErrForbidden)
	}

	if rent.Status != domain.RentalActive && rent.Status != domain.RentalPaid {
		return nil, domain.NewAppError(400, "INVALID_STATE", "Locker can only be opened for active rentals", domain.ErrInvalidRentalState)
	}

	// If PIN provided, verify PIN
	if pin != "" {
		if err := s.pinService.VerifyPIN(ctx, rentalID, pin); err != nil {
			return nil, domain.NewAppError(401, "INVALID_PIN", "Invalid or expired security PIN", err)
		}
	}

	locker, err := s.lockerRepo.GetByID(ctx, rent.LockerID)
	if err != nil {
		return nil, domain.NewAppError(404, "LOCKER_NOT_FOUND", "Locker unit not found", err)
	}

	slot, err := s.slotRepo.GetByID(ctx, rent.SlotID)
	if err != nil {
		return nil, domain.NewAppError(404, "SLOT_NOT_FOUND", "Slot not found", err)
	}

	// Dispatch command to IoT device
	corrID := fmt.Sprintf("open-%s-%d", rentalID.String()[:8], time.Now().Unix())
	_, _ = s.deviceDisp.DispatchUnlockCommand(ctx, locker.DeviceID, slot.SlotCode, corrID)

	return &domain.OpenLockerResponse{
		Success: true,
		Action:  "UNLOCKED",
		Message: fmt.Sprintf("Loker %s (%s) berhasil dibuka!", slot.SlotCode, locker.Name),
	}, nil
}

func (s *RentalService) CloseLocker(ctx context.Context, rentalID uuid.UUID, userID uuid.UUID) (*domain.Rental, error) {
	rent, err := s.rentalRepo.GetByID(ctx, rentalID)
	if err != nil {
		return nil, domain.NewAppError(404, "RENTAL_NOT_FOUND", "Rental not found", err)
	}

	if rent.UserID != userID {
		return nil, domain.NewAppError(403, "FORBIDDEN", "Unauthorized access to this rental", domain.ErrForbidden)
	}

	if rent.Status != domain.RentalActive {
		return nil, domain.NewAppError(400, "INVALID_STATE", "Only active rentals can be closed/completed", domain.ErrInvalidRentalState)
	}

	now := time.Now()
	rent.Status = domain.RentalCompleted
	rent.EndedAt = &now

	if err := s.rentalRepo.Update(ctx, rent); err != nil {
		return nil, err
	}

	// Free slot
	_ = s.slotRepo.UpdateStatus(ctx, rent.SlotID, domain.SlotAvailable, nil)

	return rent, nil
}

func (s *RentalService) GetRentalsByUser(ctx context.Context, userID uuid.UUID) ([]domain.Rental, error) {
	return s.rentalRepo.ListByUser(ctx, userID)
}

func (s *RentalService) GetRentalByID(ctx context.Context, rentalID uuid.UUID, userID uuid.UUID) (*domain.Rental, error) {
	rent, err := s.rentalRepo.GetByID(ctx, rentalID)
	if err != nil {
		return nil, err
	}
	if rent.UserID != userID {
		return nil, domain.NewAppError(403, "FORBIDDEN", "Access denied", domain.ErrForbidden)
	}
	return rent, nil
}
