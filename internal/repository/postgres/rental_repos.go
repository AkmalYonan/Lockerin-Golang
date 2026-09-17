package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"lockerin-backend/internal/domain"
	"lockerin-backend/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// RentalRepo
type RentalRepo struct {
	db *DB
}

func NewRentalRepo(db *DB) repository.RentalRepository {
	return &RentalRepo{db: db}
}

func (r *RentalRepo) Create(ctx context.Context, rent *domain.Rental) error {
	query := `
		INSERT INTO rentals (
			id, user_id, location_id, locker_id, slot_id, status, duration_hours,
			base_price_snapshot, discount_amount, total_amount,
			reservation_expires_at, started_at, expires_at, ended_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`
	_, err := r.db.Pool.Exec(ctx, query,
		rent.ID, rent.UserID, rent.LocationID, rent.LockerID, rent.SlotID, rent.Status, rent.DurationHours,
		rent.BasePriceSnapshot, rent.DiscountAmount, rent.TotalAmount,
		rent.ReservationExpiresAt, rent.StartedAt, rent.ExpiresAt, rent.EndedAt, rent.CreatedAt, rent.UpdatedAt,
	)
	return err
}

func (r *RentalRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Rental, error) {
	query := `
		SELECT 
			r.id, r.user_id, r.location_id, r.locker_id, r.slot_id, r.status, r.duration_hours,
			r.base_price_snapshot, r.discount_amount, r.total_amount,
			r.reservation_expires_at, r.started_at, r.expires_at, r.ended_at, r.created_at, r.updated_at,
			p.name as user_name, loc.name as location_name, lk.code as locker_code, s.slot_code
		FROM rentals r
		JOIN profiles p ON r.user_id = p.id
		JOIN locations loc ON r.location_id = loc.id
		JOIN lockers lk ON r.locker_id = lk.id
		JOIN locker_slots s ON r.slot_id = s.id
		WHERE r.id = $1
	`
	var rent domain.Rental
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&rent.ID, &rent.UserID, &rent.LocationID, &rent.LockerID, &rent.SlotID, &rent.Status, &rent.DurationHours,
		&rent.BasePriceSnapshot, &rent.DiscountAmount, &rent.TotalAmount,
		&rent.ReservationExpiresAt, &rent.StartedAt, &rent.ExpiresAt, &rent.EndedAt, &rent.CreatedAt, &rent.UpdatedAt,
		&rent.UserName, &rent.LocationName, &rent.LockerCode, &rent.SlotCode,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &rent, err
}

