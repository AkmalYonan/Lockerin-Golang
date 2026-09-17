package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"lockerin-backend/internal/domain"
	"lockerin-backend/internal/integration/locker_device"
	"lockerin-backend/internal/repository"

	"github.com/google/uuid"
)

type AdminService struct {
	locRepo     repository.LocationRepository
	lockerRepo  repository.LockerRepository
	slotRepo    repository.SlotRepository
	profRepo    repository.ProfileRepository
	rentRepo    repository.RentalRepository
	promoRepo   repository.PromoRepository
	payRepo     repository.PaymentRepository
	devRepo     repository.DeviceRepository
	auditRepo   repository.AuditRepository
	pricingRepo repository.PricingRepository
	deviceDisp  *locker_device.Dispatcher
}

func NewAdminService(
	locRepo repository.LocationRepository,
	lockerRepo repository.LockerRepository,
	slotRepo repository.SlotRepository,
	profRepo repository.ProfileRepository,
	rentRepo repository.RentalRepository,
	promoRepo repository.PromoRepository,
	payRepo repository.PaymentRepository,
	devRepo repository.DeviceRepository,
	auditRepo repository.AuditRepository,
	pricingRepo repository.PricingRepository,
	deviceDisp *locker_device.Dispatcher,
) *AdminService {
	return &AdminService{
		locRepo:     locRepo,
		lockerRepo:  lockerRepo,
		slotRepo:    slotRepo,
		profRepo:    profRepo,
		rentRepo:    rentRepo,
		promoRepo:   promoRepo,
		payRepo:     payRepo,
		devRepo:     devRepo,
		auditRepo:   auditRepo,
		pricingRepo: pricingRepo,
		deviceDisp:  deviceDisp,
	}
}

func (s *AdminService) GetDashboardStats(ctx context.Context) (*domain.AdminDashboardStats, error) {
	totalLoc, _ := s.locRepo.Count(ctx)
	totalLocker, _ := s.lockerRepo.Count(ctx)
	totalSlots, avail, occupied, maintenance, _ := s.slotRepo.CountByStatus(ctx)
	totalUsers, _ := s.profRepo.Count(ctx)
	activeRentals, _ := s.rentRepo.CountActive(ctx)
	totalRev, todayRev, _ := s.rentRepo.GetTotalRevenue(ctx)

	return &domain.AdminDashboardStats{
		TotalLocations:   totalLoc,
		TotalLockers:     totalLocker,
		TotalSlots:       totalSlots,
		AvailableSlots:   avail,
		OccupiedSlots:    occupied,
		MaintenanceSlots: maintenance,
		ActiveRentals:    activeRentals,
		TotalUsers:       totalUsers,
		TotalRevenue:     totalRev,
		TodayRevenue:     todayRev,
	}, nil
}

func (s *AdminService) ListLockersWithStats(ctx context.Context) ([]domain.LockerAdminItem, error) {
	return s.lockerRepo.ListWithStats(ctx)
}

func (s *AdminService) ListUsers(ctx context.Context, limit, offset int) ([]domain.Profile, error) {
	return s.profRepo.List(ctx, limit, offset)
}

func (s *AdminService) GetUserDetail(ctx context.Context, id uuid.UUID) (*domain.AdminUserDetailResponse, error) {
	prof, err := s.profRepo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.NewAppError(404, "USER_NOT_FOUND", "User profile not found", err)
	}

	rentals, _ := s.rentRepo.ListByUser(ctx, id)

	var payments []domain.Payment
	for _, rent := range rentals {
		if p, err := s.payRepo.GetByRentalID(ctx, rent.ID); err == nil && p != nil {
			payments = append(payments, *p)
		}
	}

	return &domain.AdminUserDetailResponse{
		Profile:  prof,
		Rentals:  rentals,
		Payments: payments,
	}, nil
}

