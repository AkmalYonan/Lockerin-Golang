package inmemory

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"lockerin-backend/internal/domain"
	"lockerin-backend/internal/repository"

	"github.com/google/uuid"
)

type Store struct {
	mu            sync.RWMutex
	Profiles      map[uuid.UUID]domain.Profile
	Locations     map[uuid.UUID]domain.Location
	Lockers       map[uuid.UUID]domain.Locker
	Slots         map[uuid.UUID]domain.LockerSlot
	Rentals       map[uuid.UUID]domain.Rental
	SecurityCodes map[uuid.UUID]domain.SecurityCode // keyed by rental_id
	Payments      map[uuid.UUID]domain.Payment
	Commands      map[uuid.UUID]domain.DeviceCommand
	Notifications []domain.Notification
	FCMDevices    map[string]domain.FCMDevice // keyed by token
	Promos        map[string]domain.Promo     // keyed by code
	AuditLogs     []domain.AuditLog
	PricingTiers  map[string]domain.PricingTierItem
}

func NewInMemoryStore() *Store {
	s := &Store{
		Profiles:      make(map[uuid.UUID]domain.Profile),
		Locations:     make(map[uuid.UUID]domain.Location),
		Lockers:       make(map[uuid.UUID]domain.Locker),
		Slots:         make(map[uuid.UUID]domain.LockerSlot),
		Rentals:       make(map[uuid.UUID]domain.Rental),
		SecurityCodes: make(map[uuid.UUID]domain.SecurityCode),
		Payments:      make(map[uuid.UUID]domain.Payment),
		Commands:      make(map[uuid.UUID]domain.DeviceCommand),
		Notifications: []domain.Notification{},
		FCMDevices:    make(map[string]domain.FCMDevice),
		Promos:        make(map[string]domain.Promo),
		AuditLogs:     []domain.AuditLog{},
		PricingTiers:  make(map[string]domain.PricingTierItem),
	}
	s.seedDefaultData()
	return s
}