func (r *RentalRepo) GetActiveBySlotID(ctx context.Context, slotID uuid.UUID) (*domain.Rental, error) {
	query := `
		SELECT id, user_id, location_id, locker_id, slot_id, status, duration_hours,
		       base_price_snapshot, discount_amount, total_amount,
		       reservation_expires_at, started_at, expires_at, ended_at, created_at, updated_at
		FROM rentals
		WHERE slot_id = $1 AND status IN ('reserved', 'awaiting_payment', 'paid', 'active')
		LIMIT 1
	`
	var rent domain.Rental
	err := r.db.Pool.QueryRow(ctx, query, slotID).Scan(
		&rent.ID, &rent.UserID, &rent.LocationID, &rent.LockerID, &rent.SlotID, &rent.Status, &rent.DurationHours,
		&rent.BasePriceSnapshot, &rent.DiscountAmount, &rent.TotalAmount,
		&rent.ReservationExpiresAt, &rent.StartedAt, &rent.ExpiresAt, &rent.EndedAt, &rent.CreatedAt, &rent.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &rent, err
}

func (r *RentalRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Rental, error) {
	query := `
		SELECT 
			r.id, r.user_id, r.location_id, r.locker_id, r.slot_id, r.status, r.duration_hours,
			r.base_price_snapshot, r.discount_amount, r.total_amount,
			r.reservation_expires_at, r.started_at, r.expires_at, r.ended_at, r.created_at, r.updated_at,
			p.name as user_name, loc.name as location_name, lk.code as locker_code, s.slot_code
		FROM rentals r
		JOIN profiles p ON r.user_id = p.id
		JOIN locations loc ON r.location_id = loc.id
		JOIN lockers lk ON r.locker_id = lk.id
		JOIN locker_slots s ON r.slot_id = s.id
		WHERE r.user_id = $1
		ORDER BY r.created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rentals []domain.Rental
	for rows.Next() {
		var rent domain.Rental
		if err := rows.Scan(
			&rent.ID, &rent.UserID, &rent.LocationID, &rent.LockerID, &rent.SlotID, &rent.Status, &rent.DurationHours,
			&rent.BasePriceSnapshot, &rent.DiscountAmount, &rent.TotalAmount,
			&rent.ReservationExpiresAt, &rent.StartedAt, &rent.ExpiresAt, &rent.EndedAt, &rent.CreatedAt, &rent.UpdatedAt,
			&rent.UserName, &rent.LocationName, &rent.LockerCode, &rent.SlotCode,
		); err != nil {
			return nil, err
		}
		rentals = append(rentals, rent)
	}
	return rentals, nil
}

func (r *RentalRepo) ListAll(ctx context.Context, limit, offset int) ([]domain.Rental, error) {
	query := `
		SELECT 
			r.id, r.user_id, r.location_id, r.locker_id, r.slot_id, r.status, r.duration_hours,
			r.base_price_snapshot, r.discount_amount, r.total_amount,
			r.reservation_expires_at, r.started_at, r.expires_at, r.ended_at, r.created_at, r.updated_at,
			p.name as user_name, loc.name as location_name, lk.code as locker_code, s.slot_code
		FROM rentals r
		JOIN profiles p ON r.user_id = p.id
		JOIN locations loc ON r.location_id = loc.id
		JOIN lockers lk ON r.locker_id = lk.id
		JOIN locker_slots s ON r.slot_id = s.id
		ORDER BY r.created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rentals []domain.Rental
	for rows.Next() {
		var rent domain.Rental
		if err := rows.Scan(
			&rent.ID, &rent.UserID, &rent.LocationID, &rent.LockerID, &rent.SlotID, &rent.Status, &rent.DurationHours,
			&rent.BasePriceSnapshot, &rent.DiscountAmount, &rent.TotalAmount,
			&rent.ReservationExpiresAt, &rent.StartedAt, &rent.ExpiresAt, &rent.EndedAt, &rent.CreatedAt, &rent.UpdatedAt,
			&rent.UserName, &rent.LocationName, &rent.LockerCode, &rent.SlotCode,
		); err != nil {
			return nil, err
		}
		rentals = append(rentals, rent)
	}
	return rentals, nil
}

func (r *RentalRepo) Update(ctx context.Context, rent *domain.Rental) error {
	query := `
		UPDATE rentals
		SET status = $1, started_at = $2, expires_at = $3, ended_at = $4, updated_at = NOW()
		WHERE id = $5
	`
	_, err := r.db.Pool.Exec(ctx, query, rent.Status, rent.StartedAt, rent.ExpiresAt, rent.EndedAt, rent.ID)
	return err
}

func (r *RentalRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.RentalStatus) error {
	query := `UPDATE rentals SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Pool.Exec(ctx, query, status, id)
	return err
}

func (r *RentalRepo) ExpireOverdueReservations(ctx context.Context, now time.Time) ([]domain.Rental, error) {
	query := `
		UPDATE rentals
		SET status = 'expired', updated_at = NOW()
		WHERE status = 'reserved' AND reservation_expires_at < $1
		RETURNING id, user_id, location_id, locker_id, slot_id, status, duration_hours,
		          base_price_snapshot, discount_amount, total_amount,
		          reservation_expires_at, started_at, expires_at, ended_at, created_at, updated_at
	`
	rows, err := r.db.Pool.Query(ctx, query, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expired []domain.Rental
	for rows.Next() {
		var rent domain.Rental
		if err := rows.Scan(
			&rent.ID, &rent.UserID, &rent.LocationID, &rent.LockerID, &rent.SlotID, &rent.Status, &rent.DurationHours,
			&rent.BasePriceSnapshot, &rent.DiscountAmount, &rent.TotalAmount,
			&rent.ReservationExpiresAt, &rent.StartedAt, &rent.ExpiresAt, &rent.EndedAt, &rent.CreatedAt, &rent.UpdatedAt,
		); err != nil {
			return nil, err
		}
		expired = append(expired, rent)
	}
	return expired, nil
}

func (r *RentalRepo) CountActive(ctx context.Context) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM rentals WHERE status = 'active'`).Scan(&count)
	return count, err
}