func (s *AdminService) UpdateUserStatus(ctx context.Context, actorID string, id uuid.UUID, status string) (*domain.Profile, error) {
	status = strings.ToLower(status)
	if status != "active" && status != "suspended" && status != "deactivated" {
		return nil, domain.NewAppError(400, "INVALID_STATUS", "Status must be active or suspended", nil)
	}
	if err := s.profRepo.UpdateStatus(ctx, id, status); err != nil {
		return nil, err
	}
	p, err := s.profRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Audit Log
	_ = s.auditRepo.CreateLog(ctx, &domain.AuditLog{
		ID:         uuid.New(),
		ActorType:  "admin",
		ActorID:    actorID,
		Action:     "USER_STATUS_UPDATED",
		EntityType: "user",
		EntityID:   id.String(),
		Metadata:   []byte(fmt.Sprintf(`{"status": "%s"}`, status)),
		CreatedAt:  time.Now(),
	})

	return p, nil
}

func (s *AdminService) GetPricing(ctx context.Context) ([]domain.PricingTierItem, error) {
	if s.pricingRepo != nil {
		return s.pricingRepo.GetTiers(ctx)
	}
	return []domain.PricingTierItem{
		{SlotSize: "small", HourlyRate: 4000, DepositAmount: 10000},
		{SlotSize: "medium", HourlyRate: 6000, DepositAmount: 15000},
		{SlotSize: "large", HourlyRate: 9000, DepositAmount: 20000},
	}, nil
}

func (s *AdminService) UpdatePricing(ctx context.Context, actorID string, tiers []domain.PricingTierItem) error {
	if len(tiers) == 0 {
		return domain.NewAppError(400, "EMPTY_PRICING", "No pricing tiers provided", nil)
	}

	if s.pricingRepo != nil {
		if err := s.pricingRepo.UpsertTiers(ctx, tiers); err != nil {
			return err
		}
	}

	// Sync prices in slotRepo base_price_per_hour
	priceMap := make(map[domain.SlotSize]float64)
	for _, t := range tiers {
		switch strings.ToLower(t.SlotSize) {
		case "small":
			priceMap[domain.SizeSmall] = t.HourlyRate
		case "medium":
			priceMap[domain.SizeMedium] = t.HourlyRate
		case "large":
			priceMap[domain.SizeLarge] = t.HourlyRate
		}
	}
	_ = s.slotRepo.UpdateBasePrices(ctx, priceMap)

	// Audit Log
	b, _ := json.Marshal(tiers)
	_ = s.auditRepo.CreateLog(ctx, &domain.AuditLog{
		ID:         uuid.New(),
		ActorType:  "admin",
		ActorID:    actorID,
		Action:     "PRICING_UPDATED",
		EntityType: "pricing",
		Metadata:   b,
		CreatedAt:  time.Now(),
	})

	return nil
}

func (s *AdminService) ListTransactions(ctx context.Context, limit, offset int) ([]domain.AdminTransactionItem, error) {
	return s.payRepo.ListAllDetailed(ctx, limit, offset)
}

func (s *AdminService) GetTransactionDetail(ctx context.Context, id uuid.UUID) (*domain.AdminTransactionItem, error) {
	return s.payRepo.GetDetailedByID(ctx, id)
}

func (s *AdminService) ListAuditLogs(ctx context.Context, limit, offset int) ([]domain.AuditLog, error) {
	return s.auditRepo.ListLogs(ctx, limit, offset)
}