func (s *Store) seedDefaultData() {
	// Seed Admin User
	adminID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	s.Profiles[adminID] = domain.Profile{
		ID:        adminID,
		Email:     "admin@lockerin.id",
		Username:  "admin",
		Name:      "Super Admin Lockerin",
		Phone:     "081299998888",
		Role:      domain.RoleAdmin,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Seed Standard User
	userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	s.Profiles[userID] = domain.Profile{
		ID:        userID,
		Email:     "akmal@lockerin.id",
		Username:  "akmal",
		Name:      "Akmal Maindata",
		Phone:     "081234567890",
		Role:      domain.RoleUser,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Seed Locations
	loc1ID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	s.Locations[loc1ID] = domain.Location{
		ID:             loc1ID,
		Code:           "LOC-MEDISTRA",
		Name:           "Lockerin RS Medistra",
		Address:        "Jl. Jend. Gatot Subroto No.Kav. 59, Jakarta Selatan",
		City:           "Jakarta Selatan",
		Latitude:       -6.2372,
		Longitude:      106.8335,
		Status:         "active",
		OperatingHours: []byte(`{"open": "06:00", "close": "23:00", "is_24_hours": false}`),
		ImageURL:       "https://images.unsplash.com/photo-1590381105924-c72589b9ef3f?w=600",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	loc2ID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	s.Locations[loc2ID] = domain.Location{
		ID:             loc2ID,
		Code:           "LOC-PARAMADINA",
		Name:           "Lockerin Univ. Paramadina",
		Address:        "Jl. Gatot Subroto No.Kav. 97, Mampang Prapatan, Jakarta Selatan",
		City:           "Jakarta Selatan",
		Latitude:       -6.2415,
		Longitude:      106.8329,
		Status:         "active",
		OperatingHours: []byte(`{"open": "07:00", "close": "22:00", "is_24_hours": false}`),
		ImageURL:       "https://images.unsplash.com/photo-1541829070764-84a7d30dd3f3?w=600",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Seed Lockers and Slots (A1 to E5 for each location)
	l1ID := uuid.MustParse("c1111111-1111-1111-1111-111111111111")
	s.Lockers[l1ID] = domain.Locker{
		ID:              l1ID,
		LocationID:      loc1ID,
		Code:            "LCK-MEDISTRA-01",
		Name:            "Lockerin RS Medistra Unit 1",
		DeviceID:        "HEAD-LOCKER-MDS-001",
		Status:          domain.LockerOnline,
		FirmwareVersion: "2.1.0",
		LastSeenAt:      time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	rows := []string{"A", "B", "C", "D", "E"}
	for _, r := range rows {
		for c := 1; c <= 5; c++ {
			slotID := uuid.New()
			slotCode := fmt.Sprintf("%s%d", r, c)
			size := domain.SizeMedium
			price := 6000.00
			if c == 1 {
				size = domain.SizeSmall
				price = 4000.00
			} else if c == 5 {
				size = domain.SizeLarge
				price = 9000.00
			}
			s.Slots[slotID] = domain.LockerSlot{
				ID:               slotID,
				LockerID:         l1ID,
				SlotCode:         slotCode,
				Size:             size,
				Status:           domain.SlotAvailable,
				BasePricePerHour: price,
				LockerCode:       "LCK-MEDISTRA-01",
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}
		}
	}

	// Seed Promo
	promoID := uuid.New()
	s.Promos["LOCKERIN50"] = domain.Promo{
		ID:             promoID,
		Code:           "LOCKERIN50",
		Title:          "Promo Pengguna Baru 50%",
		Description:    "Diskon 50% hingga Rp 10.000 untuk pengguna baru",
		DiscountType:   "percentage",
		DiscountValue:  50.0,
		MinOrderAmount: 5000.0,
		StartsAt:       time.Now().Add(-24 * time.Hour),
		EndsAt:         time.Now().Add(30 * 24 * time.Hour),
		Status:         "active",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// Repositories Implementations

type InMemoryProfileRepo struct{ s *Store }

func (r *InMemoryProfileRepo) Create(ctx context.Context, p *domain.Profile) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Profiles[p.ID] = *p
	return nil
}
func (r *InMemoryProfileRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Profile, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	if p, ok := r.s.Profiles[id]; ok {
		return &p, nil
	}
	return nil, domain.ErrNotFound
}
func (r *InMemoryProfileRepo) GetByEmail(ctx context.Context, email string) (*domain.Profile, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, p := range r.s.Profiles {
		if strings.EqualFold(p.Email, email) {
			return &p, nil
		}
	}
	return nil, domain.ErrNotFound
}
func (r *InMemoryProfileRepo) GetByUsername(ctx context.Context, username string) (*domain.Profile, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, p := range r.s.Profiles {
		if strings.EqualFold(p.Username, username) {
			return &p, nil
		}
	}
	return nil, domain.ErrNotFound
}
func (r *InMemoryProfileRepo) Update(ctx context.Context, p *domain.Profile) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.Profiles[p.ID]; !ok {
		return domain.ErrNotFound
	}
	r.s.Profiles[p.ID] = *p
	return nil
}
func (r *InMemoryProfileRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	p, ok := r.s.Profiles[id]
	if !ok {
		return domain.ErrNotFound
	}
	p.Status = status
	p.UpdatedAt = time.Now()
	r.s.Profiles[id] = p
	return nil
}
func (r *InMemoryProfileRepo) List(ctx context.Context, limit, offset int) ([]domain.Profile, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var list []domain.Profile
	for _, p := range r.s.Profiles {
		list = append(list, p)
	}
	return list, nil
}
func (r *InMemoryProfileRepo) Count(ctx context.Context) (int, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	return len(r.s.Profiles), nil
}

type InMemoryLocationRepo struct{ s *Store }

func (r *InMemoryLocationRepo) Create(ctx context.Context, loc *domain.Location) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Locations[loc.ID] = *loc
	return nil
}
func (r *InMemoryLocationRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Location, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	if loc, ok := r.s.Locations[id]; ok {
		return &loc, nil
	}
	return nil, domain.ErrNotFound
}
func (r *InMemoryLocationRepo) GetByCode(ctx context.Context, code string) (*domain.Location, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, loc := range r.s.Locations {
		if strings.EqualFold(loc.Code, code) {
			return &loc, nil
		}
	}
	return nil, domain.ErrNotFound
}
func (r *InMemoryLocationRepo) Update(ctx context.Context, loc *domain.Location) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Locations[loc.ID] = *loc
	return nil
}
func (r *InMemoryLocationRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if loc, ok := r.s.Locations[id]; ok {
		loc.Status = status
		loc.UpdatedAt = time.Now()
		r.s.Locations[id] = loc
		return nil
	}
	return domain.ErrNotFound
}
func (r *InMemoryLocationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	delete(r.s.Locations, id)
	return nil
}
func (r *InMemoryLocationRepo) List(ctx context.Context, limit, offset int) ([]domain.Location, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var list []domain.Location
	for _, loc := range r.s.Locations {
		// Calculate available slots
		avail := 0
		occupied := 0
		totalLockers := 0
		for _, lk := range r.s.Lockers {
			if lk.LocationID == loc.ID {
				totalLockers++
				for _, slot := range r.s.Slots {
					if slot.LockerID == lk.ID {
						if slot.Status == domain.SlotAvailable {
							avail++
						} else {
							occupied++
						}
					}
				}
			}
		}
		loc.TotalLockers = totalLockers
		loc.AvailableSlots = avail
		loc.OccupiedSlots = occupied
		list = append(list, loc)
	}
	return list, nil
}
func (r *InMemoryLocationRepo) Count(ctx context.Context) (int, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	return len(r.s.Locations), nil
}

type InMemoryLockerRepo struct{ s *Store }

func (r *InMemoryLockerRepo) Create(ctx context.Context, l *domain.Locker) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Lockers[l.ID] = *l
	return nil
}
func (r *InMemoryLockerRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Locker, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	if l, ok := r.s.Lockers[id]; ok {
		return &l, nil
	}
	return nil, domain.ErrNotFound
}
func (r *InMemoryLockerRepo) GetByDeviceID(ctx context.Context, deviceID string) (*domain.Locker, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, l := range r.s.Lockers {
		if strings.EqualFold(l.DeviceID, deviceID) {
			return &l, nil
		}
	}
	return nil, domain.ErrNotFound
}
func (r *InMemoryLockerRepo) Update(ctx context.Context, l *domain.Locker) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Lockers[l.ID] = *l
	return nil
}
func (r *InMemoryLockerRepo) Delete(ctx context.Context, id uuid.UUID) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	delete(r.s.Lockers, id)
	return nil
}
func (r *InMemoryLockerRepo) ListByLocation(ctx context.Context, locationID uuid.UUID) ([]domain.Locker, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var list []domain.Locker
	for _, l := range r.s.Lockers {
		if l.LocationID == locationID {
			list = append(list, l)
		}
	}
	return list, nil
}
func (r *InMemoryLockerRepo) ListAll(ctx context.Context) ([]domain.Locker, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var list []domain.Locker
	for _, l := range r.s.Lockers {
		list = append(list, l)
	}
	return list, nil
}
func (r *InMemoryLockerRepo) ListWithStats(ctx context.Context) ([]domain.LockerAdminItem, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var list []domain.LockerAdminItem
	for _, l := range r.s.Lockers {
		locName := ""
		if loc, ok := r.s.Locations[l.LocationID]; ok {
			locName = loc.Name
		}
		totalSlots := 0
		availSlots := 0
		occupiedSlots := 0
		for _, s := range r.s.Slots {
			if s.LockerID == l.ID {
				totalSlots++
				if s.Status == domain.SlotAvailable {
					availSlots++
				} else if s.Status == domain.SlotOccupied || s.Status == domain.SlotReserved {
					occupiedSlots++
				}
			}
		}
		list = append(list, domain.LockerAdminItem{
			ID:              l.ID,
			DeviceID:        l.DeviceID,
			Code:            l.Code,
			Name:            l.Name,
			LocationID:      l.LocationID,
			LocationName:    locName,
			Status:          l.Status,
			FirmwareVersion: l.FirmwareVersion,
			BatteryLevel:    95,
			SignalStrength:  -65,
			LastHeartbeat:   l.LastSeenAt,
			TotalSlots:      totalSlots,
			AvailableSlots:  availSlots,
			OccupiedSlots:   occupiedSlots,
		})
	}
	return list, nil
}
func (r *InMemoryLockerRepo) Count(ctx context.Context) (int, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	return len(r.s.Lockers), nil
}