func (r *RentalRepo) GetTotalRevenue(ctx context.Context) (total float64, today float64, err error) {
	query := `
		SELECT 
			COALESCE(SUM(total_amount) FILTER (WHERE status IN ('paid', 'active', 'completed')), 0) as total_rev,
			COALESCE(SUM(total_amount) FILTER (WHERE status IN ('paid', 'active', 'completed') AND created_at >= CURRENT_DATE), 0) as today_rev
		FROM rentals
	`
	err = r.db.Pool.QueryRow(ctx, query).Scan(&total, &today)
	return
}

// SecurityCodeRepo
type SecurityCodeRepo struct {
	db *DB
}

func NewSecurityCodeRepo(db *DB) repository.SecurityCodeRepository {
	return &SecurityCodeRepo{db: db}
}

func (r *SecurityCodeRepo) Create(ctx context.Context, c *domain.SecurityCode) error {
	query := `
		INSERT INTO security_codes (id, rental_id, pin_hash, salt, attempts, max_attempts, locked_until, expires_at, used_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Pool.Exec(ctx, query, c.ID, c.RentalID, c.PINHash, c.Salt, c.Attempts, c.MaxAttempts, c.LockedUntil, c.ExpiresAt, c.UsedAt, c.CreatedAt, c.UpdatedAt)
	return err
}

func (r *SecurityCodeRepo) GetByRentalID(ctx context.Context, rentalID uuid.UUID) (*domain.SecurityCode, error) {
	query := `
		SELECT id, rental_id, pin_hash, salt, attempts, max_attempts, locked_until, expires_at, used_at, created_at, updated_at
		FROM security_codes WHERE rental_id = $1
	`
	var c domain.SecurityCode
	err := r.db.Pool.QueryRow(ctx, query, rentalID).Scan(
		&c.ID, &c.RentalID, &c.PINHash, &c.Salt, &c.Attempts, &c.MaxAttempts, &c.LockedUntil, &c.ExpiresAt, &c.UsedAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &c, err
}

func (r *SecurityCodeRepo) IncrementAttempts(ctx context.Context, rentalID uuid.UUID) error {
	query := `UPDATE security_codes SET attempts = attempts + 1, updated_at = NOW() WHERE rental_id = $1`
	_, err := r.db.Pool.Exec(ctx, query, rentalID)
	return err
}

func (r *SecurityCodeRepo) SetLockout(ctx context.Context, rentalID uuid.UUID, until time.Time) error {
	query := `UPDATE security_codes SET locked_until = $1, updated_at = NOW() WHERE rental_id = $2`
	_, err := r.db.Pool.Exec(ctx, query, until, rentalID)
	return err
}

func (r *SecurityCodeRepo) MarkUsed(ctx context.Context, rentalID uuid.UUID, usedAt time.Time) error {
	query := `UPDATE security_codes SET used_at = $1, updated_at = NOW() WHERE rental_id = $2`
	_, err := r.db.Pool.Exec(ctx, query, usedAt, rentalID)
	return err
}

func (r *SecurityCodeRepo) Update(ctx context.Context, c *domain.SecurityCode) error {
	query := `
		UPDATE security_codes
		SET pin_hash = $1, attempts = $2, locked_until = $3, expires_at = $4, used_at = $5, updated_at = NOW()
		WHERE id = $6
	`
	_, err := r.db.Pool.Exec(ctx, query, c.PINHash, c.Attempts, c.LockedUntil, c.ExpiresAt, c.UsedAt, c.ID)
	return err
}

// PaymentRepo
type PaymentRepo struct {
	db *DB
}

func NewPaymentRepo(db *DB) repository.PaymentRepository {
	return &PaymentRepo{db: db}
}

func (r *PaymentRepo) Create(ctx context.Context, p *domain.Payment) error {
	query := `
		INSERT INTO payments (id, rental_id, provider, order_id, provider_transaction_id, amount, status, payment_type, snap_token, snap_redirect_url, raw_response, paid_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := r.db.Pool.Exec(ctx, query,
		p.ID, p.RentalID, p.Provider, p.OrderID, p.ProviderTransactionID, p.Amount, p.Status,
		p.PaymentType, p.SnapToken, p.SnapRedirectURL, p.RawResponse, p.PaidAt, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *PaymentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	query := `
		SELECT id, rental_id, provider, order_id, COALESCE(provider_transaction_id, ''), amount, status, COALESCE(payment_type, ''), COALESCE(snap_token, ''), COALESCE(snap_redirect_url, ''), raw_response, paid_at, created_at, updated_at
		FROM payments WHERE id = $1
	`
	var p domain.Payment
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.RentalID, &p.Provider, &p.OrderID, &p.ProviderTransactionID, &p.Amount, &p.Status,
		&p.PaymentType, &p.SnapToken, &p.SnapRedirectURL, &p.RawResponse, &p.PaidAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &p, err
}

func (r *PaymentRepo) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	query := `
		SELECT id, rental_id, provider, order_id, COALESCE(provider_transaction_id, ''), amount, status, COALESCE(payment_type, ''), COALESCE(snap_token, ''), COALESCE(snap_redirect_url, ''), raw_response, paid_at, created_at, updated_at
		FROM payments WHERE order_id = $1
	`
	var p domain.Payment
	err := r.db.Pool.QueryRow(ctx, query, orderID).Scan(
		&p.ID, &p.RentalID, &p.Provider, &p.OrderID, &p.ProviderTransactionID, &p.Amount, &p.Status,
		&p.PaymentType, &p.SnapToken, &p.SnapRedirectURL, &p.RawResponse, &p.PaidAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &p, err
}

func (r *PaymentRepo) GetByRentalID(ctx context.Context, rentalID uuid.UUID) (*domain.Payment, error) {
	query := `
		SELECT id, rental_id, provider, order_id, COALESCE(provider_transaction_id, ''), amount, status, COALESCE(payment_type, ''), COALESCE(snap_token, ''), COALESCE(snap_redirect_url, ''), raw_response, paid_at, created_at, updated_at
		FROM payments WHERE rental_id = $1 ORDER BY created_at DESC LIMIT 1
	`
	var p domain.Payment
	err := r.db.Pool.QueryRow(ctx, query, rentalID).Scan(
		&p.ID, &p.RentalID, &p.Provider, &p.OrderID, &p.ProviderTransactionID, &p.Amount, &p.Status,
		&p.PaymentType, &p.SnapToken, &p.SnapRedirectURL, &p.RawResponse, &p.PaidAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &p, err
}

func (r *PaymentRepo) Update(ctx context.Context, p *domain.Payment) error {
	query := `
		UPDATE payments
		SET status = $1, provider_transaction_id = $2, payment_type = $3, raw_response = $4, paid_at = $5, updated_at = NOW()
		WHERE id = $6
	`
	_, err := r.db.Pool.Exec(ctx, query, p.Status, p.ProviderTransactionID, p.PaymentType, p.RawResponse, p.PaidAt, p.ID)
	return err
}

func (r *PaymentRepo) ListAll(ctx context.Context, limit, offset int) ([]domain.Payment, error) {
	query := `
		SELECT id, rental_id, provider, order_id, COALESCE(provider_transaction_id, ''), amount, status, COALESCE(payment_type, ''), COALESCE(snap_token, ''), COALESCE(snap_redirect_url, ''), raw_response, paid_at, created_at, updated_at
		FROM payments ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []domain.Payment
	for rows.Next() {
		var p domain.Payment
		if err := rows.Scan(
			&p.ID, &p.RentalID, &p.Provider, &p.OrderID, &p.ProviderTransactionID, &p.Amount, &p.Status,
			&p.PaymentType, &p.SnapToken, &p.SnapRedirectURL, &p.RawResponse, &p.PaidAt, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, nil
}

func (r *PaymentRepo) ListAllDetailed(ctx context.Context, limit, offset int) ([]domain.AdminTransactionItem, error) {
	query := `
		SELECT 
			p.id,
			p.order_id,
			p.amount,
			COALESCE(p.payment_type, 'qris') as payment_type,
			p.status,
			COALESCE(p.provider_transaction_id, '') as midtrans_id,
			COALESCE(p.snap_token, '') as snap_token,
			p.created_at,
			p.paid_at,
			COALESCE(u.name, 'Pelanggan Lockerin') as user_name,
			COALESCE(u.email, 'user@lockerin.id') as user_email,
			COALESCE(u.phone, '081234567890') as user_phone,
			COALESCE(loc.name, 'Lockerin RS Medistra') as location_name,
			COALESCE(loc.address, 'Jl. Gatot Subroto No. 59') as location_address,
			COALESCE(l.device_id, 'HEAD-LOCKER-MDS-001') as locker_code,
			COALESCE(s.slot_code, 'A1') as slot_code,
			COALESCE(s.size, 'Small') as slot_size,
			r.created_at as rental_created_at
		FROM payments p
		LEFT JOIN rentals r ON p.rental_id = r.id
		LEFT JOIN profiles u ON r.user_id = u.id
		LEFT JOIN locker_slots s ON r.slot_id = s.id
		LEFT JOIN lockers l ON s.locker_id = l.id
		LEFT JOIN locations loc ON l.location_id = loc.id
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.AdminTransactionItem
	for rows.Next() {
		var item domain.AdminTransactionItem
		var rentalCreatedAt sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.Amount,
			&item.PaymentMethod,
			&item.Status,
			&item.MidtransID,
			&item.SnapToken,
			&item.CreatedAt,
			&item.SettledAt,
			&item.UserName,
			&item.UserEmail,
			&item.UserPhone,
			&item.LocationName,
			&item.LocationAddress,
			&item.LockerCode,
			&item.SlotCode,
			&item.SlotSize,
			&rentalCreatedAt,
		); err != nil {
			return nil, err
		}

		rTimeStr := item.CreatedAt.Format("2006-01-02 15:04:05")
		if rentalCreatedAt.Valid {
			rTimeStr = rentalCreatedAt.Time.Format("2006-01-02 15:04:05")
		}
		pTimeStr := item.CreatedAt.Format("2006-01-02 15:04:05")

		events := []domain.AdminAuditEvent{
			{Step: "Slot Reserved", Time: rTimeStr, Detail: fmt.Sprintf("TTL 15 Menit terkunci untuk User (%s)", item.UserName)},
			{Step: "Payment Created", Time: pTimeStr, Detail: fmt.Sprintf("Snap %s token generated (Rp %.0f)", strings.ToUpper(item.PaymentMethod), item.Amount)},
		}
		if item.SettledAt != nil {
			sTimeStr := item.SettledAt.Format("2006-01-02 15:04:05")
			events = append(events, domain.AdminAuditEvent{
				Step:   "Payment Settled",
				Time:   sTimeStr,
				Detail: "Webhook Midtrans berhasil diverifikasi, PIN Solenoid aktif",
			})
		}
		item.AuditEvents = events
		items = append(items, item)
	}
	return items, nil
}

func (r *PaymentRepo) GetDetailedByID(ctx context.Context, id uuid.UUID) (*domain.AdminTransactionItem, error) {
	query := `
		SELECT 
			p.id,
			p.order_id,
			p.amount,
			COALESCE(p.payment_type, 'qris') as payment_type,
			p.status,
			COALESCE(p.provider_transaction_id, '') as midtrans_id,
			COALESCE(p.snap_token, '') as snap_token,
			p.created_at,
			p.paid_at,
			COALESCE(u.name, 'Pelanggan Lockerin') as user_name,
			COALESCE(u.email, 'user@lockerin.id') as user_email,
			COALESCE(u.phone, '081234567890') as user_phone,
			COALESCE(loc.name, 'Lockerin RS Medistra') as location_name,
			COALESCE(loc.address, 'Jl. Gatot Subroto No. 59') as location_address,
			COALESCE(l.device_id, 'HEAD-LOCKER-MDS-001') as locker_code,
			COALESCE(s.slot_code, 'A1') as slot_code,
			COALESCE(s.size, 'Small') as slot_size,
			r.created_at as rental_created_at
		FROM payments p
		LEFT JOIN rentals r ON p.rental_id = r.id
		LEFT JOIN profiles u ON r.user_id = u.id
		LEFT JOIN locker_slots s ON r.slot_id = s.id
		LEFT JOIN lockers l ON s.locker_id = l.id
		LEFT JOIN locations loc ON l.location_id = loc.id
		WHERE p.id = $1
	`
	var item domain.AdminTransactionItem
	var rentalCreatedAt sql.NullTime
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.OrderID,
		&item.Amount,
		&item.PaymentMethod,
		&item.Status,
		&item.MidtransID,
		&item.SnapToken,
		&item.CreatedAt,
		&item.SettledAt,
		&item.UserName,
		&item.UserEmail,
		&item.UserPhone,
		&item.LocationName,
		&item.LocationAddress,
		&item.LockerCode,
		&item.SlotCode,
		&item.SlotSize,
		&rentalCreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	rTimeStr := item.CreatedAt.Format("2006-01-02 15:04:05")
	if rentalCreatedAt.Valid {
		rTimeStr = rentalCreatedAt.Time.Format("2006-01-02 15:04:05")
	}
	pTimeStr := item.CreatedAt.Format("2006-01-02 15:04:05")

	events := []domain.AdminAuditEvent{
		{Step: "Slot Reserved", Time: rTimeStr, Detail: fmt.Sprintf("TTL 15 Menit terkunci untuk User (%s)", item.UserName)},
		{Step: "Payment Created", Time: pTimeStr, Detail: fmt.Sprintf("Snap %s token generated (Rp %.0f)", strings.ToUpper(item.PaymentMethod), item.Amount)},
	}
	if item.SettledAt != nil {
		sTimeStr := item.SettledAt.Format("2006-01-02 15:04:05")
		events = append(events, domain.AdminAuditEvent{
			Step:   "Payment Settled",
			Time:   sTimeStr,
			Detail: "Webhook Midtrans berhasil diverifikasi, PIN Solenoid aktif",
		})
	}
	item.AuditEvents = events
	return &item, nil
}

// DeviceRepo
type DeviceRepo struct {
	db *DB
}

func NewDeviceRepo(db *DB) repository.DeviceRepository {
	return &DeviceRepo{db: db}
}

func (r *DeviceRepo) CreateCommand(ctx context.Context, cmd *domain.DeviceCommand) error {
	query := `
		INSERT INTO device_commands (id, locker_id, slot_id, command, correlation_id, payload, status, requested_at, acknowledged_at, error_message, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.Pool.Exec(ctx, query, cmd.ID, cmd.LockerID, cmd.SlotID, cmd.Command, cmd.CorrelationID, cmd.Payload, cmd.Status, cmd.RequestedAt, cmd.AcknowledgedAt, cmd.ErrorMessage, cmd.CreatedAt, cmd.UpdatedAt)
	return err
}

func (r *DeviceRepo) GetCommandByID(ctx context.Context, id uuid.UUID) (*domain.DeviceCommand, error) {
	query := `
		SELECT id, locker_id, slot_id, command, correlation_id, payload, status, requested_at, acknowledged_at, COALESCE(error_message, ''), created_at, updated_at
		FROM device_commands WHERE id = $1
	`
	var cmd domain.DeviceCommand
	var slotID sql.NullString
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&cmd.ID, &cmd.LockerID, &slotID, &cmd.Command, &cmd.CorrelationID, &cmd.Payload, &cmd.Status, &cmd.RequestedAt, &cmd.AcknowledgedAt, &cmd.ErrorMessage, &cmd.CreatedAt, &cmd.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if slotID.Valid {
		uid, _ := uuid.Parse(slotID.String)
		cmd.SlotID = &uid
	}
	return &cmd, err
}

func (r *DeviceRepo) GetCommandByCorrelationID(ctx context.Context, corrID string) (*domain.DeviceCommand, error) {
	query := `
		SELECT id, locker_id, slot_id, command, correlation_id, payload, status, requested_at, acknowledged_at, COALESCE(error_message, ''), created_at, updated_at
		FROM device_commands WHERE correlation_id = $1
	`
	var cmd domain.DeviceCommand
	var slotID sql.NullString
	err := r.db.Pool.QueryRow(ctx, query, corrID).Scan(
		&cmd.ID, &cmd.LockerID, &slotID, &cmd.Command, &cmd.CorrelationID, &cmd.Payload, &cmd.Status, &cmd.RequestedAt, &cmd.AcknowledgedAt, &cmd.ErrorMessage, &cmd.CreatedAt, &cmd.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if slotID.Valid {
		uid, _ := uuid.Parse(slotID.String)
		cmd.SlotID = &uid
	}
	return &cmd, err
}

func (r *DeviceRepo) UpdateCommandStatus(ctx context.Context, corrID string, status domain.DeviceCommandStatus, ackTime *time.Time, errMsg string) error {
	query := `
		UPDATE device_commands
		SET status = $1, acknowledged_at = $2, error_message = $3, updated_at = NOW()
		WHERE correlation_id = $4
	`
	_, err := r.db.Pool.Exec(ctx, query, status, ackTime, errMsg, corrID)
	return err
}

func (r *DeviceRepo) UpdateLockerHeartbeat(ctx context.Context, deviceID string, firmware string, lastSeen time.Time) error {
	query := `
		UPDATE lockers
		SET firmware_version = CASE WHEN $1 != '' THEN $1 ELSE firmware_version END,
		    status = 'online',
		    last_seen_at = $2,
		    updated_at = NOW()
		WHERE device_id = $3
	`
	_, err := r.db.Pool.Exec(ctx, query, firmware, lastSeen, deviceID)
	return err
}

// NotificationRepo
type NotificationRepo struct {
	db *DB
}

func NewNotificationRepo(db *DB) repository.NotificationRepository {
	return &NotificationRepo{db: db}
}

func (r *NotificationRepo) Create(ctx context.Context, n *domain.Notification) error {
	query := `
		INSERT INTO notifications (id, user_id, type, title, body, metadata, read_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Pool.Exec(ctx, query, n.ID, n.UserID, n.Type, n.Title, n.Body, n.Metadata, n.ReadAt, n.CreatedAt)
	return err
}

func (r *NotificationRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Notification, error) {
	query := `
		SELECT id, user_id, type, title, body, metadata, read_at, created_at
		FROM notifications WHERE user_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifs []domain.Notification
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.Metadata, &n.ReadAt, &n.CreatedAt); err != nil {
			return nil, err
		}
		notifs = append(notifs, n)
	}
	return notifs, nil
}

func (r *NotificationRepo) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	query := `UPDATE notifications SET read_at = NOW() WHERE id = $1 AND user_id = $2`
	_, err := r.db.Pool.Exec(ctx, query, id, userID)
	return err
}

func (r *NotificationRepo) RegisterFCMToken(ctx context.Context, dev *domain.FCMDevice) error {
	query := `
		INSERT INTO fcm_devices (id, user_id, token, platform, last_seen_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (token) DO UPDATE 
		SET user_id = EXCLUDED.user_id, platform = EXCLUDED.platform, last_seen_at = NOW()
	`
	_, err := r.db.Pool.Exec(ctx, query, dev.ID, dev.UserID, dev.Token, dev.Platform, dev.LastSeenAt, dev.CreatedAt)
	return err
}

func (r *NotificationRepo) GetFCMTokensByUser(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query := `SELECT token FROM fcm_devices WHERE user_id = $1`
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, nil
}

// PromoRepo
type PromoRepo struct {
	db *DB
}

func NewPromoRepo(db *DB) repository.PromoRepository {
	return &PromoRepo{db: db}
}

func (r *PromoRepo) Create(ctx context.Context, p *domain.Promo) error {
	query := `
		INSERT INTO promos (id, code, title, description, discount_type, discount_value, min_order_amount, max_discount_amount, starts_at, ends_at, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.Pool.Exec(ctx, query, p.ID, p.Code, p.Title, p.Description, p.DiscountType, p.DiscountValue, p.MinOrderAmount, p.MaxDiscountAmount, p.StartsAt, p.EndsAt, p.Status, p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *PromoRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Promo, error) {
	query := `
		SELECT id, code, title, description, discount_type, discount_value, min_order_amount, max_discount_amount, starts_at, ends_at, status, created_at, updated_at
		FROM promos WHERE id = $1
	`
	var p domain.Promo
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Code, &p.Title, &p.Description, &p.DiscountType, &p.DiscountValue, &p.MinOrderAmount, &p.MaxDiscountAmount, &p.StartsAt, &p.EndsAt, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &p, err
}

func (r *PromoRepo) GetByCode(ctx context.Context, code string) (*domain.Promo, error) {
	query := `
		SELECT id, code, title, description, discount_type, discount_value, min_order_amount, max_discount_amount, starts_at, ends_at, status, created_at, updated_at
		FROM promos WHERE LOWER(code) = LOWER($1)
	`
	var p domain.Promo
	err := r.db.Pool.QueryRow(ctx, query, code).Scan(
		&p.ID, &p.Code, &p.Title, &p.Description, &p.DiscountType, &p.DiscountValue, &p.MinOrderAmount, &p.MaxDiscountAmount, &p.StartsAt, &p.EndsAt, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &p, err
}

func (r *PromoRepo) Update(ctx context.Context, p *domain.Promo) error {
	query := `
		UPDATE promos
		SET title = $1, description = $2, discount_type = $3, discount_value = $4, min_order_amount = $5, max_discount_amount = $6, starts_at = $7, ends_at = $8, status = $9, updated_at = NOW()
		WHERE id = $10
	`
	_, err := r.db.Pool.Exec(ctx, query, p.Title, p.Description, p.DiscountType, p.DiscountValue, p.MinOrderAmount, p.MaxDiscountAmount, p.StartsAt, p.EndsAt, p.Status, p.ID)
	return err
}

func (r *PromoRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE promos SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Pool.Exec(ctx, query, status, id)
	return err
}

func (r *PromoRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM promos WHERE id = $1`, id)
	return err
}

func (r *PromoRepo) List(ctx context.Context) ([]domain.Promo, error) {
	query := `
		SELECT id, code, title, description, discount_type, discount_value, min_order_amount, max_discount_amount, starts_at, ends_at, status, created_at, updated_at
		FROM promos ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var promos []domain.Promo
	for rows.Next() {
		var p domain.Promo
		if err := rows.Scan(&p.ID, &p.Code, &p.Title, &p.Description, &p.DiscountType, &p.DiscountValue, &p.MinOrderAmount, &p.MaxDiscountAmount, &p.StartsAt, &p.EndsAt, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		promos = append(promos, p)
	}
	return promos, nil
}

// AuditRepo
type AuditRepo struct {
	db *DB
}

func NewAuditRepo(db *DB) repository.AuditRepository {
	return &AuditRepo{db: db}
}

func (r *AuditRepo) CreateLog(ctx context.Context, l *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, actor_type, actor_id, action, entity_type, entity_id, metadata, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Pool.Exec(ctx, query, l.ID, l.ActorType, l.ActorID, l.Action, l.EntityType, l.EntityID, l.Metadata, l.IPAddress, l.UserAgent, l.CreatedAt)
	return err
}

func (r *AuditRepo) ListLogs(ctx context.Context, limit, offset int) ([]domain.AuditLog, error) {
	query := `
		SELECT id, actor_type, COALESCE(actor_id, ''), action, entity_type, COALESCE(entity_id, ''), metadata, COALESCE(ip_address, ''), COALESCE(user_agent, ''), created_at
		FROM audit_logs ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var l domain.AuditLog
		if err := rows.Scan(&l.ID, &l.ActorType, &l.ActorID, &l.Action, &l.EntityType, &l.EntityID, &l.Metadata, &l.IPAddress, &l.UserAgent, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}
