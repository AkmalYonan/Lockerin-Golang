package repository

import (
	"context"
	"time"

	"lockerin-backend/internal/domain"

	"github.com/google/uuid"
)

type ProfileRepository interface {
	Create(ctx context.Context, p *domain.Profile) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Profile, error)
	GetByEmail(ctx context.Context, email string) (*domain.Profile, error)
	GetByUsername(ctx context.Context, username string) (*domain.Profile, error)
	Update(ctx context.Context, p *domain.Profile) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	List(ctx context.Context, limit, offset int) ([]domain.Profile, error)
	Count(ctx context.Context) (int, error)
}

type LocationRepository interface {
	Create(ctx context.Context, loc *domain.Location) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Location, error)
	GetByCode(ctx context.Context, code string) (*domain.Location, error)
	Update(ctx context.Context, loc *domain.Location) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]domain.Location, error)
	Count(ctx context.Context) (int, error)
}

type LockerRepository interface {
	Create(ctx context.Context, locker *domain.Locker) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Locker, error)
	GetByDeviceID(ctx context.Context, deviceID string) (*domain.Locker, error)
	Update(ctx context.Context, locker *domain.Locker) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByLocation(ctx context.Context, locationID uuid.UUID) ([]domain.Locker, error)
	ListAll(ctx context.Context) ([]domain.Locker, error)
	ListWithStats(ctx context.Context) ([]domain.LockerAdminItem, error)
	Count(ctx context.Context) (int, error)
}

type SlotRepository interface {
	Create(ctx context.Context, slot *domain.LockerSlot) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.LockerSlot, error)
	GetByLockerAndCode(ctx context.Context, lockerID uuid.UUID, slotCode string) (*domain.LockerSlot, error)
	Update(ctx context.Context, slot *domain.LockerSlot) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.SlotStatus, currentRentalID *uuid.UUID) error
	ListByLocker(ctx context.Context, lockerID uuid.UUID) ([]domain.LockerSlot, error)
	CountByStatus(ctx context.Context) (total, available, occupied, maintenance int, err error)
	GetPricing(ctx context.Context) ([]domain.PricingTier, error)
	UpdateBasePrices(ctx context.Context, prices map[domain.SlotSize]float64) error
}

type RentalRepository interface {
	Create(ctx context.Context, rental *domain.Rental) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Rental, error)
	GetActiveBySlotID(ctx context.Context, slotID uuid.UUID) (*domain.Rental, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Rental, error)
	ListAll(ctx context.Context, limit, offset int) ([]domain.Rental, error)
	Update(ctx context.Context, rental *domain.Rental) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.RentalStatus) error
	ExpireOverdueReservations(ctx context.Context, now time.Time) ([]domain.Rental, error)
	CountActive(ctx context.Context) (int, error)
	GetTotalRevenue(ctx context.Context) (total float64, today float64, err error)
}

type SecurityCodeRepository interface {
	Create(ctx context.Context, code *domain.SecurityCode) error
	GetByRentalID(ctx context.Context, rentalID uuid.UUID) (*domain.SecurityCode, error)
	IncrementAttempts(ctx context.Context, rentalID uuid.UUID) error
	SetLockout(ctx context.Context, rentalID uuid.UUID, until time.Time) error
	MarkUsed(ctx context.Context, rentalID uuid.UUID, usedAt time.Time) error
	Update(ctx context.Context, code *domain.SecurityCode) error
}

type PaymentRepository interface {
	Create(ctx context.Context, payment *domain.Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error)
	GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error)
	GetByRentalID(ctx context.Context, rentalID uuid.UUID) (*domain.Payment, error)
	Update(ctx context.Context, payment *domain.Payment) error
	ListAll(ctx context.Context, limit, offset int) ([]domain.Payment, error)
	ListAllDetailed(ctx context.Context, limit, offset int) ([]domain.AdminTransactionItem, error)
	GetDetailedByID(ctx context.Context, id uuid.UUID) (*domain.AdminTransactionItem, error)
}

type PricingRepository interface {
	GetTiers(ctx context.Context) ([]domain.PricingTierItem, error)
	UpsertTiers(ctx context.Context, tiers []domain.PricingTierItem) error
}

type DeviceRepository interface {
	CreateCommand(ctx context.Context, cmd *domain.DeviceCommand) error
	GetCommandByID(ctx context.Context, id uuid.UUID) (*domain.DeviceCommand, error)
	GetCommandByCorrelationID(ctx context.Context, corrID string) (*domain.DeviceCommand, error)
	UpdateCommandStatus(ctx context.Context, corrID string, status domain.DeviceCommandStatus, ackTime *time.Time, errMsg string) error
	UpdateLockerHeartbeat(ctx context.Context, deviceID string, firmware string, lastSeen time.Time) error
}

type NotificationRepository interface {
	Create(ctx context.Context, notif *domain.Notification) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Notification, error)
	MarkRead(ctx context.Context, id, userID uuid.UUID) error
	RegisterFCMToken(ctx context.Context, dev *domain.FCMDevice) error
	GetFCMTokensByUser(ctx context.Context, userID uuid.UUID) ([]string, error)
}

type PromoRepository interface {
	Create(ctx context.Context, promo *domain.Promo) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Promo, error)
	GetByCode(ctx context.Context, code string) (*domain.Promo, error)
	Update(ctx context.Context, promo *domain.Promo) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]domain.Promo, error)
}

type AuditRepository interface {
	CreateLog(ctx context.Context, log *domain.AuditLog) error
	ListLogs(ctx context.Context, limit, offset int) ([]domain.AuditLog, error)
}