type InMemorySlotRepo struct{ s *Store }

func (r *InMemorySlotRepo) Create(ctx context.Context, slot *domain.LockerSlot) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Slots[slot.ID] = *slot
	return nil
}
func (r *InMemorySlotRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.LockerSlot, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	if slot, ok := r.s.Slots[id]; ok {
		if lk, ok2 := r.s.Lockers[slot.LockerID]; ok2 {
			slot.LockerCode = lk.Code
		}
		return &slot, nil
	}
	return nil, domain.ErrNotFound
}
func (r *InMemorySlotRepo) GetByLockerAndCode(ctx context.Context, lockerID uuid.UUID, slotCode string) (*domain.LockerSlot, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, slot := range r.s.Slots {
		if slot.LockerID == lockerID && strings.EqualFold(slot.SlotCode, slotCode) {
			return &slot, nil
		}
	}
	return nil, domain.ErrNotFound
}
func (r *InMemorySlotRepo) Update(ctx context.Context, slot *domain.LockerSlot) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Slots[slot.ID] = *slot
	return nil
}
func (r *InMemorySlotRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.SlotStatus, currentRentalID *uuid.UUID) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	slot, ok := r.s.Slots[id]
	if !ok {
		return domain.ErrNotFound
	}
	slot.Status = status
	slot.CurrentRentalID = currentRentalID
	slot.UpdatedAt = time.Now()
	r.s.Slots[id] = slot
	return nil
}
func (r *InMemorySlotRepo) ListByLocker(ctx context.Context, lockerID uuid.UUID) ([]domain.LockerSlot, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var list []domain.LockerSlot
	for _, slot := range r.s.Slots {
		if slot.LockerID == lockerID {
			list = append(list, slot)
		}
	}
	return list, nil
}
func (r *InMemorySlotRepo) CountByStatus(ctx context.Context) (total, available, occupied, maintenance int, err error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	total = len(r.s.Slots)
	for _, s := range r.s.Slots {
		switch s.Status {
		case domain.SlotAvailable:
			available++
		case domain.SlotOccupied, domain.SlotReserved:
			occupied++
		case domain.SlotMaintenance, domain.SlotOffline:
			maintenance++
		}
	}
	return
}
func (r *InMemorySlotRepo) GetPricing(ctx context.Context) ([]domain.PricingTier, error) {
	return []domain.PricingTier{
		{Size: domain.SizeSmall, PricePerHour: 4000, Description: "Cocok untuk gadget & dompet", RecommendedUsage: "Smartphone, dompet"},
		{Size: domain.SizeMedium, PricePerHour: 6000, Description: "Cocok untuk tas ransel & barang sehari-hari", RecommendedUsage: "Tas ransel, laptop, jaket"},
		{Size: domain.SizeLarge, PricePerHour: 9000, Description: "Cocok untuk koper kabin & koper besar", RecommendedUsage: "Koper, tas belanja banyak"},
	}, nil
}
func (r *InMemorySlotRepo) UpdateBasePrices(ctx context.Context, prices map[domain.SlotSize]float64) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for id, s := range r.s.Slots {
		if p, ok := prices[s.Size]; ok && p > 0 {
			s.BasePricePerHour = p
			s.UpdatedAt = time.Now()
			r.s.Slots[id] = s
		}
	}
	return nil
}

