package postgres

import (
	"context"
	"database/sql"
	"errors"

	"lockerin-backend/internal/domain"
	"lockerin-backend/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ProfileRepo
type ProfileRepo struct {
	db *DB
}

func NewProfileRepo(db *DB) repository.ProfileRepository {
	return &ProfileRepo{db: db}
}

func (r *ProfileRepo) Create(ctx context.Context, p *domain.Profile) error {
	query := `
		INSERT INTO profiles (id, email, username, name, phone, avatar_url, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Pool.Exec(ctx, query,
		p.ID, p.Email, p.Username, p.Name, p.Phone, p.AvatarURL, p.Role, p.Status, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *ProfileRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Profile, error) {
	query := `
		SELECT id, email, username, name, phone, COALESCE(avatar_url, ''), role, status, created_at, updated_at
		FROM profiles WHERE id = $1
	`
	var p domain.Profile
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Email, &p.Username, &p.Name, &p.Phone, &p.AvatarURL, &p.Role, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &p, err
}

func (r *ProfileRepo) GetByEmail(ctx context.Context, email string) (*domain.Profile, error) {
	query := `
		SELECT id, email, username, name, phone, COALESCE(avatar_url, ''), role, status, created_at, updated_at
		FROM profiles WHERE LOWER(email) = LOWER($1)
	`
	var p domain.Profile
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(
		&p.ID, &p.Email, &p.Username, &p.Name, &p.Phone, &p.AvatarURL, &p.Role, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &p, err
}

func (r *ProfileRepo) GetByUsername(ctx context.Context, username string) (*domain.Profile, error) {
	query := `
		SELECT id, email, username, name, phone, COALESCE(avatar_url, ''), role, status, created_at, updated_at
		FROM profiles WHERE LOWER(username) = LOWER($1)
	`
	var p domain.Profile
	err := r.db.Pool.QueryRow(ctx, query, username).Scan(
		&p.ID, &p.Email, &p.Username, &p.Name, &p.Phone, &p.AvatarURL, &p.Role, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &p, err
}

func (r *ProfileRepo) Update(ctx context.Context, p *domain.Profile) error {
	query := `
		UPDATE profiles
		SET name = $1, phone = $2, avatar_url = $3, role = $4, status = $5, updated_at = NOW()
		WHERE id = $6
	`
	_, err := r.db.Pool.Exec(ctx, query, p.Name, p.Phone, p.AvatarURL, p.Role, p.Status, p.ID)
	return err
}

func (r *ProfileRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE profiles SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Pool.Exec(ctx, query, status, id)
	return err
}

func (r *ProfileRepo) List(ctx context.Context, limit, offset int) ([]domain.Profile, error) {
	query := `
		SELECT id, email, username, name, phone, COALESCE(avatar_url, ''), role, status, created_at, updated_at
		FROM profiles ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []domain.Profile
	for rows.Next() {
		var p domain.Profile
		if err := rows.Scan(&p.ID, &p.Email, &p.Username, &p.Name, &p.Phone, &p.AvatarURL, &p.Role, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}
	return profiles, nil
}

func (r *ProfileRepo) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM profiles`).Scan(&count)
	return count, err
}

// LocationRepo
type LocationRepo struct {
	db *DB
}

func NewLocationRepo(db *DB) repository.LocationRepository {
	return &LocationRepo{db: db}
}

func (r *LocationRepo) Create(ctx context.Context, loc *domain.Location) error {
	query := `
		INSERT INTO locations (id, code, name, address, city, latitude, longitude, status, operating_hours, image_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.Pool.Exec(ctx, query,
		loc.ID, loc.Code, loc.Name, loc.Address, loc.City, loc.Latitude, loc.Longitude,
		loc.Status, loc.OperatingHours, loc.ImageURL, loc.CreatedAt, loc.UpdatedAt,
	)
	return err
}

func (r *LocationRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Location, error) {
	query := `
		SELECT id, code, name, address, city, latitude, longitude, status, operating_hours, COALESCE(image_url, ''), created_at, updated_at
		FROM locations WHERE id = $1
	`
	var loc domain.Location
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&loc.ID, &loc.Code, &loc.Name, &loc.Address, &loc.City, &loc.Latitude, &loc.Longitude,
		&loc.Status, &loc.OperatingHours, &loc.ImageURL, &loc.CreatedAt, &loc.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &loc, err
}

func (r *LocationRepo) GetByCode(ctx context.Context, code string) (*domain.Location, error) {
	query := `
		SELECT id, code, name, address, city, latitude, longitude, status, operating_hours, COALESCE(image_url, ''), created_at, updated_at
		FROM locations WHERE LOWER(code) = LOWER($1)
	`
	var loc domain.Location
	err := r.db.Pool.QueryRow(ctx, query, code).Scan(
		&loc.ID, &loc.Code, &loc.Name, &loc.Address, &loc.City, &loc.Latitude, &loc.Longitude,
		&loc.Status, &loc.OperatingHours, &loc.ImageURL, &loc.CreatedAt, &loc.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &loc, err
}

func (r *LocationRepo) Update(ctx context.Context, loc *domain.Location) error {
	query := `
		UPDATE locations
		SET name = $1, address = $2, city = $3, latitude = $4, longitude = $5, status = $6, operating_hours = $7, image_url = $8, updated_at = NOW()
		WHERE id = $9
	`
	_, err := r.db.Pool.Exec(ctx, query, loc.Name, loc.Address, loc.City, loc.Latitude, loc.Longitude, loc.Status, loc.OperatingHours, loc.ImageURL, loc.ID)
	return err
}

func (r *LocationRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE locations SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Pool.Exec(ctx, query, status, id)
	return err
}

func (r *LocationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM locations WHERE id = $1`, id)
	return err
}

func (r *LocationRepo) List(ctx context.Context, limit, offset int) ([]domain.Location, error) {
	query := `
		SELECT 
			l.id, l.code, l.name, l.address, l.city, l.latitude, l.longitude, l.status, l.operating_hours, COALESCE(l.image_url, ''), l.created_at, l.updated_at,
			COUNT(DISTINCT lk.id) as total_lockers,
			COUNT(s.id) FILTER (WHERE s.status = 'available') as available_slots,
			COUNT(s.id) FILTER (WHERE s.status IN ('occupied', 'reserved')) as occupied_slots,
			COUNT(s.id) FILTER (WHERE s.status = 'maintenance') as maintenance_slots
		FROM locations l
		LEFT JOIN lockers lk ON l.id = lk.location_id
		LEFT JOIN locker_slots s ON lk.id = s.locker_id
		GROUP BY l.id
		ORDER BY l.name ASC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []domain.Location
	for rows.Next() {
		var loc domain.Location
		if err := rows.Scan(
			&loc.ID, &loc.Code, &loc.Name, &loc.Address, &loc.City, &loc.Latitude, &loc.Longitude,
			&loc.Status, &loc.OperatingHours, &loc.ImageURL, &loc.CreatedAt, &loc.UpdatedAt,
			&loc.TotalLockers, &loc.AvailableSlots, &loc.OccupiedSlots, &loc.MaintenanceSlots,
		); err != nil {
			return nil, err
		}
		locations = append(locations, loc)
	}
	return locations, nil
}

func (r *LocationRepo) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM locations`).Scan(&count)
	return count, err
}

// Locker & Slot Repos
type LockerRepo struct {
	db *DB
}

func NewLockerRepo(db *DB) repository.LockerRepository {
	return &LockerRepo{db: db}
}

func (r *LockerRepo) Create(ctx context.Context, l *domain.Locker) error {
	query := `
		INSERT INTO lockers (id, location_id, code, name, device_id, status, firmware_version, last_seen_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Pool.Exec(ctx, query, l.ID, l.LocationID, l.Code, l.Name, l.DeviceID, l.Status, l.FirmwareVersion, l.LastSeenAt, l.CreatedAt, l.UpdatedAt)
	return err
}

func (r *LockerRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Locker, error) {
	query := `
		SELECT id, location_id, code, name, device_id, status, firmware_version, last_seen_at, created_at, updated_at
		FROM lockers WHERE id = $1
	`
	var l domain.Locker
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&l.ID, &l.LocationID, &l.Code, &l.Name, &l.DeviceID, &l.Status, &l.FirmwareVersion, &l.LastSeenAt, &l.CreatedAt, &l.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &l, err
}

func (r *LockerRepo) GetByDeviceID(ctx context.Context, deviceID string) (*domain.Locker, error) {
	query := `
		SELECT id, location_id, code, name, device_id, status, firmware_version, last_seen_at, created_at, updated_at
		FROM lockers WHERE device_id = $1
	`
	var l domain.Locker
	err := r.db.Pool.QueryRow(ctx, query, deviceID).Scan(
		&l.ID, &l.LocationID, &l.Code, &l.Name, &l.DeviceID, &l.Status, &l.FirmwareVersion, &l.LastSeenAt, &l.CreatedAt, &l.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &l, err
}

func (r *LockerRepo) Update(ctx context.Context, l *domain.Locker) error {
	query := `
		UPDATE lockers
		SET name = $1, status = $2, firmware_version = $3, last_seen_at = $4, updated_at = NOW()
		WHERE id = $5
	`
	_, err := r.db.Pool.Exec(ctx, query, l.Name, l.Status, l.FirmwareVersion, l.LastSeenAt, l.ID)
	return err
}

func (r *LockerRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM lockers WHERE id = $1`, id)
	return err
}

func (r *LockerRepo) ListByLocation(ctx context.Context, locationID uuid.UUID) ([]domain.Locker, error) {
	query := `
		SELECT id, location_id, code, name, device_id, status, firmware_version, last_seen_at, created_at, updated_at
		FROM lockers WHERE location_id = $1 ORDER BY code ASC
	`
	rows, err := r.db.Pool.Query(ctx, query, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lockers []domain.Locker
	for rows.Next() {
		var l domain.Locker
		if err := rows.Scan(&l.ID, &l.LocationID, &l.Code, &l.Name, &l.DeviceID, &l.Status, &l.FirmwareVersion, &l.LastSeenAt, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		lockers = append(lockers, l)
	}
	return lockers, nil
}

func (r *LockerRepo) ListAll(ctx context.Context) ([]domain.Locker, error) {
	query := `
		SELECT id, location_id, code, name, device_id, status, firmware_version, last_seen_at, created_at, updated_at
		FROM lockers ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lockers []domain.Locker
	for rows.Next() {
		var l domain.Locker
		if err := rows.Scan(&l.ID, &l.LocationID, &l.Code, &l.Name, &l.DeviceID, &l.Status, &l.FirmwareVersion, &l.LastSeenAt, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		lockers = append(lockers, l)
	}
	return lockers, nil
}

func (r *LockerRepo) ListWithStats(ctx context.Context) ([]domain.LockerAdminItem, error) {
	query := `
		SELECT 
			lk.id, lk.device_id, lk.code, lk.name, lk.location_id, loc.name as location_name,
			lk.status, COALESCE(lk.firmware_version, '1.0.0'),
			95 as battery_level, -65 as signal_strength, lk.last_seen_at as last_heartbeat,
			COUNT(s.id) as total_slots,
			COUNT(s.id) FILTER (WHERE s.status = 'available') as available_slots,
			COUNT(s.id) FILTER (WHERE s.status IN ('occupied', 'reserved')) as occupied_slots
		FROM lockers lk
		JOIN locations loc ON lk.location_id = loc.id
		LEFT JOIN locker_slots s ON lk.id = s.locker_id
		GROUP BY lk.id, loc.name
		ORDER BY lk.code ASC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.LockerAdminItem
	for rows.Next() {
		var item domain.LockerAdminItem
		if err := rows.Scan(
			&item.ID, &item.DeviceID, &item.Code, &item.Name, &item.LocationID, &item.LocationName,
			&item.Status, &item.FirmwareVersion, &item.BatteryLevel, &item.SignalStrength, &item.LastHeartbeat,
			&item.TotalSlots, &item.AvailableSlots, &item.OccupiedSlots,
		); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, nil
}

func (r *LockerRepo) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM lockers`).Scan(&count)
	return count, err
}

// SlotRepo
type SlotRepo struct {
	db *DB
}

func NewSlotRepo(db *DB) repository.SlotRepository {
	return &SlotRepo{db: db}
}

func (r *SlotRepo) Create(ctx context.Context, s *domain.LockerSlot) error {
	query := `
		INSERT INTO locker_slots (id, locker_id, slot_code, size, status, base_price_per_hour, current_rental_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Pool.Exec(ctx, query, s.ID, s.LockerID, s.SlotCode, s.Size, s.Status, s.BasePricePerHour, s.CurrentRentalID, s.CreatedAt, s.UpdatedAt)
	return err
}

func (r *SlotRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.LockerSlot, error) {
	query := `
		SELECT s.id, s.locker_id, s.slot_code, s.size, s.status, s.base_price_per_hour, s.current_rental_id, s.created_at, s.updated_at, lk.code
		FROM locker_slots s
		JOIN lockers lk ON s.locker_id = lk.id
		WHERE s.id = $1
	`
	var s domain.LockerSlot
	var curRental sql.NullString
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.LockerID, &s.SlotCode, &s.Size, &s.Status, &s.BasePricePerHour, &curRental, &s.CreatedAt, &s.UpdatedAt, &s.LockerCode,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if curRental.Valid {
		uid, _ := uuid.Parse(curRental.String)
		s.CurrentRentalID = &uid
	}
	return &s, err
}

func (r *SlotRepo) GetByLockerAndCode(ctx context.Context, lockerID uuid.UUID, slotCode string) (*domain.LockerSlot, error) {
	query := `
		SELECT id, locker_id, slot_code, size, status, base_price_per_hour, current_rental_id, created_at, updated_at
		FROM locker_slots WHERE locker_id = $1 AND LOWER(slot_code) = LOWER($2)
	`
	var s domain.LockerSlot
	var curRental sql.NullString
	err := r.db.Pool.QueryRow(ctx, query, lockerID, slotCode).Scan(
		&s.ID, &s.LockerID, &s.SlotCode, &s.Size, &s.Status, &s.BasePricePerHour, &curRental, &s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if curRental.Valid {
		uid, _ := uuid.Parse(curRental.String)
		s.CurrentRentalID = &uid
	}
	return &s, err
}

func (r *SlotRepo) Update(ctx context.Context, s *domain.LockerSlot) error {
	query := `
		UPDATE locker_slots
		SET size = $1, status = $2, base_price_per_hour = $3, current_rental_id = $4, updated_at = NOW()
		WHERE id = $5
	`
	_, err := r.db.Pool.Exec(ctx, query, s.Size, s.Status, s.BasePricePerHour, s.CurrentRentalID, s.ID)
	return err
}

func (r *SlotRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.SlotStatus, currentRentalID *uuid.UUID) error {
	query := `
		UPDATE locker_slots
		SET status = $1, current_rental_id = $2, updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.Pool.Exec(ctx, query, status, currentRentalID, id)
	return err
}

func (r *SlotRepo) ListByLocker(ctx context.Context, lockerID uuid.UUID) ([]domain.LockerSlot, error) {
	query := `
		SELECT id, locker_id, slot_code, size, status, base_price_per_hour, current_rental_id, created_at, updated_at
		FROM locker_slots WHERE locker_id = $1 ORDER BY slot_code ASC
	`
	rows, err := r.db.Pool.Query(ctx, query, lockerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []domain.LockerSlot
	for rows.Next() {
		var s domain.LockerSlot
		var curRental sql.NullString
		if err := rows.Scan(&s.ID, &s.LockerID, &s.SlotCode, &s.Size, &s.Status, &s.BasePricePerHour, &curRental, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		if curRental.Valid {
			uid, _ := uuid.Parse(curRental.String)
			s.CurrentRentalID = &uid
		}
		slots = append(slots, s)
	}
	return slots, nil
}

func (r *SlotRepo) CountByStatus(ctx context.Context) (total, available, occupied, maintenance int, err error) {
	query := `
		SELECT 
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'available'),
			COUNT(*) FILTER (WHERE status IN ('occupied', 'reserved')),
			COUNT(*) FILTER (WHERE status = 'maintenance')
		FROM locker_slots
	`
	err = r.db.Pool.QueryRow(ctx, query).Scan(&total, &available, &occupied, &maintenance)
	return
}

func (r *SlotRepo) GetPricing(ctx context.Context) ([]domain.PricingTier, error) {
	query := `
		SELECT size, COALESCE(AVG(base_price_per_hour), 5000.00) as avg_price
		FROM locker_slots
		GROUP BY size
		ORDER BY avg_price ASC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tiers := []domain.PricingTier{}
	for rows.Next() {
		var size domain.SlotSize
		var price float64
		if err := rows.Scan(&size, &price); err != nil {
			return nil, err
		}
		desc := "Cocok untuk tas ransel & barang pribadi"
		usage := "Tas ransel, jaket, dokumen"
		if size == domain.SizeSmall {
			desc = "Cocok untuk barang kecil, dompet, gadget"
			usage = "Smartphone, dompet, helm kecil"
		} else if size == domain.SizeLarge || size == domain.SizeXL {
			desc = "Cocok untuk koper kabin & koper besar"
			usage = "Koper 20-24 inch, tas belanja banyak"
		}
		tiers = append(tiers, domain.PricingTier{
			Size:             size,
			PricePerHour:     price,
			Description:      desc,
			RecommendedUsage: usage,
		})
	}

	if len(tiers) == 0 {
		tiers = []domain.PricingTier{
			{Size: domain.SizeSmall, PricePerHour: 4000, Description: "Cocok untuk gadget & dompet", RecommendedUsage: "Smartphone, dompet"},
			{Size: domain.SizeMedium, PricePerHour: 6000, Description: "Cocok untuk tas ransel & barang sehari-hari", RecommendedUsage: "Tas ransel, laptop, jaket"},
			{Size: domain.SizeLarge, PricePerHour: 9000, Description: "Cocok untuk koper kabin & koper besar", RecommendedUsage: "Koper, tas belanja banyak"},
		}
	}
	return tiers, nil
}

func (r *SlotRepo) UpdateBasePrices(ctx context.Context, prices map[domain.SlotSize]float64) error {
	for size, price := range prices {
		if price > 0 {
			_, err := r.db.Pool.Exec(ctx, `UPDATE locker_slots SET base_price_per_hour = $1, updated_at = NOW() WHERE size = $2`, price, size)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// PricingRepo
type PricingRepo struct {
	db *DB
}

func NewPricingRepo(db *DB) repository.PricingRepository {
	return &PricingRepo{db: db}
}

func (r *PricingRepo) GetTiers(ctx context.Context) ([]domain.PricingTierItem, error) {
	query := `
		SELECT id, slot_size, hourly_rate, deposit_amount, updated_at
		FROM pricing_tiers
		ORDER BY hourly_rate ASC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tiers []domain.PricingTierItem
	for rows.Next() {
		var t domain.PricingTierItem
		if err := rows.Scan(&t.ID, &t.SlotSize, &t.HourlyRate, &t.DepositAmount, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tiers = append(tiers, t)
	}

	if len(tiers) == 0 {
		tiers = []domain.PricingTierItem{
			{SlotSize: "small", HourlyRate: 4000, DepositAmount: 10000},
			{SlotSize: "medium", HourlyRate: 6000, DepositAmount: 15000},
			{SlotSize: "large", HourlyRate: 9000, DepositAmount: 20000},
		}
	}

	return tiers, nil
}

func (r *PricingRepo) UpsertTiers(ctx context.Context, tiers []domain.PricingTierItem) error {
	for _, t := range tiers {
		query := `
			INSERT INTO pricing_tiers (id, slot_size, hourly_rate, deposit_amount, updated_at)
			VALUES (gen_random_uuid(), LOWER($1), $2, $3, NOW())
			ON CONFLICT (slot_size) DO UPDATE
			SET hourly_rate = EXCLUDED.hourly_rate, deposit_amount = EXCLUDED.deposit_amount, updated_at = NOW()
		`
		if _, err := r.db.Pool.Exec(ctx, query, t.SlotSize, t.HourlyRate, t.DepositAmount); err != nil {
			return err
		}
	}
	return nil
}