func (s *AdminService) CreateLocation(ctx context.Context, loc *domain.Location) error {
	if loc.ID == uuid.Nil {
		loc.ID = uuid.New()
	}
	if loc.Code == "" {
		codeName := strings.ToUpper(strings.ReplaceAll(loc.Name, " ", "-"))
		if len(codeName) > 20 {
			codeName = codeName[:20]
		}
		loc.Code = "LOC-" + codeName
	}
	if loc.Status == "" {
		loc.Status = "active"
	}
	loc.CreatedAt = time.Now()
	loc.UpdatedAt = time.Now()

	// 1. Insert Location
	if err := s.locRepo.Create(ctx, loc); err != nil {
		return err
	}

	// 2. Insert Default Unit Loker
	cleanCode := strings.TrimPrefix(loc.Code, "LOC-")
	if len(cleanCode) > 15 {
		cleanCode = cleanCode[:15]
	}
	lockerCode := fmt.Sprintf("LCK-%s-01", cleanCode)
	deviceID := fmt.Sprintf("HEAD-LOCKER-%s-001", cleanCode)

	locker := &domain.Locker{
		ID:              uuid.New(),
		LocationID:      loc.ID,
		Code:            lockerCode,
		Name:            loc.Name + " Unit 1",
		DeviceID:        deviceID,
		Status:          domain.LockerOnline,
		FirmwareVersion: "2.1.0",
		LastSeenAt:      time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if err := s.lockerRepo.Create(ctx, locker); err != nil {
		return err
	}

	// 3. Insert 25 Slots Matrix (A1 to E5)
	rows := []string{"A", "B", "C", "D", "E"}
	for _, r := range rows {
		for c := 1; c <= 5; c++ {
			slotCode := fmt.Sprintf("%s%d", r, c)
			var size domain.SlotSize
			var price float64

			switch r {
			case "A", "B":
				size = domain.SizeSmall
				price = 4000.00
			case "C", "D":
				size = domain.SizeMedium
				price = 6000.00
			case "E":
				size = domain.SizeLarge
				price = 9000.00
			}

			slot := &domain.LockerSlot{
				ID:               uuid.New(),
				LockerID:         locker.ID,
				SlotCode:         slotCode,
				Size:             size,
				Status:           domain.SlotAvailable,
				BasePricePerHour: price,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}
			_ = s.slotRepo.Create(ctx, slot)
		}
	}

	return nil
}

func (s *AdminService) UpdateLocation(ctx context.Context, loc *domain.Location) error {
	return s.locRepo.Update(ctx, loc)
}

func (s *AdminService) UpdateLocationStatus(ctx context.Context, actorID string, id uuid.UUID, status string) error {
	status = strings.ToLower(status)
	if status != "active" && status != "inactive" && status != "maintenance" {
		return domain.NewAppError(400, "INVALID_STATUS", "Status must be active, inactive, or maintenance", nil)
	}
	if err := s.locRepo.UpdateStatus(ctx, id, status); err != nil {
		return err
	}

	_ = s.auditRepo.CreateLog(ctx, &domain.AuditLog{
		ID:         uuid.New(),
		ActorType:  "admin",
		ActorID:    actorID,
		Action:     "LOCATION_STATUS_UPDATED",
		EntityType: "location",
		EntityID:   id.String(),
		Metadata:   []byte(fmt.Sprintf(`{"status": "%s"}`, status)),
		CreatedAt:  time.Now(),
	})
	return nil
}

func (s *AdminService) DeleteLocation(ctx context.Context, id uuid.UUID) error {
	return s.locRepo.Delete(ctx, id)
}

func (s *AdminService) CreatePromo(ctx context.Context, req *domain.CreatePromoRequest) (*domain.Promo, error) {
	if req.Code == "" {
		return nil, domain.NewAppError(400, "MISSING_CODE", "Promo code is required", nil)
	}
	if req.Title == "" {
		req.Title = "Voucher " + req.Code
	}

	discType := strings.ToLower(req.DiscountType)
	if discType != "percentage" && discType != "fixed" {
		discType = "percentage"
	}

	discVal := req.DiscountValue
	if discVal <= 0 && req.DiscountPercent > 0 {
		discVal = req.DiscountPercent
	}
	if discVal <= 0 {
		return nil, domain.NewAppError(400, "INVALID_DISCOUNT", "Discount value must be greater than 0", nil)
	}

	maxDisc := req.MaxDiscountAmount
	if maxDisc == nil && req.MaxDiscount != nil {
		maxDisc = req.MaxDiscount
	}

	startsAt := req.StartsAt
	if startsAt.IsZero() {
		startsAt = time.Now()
	}

	endsAt := req.EndsAt
	if endsAt.IsZero() && !req.ValidUntil.IsZero() {
		endsAt = req.ValidUntil
	}
	if endsAt.IsZero() {
		endsAt = startsAt.Add(30 * 24 * time.Hour)
	}

	status := strings.ToLower(req.Status)
	if status != "active" && status != "inactive" && status != "expired" {
		status = "active"
	}

	promo := &domain.Promo{
		ID:                uuid.New(),
		Code:              strings.ToUpper(req.Code),
		Title:             req.Title,
		Description:       req.Description,
		DiscountType:      discType,
		DiscountValue:     discVal,
		MinOrderAmount:    req.MinOrderAmount,
		MaxDiscountAmount: maxDisc,
		StartsAt:          startsAt,
		EndsAt:            endsAt,
		Status:            status,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.promoRepo.Create(ctx, promo); err != nil {
		return nil, err
	}
	return promo, nil
}

func (s *AdminService) ListPromos(ctx context.Context) ([]domain.Promo, error) {
	return s.promoRepo.List(ctx)
}

func (s *AdminService) UpdatePromoStatus(ctx context.Context, id uuid.UUID, req *domain.UpdatePromoStatusRequest) error {
	status := strings.ToLower(req.Status)
	if status == "" && req.IsActive != nil {
		if *req.IsActive {
			status = "active"
		} else {
			status = "inactive"
		}
	}
	if status != "active" && status != "inactive" && status != "expired" {
		return domain.NewAppError(400, "INVALID_STATUS", "Status must be active, inactive, or expired", nil)
	}
	return s.promoRepo.UpdateStatus(ctx, id, status)
}

func (s *AdminService) DispatchAdminCommand(ctx context.Context, actorID string, req *domain.AdminDeviceCommandRequest) (*domain.DeviceCommand, error) {
	locker, err := s.lockerRepo.GetByDeviceID(ctx, req.DeviceID)
	if err != nil {
		return nil, domain.NewAppError(404, "DEVICE_NOT_FOUND", "Locker hardware unit not found", err)
	}

	slotCode, _ := req.Payload["slot_code"].(string)
	var slotID *uuid.UUID
	if slotCode != "" {
		slot, err := s.slotRepo.GetByLockerAndCode(ctx, locker.ID, slotCode)
		if err == nil && slot != nil {
			slotID = &slot.ID
		}
	}

	corrID := fmt.Sprintf("admin-override-%s-%d", req.DeviceID, time.Now().Unix())
	payloadBytes, _ := json.Marshal(req.Payload)

	action := domain.CommandOpenSlot
	if req.CommandType == "REBOOT" {
		action = domain.CommandReboot
	} else if req.CommandType == "STATUS_CHECK" || req.CommandType == "SYNC_STATUS" {
		action = domain.CommandStatusCheck
	}

	cmd := &domain.DeviceCommand{
		ID:            uuid.New(),
		LockerID:      locker.ID,
		SlotID:        slotID,
		Command:       action,
		CorrelationID: corrID,
		Payload:       payloadBytes,
		Status:        domain.CommandSent,
		RequestedAt:   time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if s.devRepo != nil {
		_ = s.devRepo.CreateCommand(ctx, cmd)
	}

	// Create audit log for emergency override
	auditLog := &domain.AuditLog{
		ID:         uuid.New(),
		ActorType:  "admin",
		ActorID:    actorID,
		Action:     fmt.Sprintf("REMOTE_COMMAND_%s", req.CommandType),
		EntityType: "device",
		EntityID:   req.DeviceID,
		Metadata:   payloadBytes,
		CreatedAt:  time.Now(),
	}
	_ = s.auditRepo.CreateLog(ctx, auditLog)

	return cmd, nil
}