type InMemoryRentalRepo struct{ s *Store }

func (r *InMemoryRentalRepo) Create(ctx context.Context, rent *domain.Rental) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()

	// Active slot concurrency check (mimicking partial unique index)
	for _, ex := range r.s.Rentals {
		if ex.SlotID == rent.SlotID &&
			(ex.Status == domain.RentalReserved || ex.Status == domain.RentalAwaitingPayment ||
				ex.Status == domain.RentalPaid || ex.Status == domain.RentalActive) {
			return domain.ErrSlotUnavailable
		}
	}

	r.s.Rentals[rent.ID] = *rent
	return nil
}

func (r *InMemoryRentalRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Rental, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	if rent, ok := r.s.Rentals[id]; ok {
		if u, ok2 := r.s.Profiles[rent.UserID]; ok2 {
			rent.UserName = u.Name
		}
		if loc, ok3 := r.s.Locations[rent.LocationID]; ok3 {
			rent.LocationName = loc.Name
		}
		if lk, ok4 := r.s.Lockers[rent.LockerID]; ok4 {
			rent.LockerCode = lk.Code
		}
		if slot, ok5 := r.s.Slots[rent.SlotID]; ok5 {
			rent.SlotCode = slot.SlotCode
		}
		return &rent, nil
	}
	return nil, domain.ErrNotFound
}

func (r *InMemoryRentalRepo) GetActiveBySlotID(ctx context.Context, slotID uuid.UUID) (*domain.Rental, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, rent := range r.s.Rentals {
		if rent.SlotID == slotID && (rent.Status == domain.RentalReserved || rent.Status == domain.RentalAwaitingPayment || rent.Status == domain.RentalPaid || rent.Status == domain.RentalActive) {
			return &rent, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *InMemoryRentalRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Rental, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var list []domain.Rental
	for _, rent := range r.s.Rentals {
		if rent.UserID == userID {
			list = append(list, rent)
		}
	}
	return list, nil
}

func (r *InMemoryRentalRepo) ListAll(ctx context.Context, limit, offset int) ([]domain.Rental, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var list []domain.Rental
	for _, rent := range r.s.Rentals {
		list = append(list, rent)
	}
	return list, nil
}

func (r *InMemoryRentalRepo) Update(ctx context.Context, rent *domain.Rental) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Rentals[rent.ID] = *rent
	return nil
}

func (r *InMemoryRentalRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.RentalStatus) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	rent, ok := r.s.Rentals[id]
	if !ok {
		return domain.ErrNotFound
	}
	rent.Status = status
	rent.UpdatedAt = time.Now()
	r.s.Rentals[id] = rent
	return nil
}

func (r *InMemoryRentalRepo) ExpireOverdueReservations(ctx context.Context, now time.Time) ([]domain.Rental, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	var expired []domain.Rental
	for id, rent := range r.s.Rentals {
		if rent.Status == domain.RentalReserved && rent.ReservationExpiresAt != nil && rent.ReservationExpiresAt.Before(now) {
			rent.Status = domain.RentalExpired
			rent.UpdatedAt = now
			r.s.Rentals[id] = rent
			expired = append(expired, rent)

			// Free the slot
			if slot, ok := r.s.Slots[rent.SlotID]; ok {
				slot.Status = domain.SlotAvailable
				slot.CurrentRentalID = nil
				r.s.Slots[rent.SlotID] = slot
			}
		}
	}
	return expired, nil
}

func (r *InMemoryRentalRepo) CountActive(ctx context.Context) (int, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	count := 0
	for _, rent := range r.s.Rentals {
		if rent.Status == domain.RentalActive {
			count++
		}
	}
	return count, nil
}

func (r *InMemoryRentalRepo) GetTotalRevenue(ctx context.Context) (total float64, today float64, err error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	todayStart := time.Now().Truncate(24 * time.Hour)
	for _, rent := range r.s.Rentals {
		if rent.Status == domain.RentalPaid || rent.Status == domain.RentalActive || rent.Status == domain.RentalCompleted {
			total += rent.TotalAmount
			if rent.CreatedAt.After(todayStart) {
				today += rent.TotalAmount
			}
		}
	}
	return
}

type InMemorySecurityCodeRepo struct{ s *Store }

func (r *InMemorySecurityCodeRepo) Create(ctx context.Context, c *domain.SecurityCode) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.SecurityCodes[c.RentalID] = *c
	return nil
}
func (r *InMemorySecurityCodeRepo) GetByRentalID(ctx context.Context, rentalID uuid.UUID) (*domain.SecurityCode, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	if c, ok := r.s.SecurityCodes[rentalID]; ok {
		return &c, nil
	}
	return nil, domain.ErrNotFound
}
func (r *InMemorySecurityCodeRepo) IncrementAttempts(ctx context.Context, rentalID uuid.UUID) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if c, ok := r.s.SecurityCodes[rentalID]; ok {
		c.Attempts++
		c.UpdatedAt = time.Now()
		r.s.SecurityCodes[rentalID] = c
		return nil
	}
	return domain.ErrNotFound
}
func (r *InMemorySecurityCodeRepo) SetLockout(ctx context.Context, rentalID uuid.UUID, until time.Time) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if c, ok := r.s.SecurityCodes[rentalID]; ok {
		c.LockedUntil = &until
		c.UpdatedAt = time.Now()
		r.s.SecurityCodes[rentalID] = c
		return nil
	}
	return domain.ErrNotFound
}
func (r *InMemorySecurityCodeRepo) MarkUsed(ctx context.Context, rentalID uuid.UUID, usedAt time.Time) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if c, ok := r.s.SecurityCodes[rentalID]; ok {
		c.UsedAt = &usedAt
		c.UpdatedAt = time.Now()
		r.s.SecurityCodes[rentalID] = c
		return nil
	}
	return domain.ErrNotFound
}
func (r *InMemorySecurityCodeRepo) Update(ctx context.Context, c *domain.SecurityCode) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.SecurityCodes[c.RentalID] = *c
	return nil
}

type InMemoryPaymentRepo struct{ s *Store }

func (r *InMemoryPaymentRepo) Create(ctx context.Context, p *domain.Payment) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Payments[p.ID] = *p
	return nil
}
func (r *InMemoryPaymentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	if p, ok := r.s.Payments[id]; ok {
		return &p, nil
	}
	return nil, domain.ErrNotFound
}
func (r *InMemoryPaymentRepo) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, p := range r.s.Payments {
		if strings.EqualFold(p.OrderID, orderID) {
			return &p, nil
		}
	}
	return nil, domain.ErrNotFound
}
func (r *InMemoryPaymentRepo) GetByRentalID(ctx context.Context, rentalID uuid.UUID) (*domain.Payment, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, p := range r.s.Payments {
		if p.RentalID == rentalID {
			return &p, nil
		}
	}
	return nil, domain.ErrNotFound
}
func (r *InMemoryPaymentRepo) Update(ctx context.Context, p *domain.Payment) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Payments[p.ID] = *p
	return nil
}
func (r *InMemoryPaymentRepo) ListAll(ctx context.Context, limit, offset int) ([]domain.Payment, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var list []domain.Payment
	for _, p := range r.s.Payments {
		list = append(list, p)
	}
	return list, nil
}

func (r *InMemoryPaymentRepo) ListAllDetailed(ctx context.Context, limit, offset int) ([]domain.AdminTransactionItem, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var list []domain.AdminTransactionItem
	for _, p := range r.s.Payments {
		userName := "Budi Prakoso"
		userEmail := "budi@lockerin.id"
		userPhone := "081298765432"
		locName := "Lockerin RS Medistra"
		locAddr := "Jl. Jend. Gatot Subroto No. 59"
		lockerCode := "HEAD-LOCKER-MDS-001"
		slotCode := "A2"
		slotSize := "Small"

		if rent, ok := r.s.Rentals[p.RentalID]; ok {
			if u, ok2 := r.s.Profiles[rent.UserID]; ok2 {
				userName = u.Name
				userEmail = u.Email
				userPhone = u.Phone
			}
			if loc, ok3 := r.s.Locations[rent.LocationID]; ok3 {
				locName = loc.Name
				locAddr = loc.Address
			}
			if lk, ok4 := r.s.Lockers[rent.LockerID]; ok4 {
				lockerCode = lk.DeviceID
			}
			if sl, ok5 := r.s.Slots[rent.SlotID]; ok5 {
				slotCode = sl.SlotCode
				slotSize = string(sl.Size)
			}
		}

		pTimeStr := p.CreatedAt.Format("2006-01-02 15:04:05")
		events := []domain.AdminAuditEvent{
			{Step: "Slot Reserved", Time: pTimeStr, Detail: fmt.Sprintf("TTL 15 Menit terkunci untuk User (%s)", userName)},
			{Step: "Payment Created", Time: pTimeStr, Detail: fmt.Sprintf("Snap %s token generated", strings.ToUpper(p.PaymentType))},
		}
		if p.PaidAt != nil {
			sTimeStr := p.PaidAt.Format("2006-01-02 15:04:05")
			events = append(events, domain.AdminAuditEvent{
				Step:   "Payment Settled",
				Time:   sTimeStr,
				Detail: "Webhook Midtrans berhasil diverifikasi, PIN Solenoid aktif",
			})
		}

		list = append(list, domain.AdminTransactionItem{
			ID:              p.ID,
			OrderID:         p.OrderID,
			Amount:          p.Amount,
			PaymentMethod:   p.PaymentType,
			Status:          string(p.Status),
			MidtransID:      p.ProviderTransactionID,
			SnapToken:       p.SnapToken,
			CreatedAt:       p.CreatedAt,
			SettledAt:       p.PaidAt,
			UserName:        userName,
			UserEmail:       userEmail,
			UserPhone:       userPhone,
			LocationName:    locName,
			LocationAddress: locAddr,
			LockerCode:      lockerCode,
			SlotCode:        slotCode,
			SlotSize:        slotSize,
			AuditEvents:     events,
		})
	}
	return list, nil
}

func (r *InMemoryPaymentRepo) GetDetailedByID(ctx context.Context, id uuid.UUID) (*domain.AdminTransactionItem, error) {
	list, err := r.ListAllDetailed(ctx, 100, 0)
	if err != nil {
		return nil, err
	}
	for _, item := range list {
		if item.ID == id {
			return &item, nil
		}
	}
	return nil, domain.ErrNotFound
}

type InMemoryDeviceRepo struct{ s *Store }

func (r *InMemoryDeviceRepo) CreateCommand(ctx context.Context, cmd *domain.DeviceCommand) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Commands[cmd.ID] = *cmd
	return nil
}
func (r *InMemoryDeviceRepo) GetCommandByID(ctx context.Context, id uuid.UUID) (*domain.DeviceCommand, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	if cmd, ok := r.s.Commands[id]; ok {
		return &cmd, nil
	}
	return nil, domain.ErrNotFound
}
func (r *InMemoryDeviceRepo) GetCommandByCorrelationID(ctx context.Context, corrID string) (*domain.DeviceCommand, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, cmd := range r.s.Commands {
		if cmd.CorrelationID == corrID {
			return &cmd, nil
		}
	}
	return nil, domain.ErrNotFound
}
func (r *InMemoryDeviceRepo) UpdateCommandStatus(ctx context.Context, corrID string, status domain.DeviceCommandStatus, ackTime *time.Time, errMsg string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for id, cmd := range r.s.Commands {
		if cmd.CorrelationID == corrID {
			cmd.Status = status
			cmd.AcknowledgedAt = ackTime
			cmd.ErrorMessage = errMsg
			cmd.UpdatedAt = time.Now()
			r.s.Commands[id] = cmd
			return nil
		}
	}
	return domain.ErrNotFound
}
func (r *InMemoryDeviceRepo) UpdateLockerHeartbeat(ctx context.Context, deviceID string, firmware string, lastSeen time.Time) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for id, lk := range r.s.Lockers {
		if lk.DeviceID == deviceID {
			lk.Status = domain.LockerOnline
			if firmware != "" {
				lk.FirmwareVersion = firmware
			}
			lk.LastSeenAt = lastSeen
			lk.UpdatedAt = time.Now()
			r.s.Lockers[id] = lk
			return nil
		}
	}
	return domain.ErrNotFound
}

type InMemoryNotificationRepo struct{ s *Store }

func (r *InMemoryNotificationRepo) Create(ctx context.Context, n *domain.Notification) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Notifications = append([]domain.Notification{*n}, r.s.Notifications...)
	return nil
}
func (r *InMemoryNotificationRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Notification, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var list []domain.Notification
	for _, n := range r.s.Notifications {
		if n.UserID == userID {
			list = append(list, n)
		}
	}
	return list, nil
}
func (r *InMemoryNotificationRepo) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i, n := range r.s.Notifications {
		if n.ID == id && n.UserID == userID {
			now := time.Now()
			r.s.Notifications[i].ReadAt = &now
			return nil
		}
	}
	return domain.ErrNotFound
}
func (r *InMemoryNotificationRepo) RegisterFCMToken(ctx context.Context, dev *domain.FCMDevice) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.FCMDevices[dev.Token] = *dev
	return nil
}
func (r *InMemoryNotificationRepo) GetFCMTokensByUser(ctx context.Context, userID uuid.UUID) ([]string, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var tokens []string
	for _, dev := range r.s.FCMDevices {
		if dev.UserID == userID {
			tokens = append(tokens, dev.Token)
		}
	}
	return tokens, nil
}

type InMemoryPromoRepo struct{ s *Store }

func (r *InMemoryPromoRepo) Create(ctx context.Context, p *domain.Promo) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Promos[strings.ToUpper(p.Code)] = *p
	return nil
}
func (r *InMemoryPromoRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Promo, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, p := range r.s.Promos {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, domain.ErrNotFound
}
func (r *InMemoryPromoRepo) GetByCode(ctx context.Context, code string) (*domain.Promo, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	if p, ok := r.s.Promos[strings.ToUpper(code)]; ok {
		return &p, nil
	}
	return nil, domain.ErrNotFound
}
func (r *InMemoryPromoRepo) Update(ctx context.Context, p *domain.Promo) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Promos[strings.ToUpper(p.Code)] = *p
	return nil
}
func (r *InMemoryPromoRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for code, p := range r.s.Promos {
		if p.ID == id {
			p.Status = status
			p.UpdatedAt = time.Now()
			r.s.Promos[code] = p
			return nil
		}
	}
	return domain.ErrNotFound
}
func (r *InMemoryPromoRepo) Delete(ctx context.Context, id uuid.UUID) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for code, p := range r.s.Promos {
		if p.ID == id {
			delete(r.s.Promos, code)
			return nil
		}
	}
	return domain.ErrNotFound
}
func (r *InMemoryPromoRepo) List(ctx context.Context) ([]domain.Promo, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var list []domain.Promo
	for _, p := range r.s.Promos {
		list = append(list, p)
	}
	return list, nil
}

type InMemoryAuditRepo struct{ s *Store }

func (r *InMemoryAuditRepo) CreateLog(ctx context.Context, l *domain.AuditLog) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.AuditLogs = append([]domain.AuditLog{*l}, r.s.AuditLogs...)
	return nil
}
func (r *InMemoryAuditRepo) ListLogs(ctx context.Context, limit, offset int) ([]domain.AuditLog, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	return r.s.AuditLogs, nil
}

type InMemoryPricingRepo struct{ s *Store }

func (r *InMemoryPricingRepo) GetTiers(ctx context.Context) ([]domain.PricingTierItem, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var list []domain.PricingTierItem
	for _, t := range r.s.PricingTiers {
		list = append(list, t)
	}
	if len(list) == 0 {
		list = []domain.PricingTierItem{
			{SlotSize: "small", HourlyRate: 4000, DepositAmount: 10000},
			{SlotSize: "medium", HourlyRate: 6000, DepositAmount: 15000},
			{SlotSize: "large", HourlyRate: 9000, DepositAmount: 20000},
		}
	}
	return list, nil
}

func (r *InMemoryPricingRepo) UpsertTiers(ctx context.Context, tiers []domain.PricingTierItem) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for _, t := range tiers {
		t.UpdatedAt = time.Now()
		r.s.PricingTiers[strings.ToLower(t.SlotSize)] = t
	}
	return nil
}

type Repositories struct {
	Profiles      repository.ProfileRepository
	Locations     repository.LocationRepository
	Lockers       repository.LockerRepository
	Slots         repository.SlotRepository
	Rentals       repository.RentalRepository
	SecurityCodes repository.SecurityCodeRepository
	Payments      repository.PaymentRepository
	Devices       repository.DeviceRepository
	Notifications repository.NotificationRepository
	Promos        repository.PromoRepository
	Audit         repository.AuditRepository
	Pricing       repository.PricingRepository
}

func NewRepositories(store *Store) *Repositories {
	return &Repositories{
		Profiles:      &InMemoryProfileRepo{s: store},
		Locations:     &InMemoryLocationRepo{s: store},
		Lockers:       &InMemoryLockerRepo{s: store},
		Slots:         &InMemorySlotRepo{s: store},
		Rentals:       &InMemoryRentalRepo{s: store},
		SecurityCodes: &InMemorySecurityCodeRepo{s: store},
		Payments:      &InMemoryPaymentRepo{s: store},
		Devices:       &InMemoryDeviceRepo{s: store},
		Notifications: &InMemoryNotificationRepo{s: store},
		Promos:        &InMemoryPromoRepo{s: store},
		Audit:         &InMemoryAuditRepo{s: store},
		Pricing:       &InMemoryPricingRepo{s: store},
	}
}
