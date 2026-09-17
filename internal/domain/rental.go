package domain

import (
	"time"

	"github.com/google/uuid"
)

type RentalStatus string

const (
	RentalDraft           RentalStatus = "draft"
	RentalReserved        RentalStatus = "reserved"
	RentalAwaitingPayment RentalStatus = "awaiting_payment"
	RentalPaid            RentalStatus = "paid"
	RentalActive          RentalStatus = "active"
	RentalCompleted       RentalStatus = "completed"
	RentalExpired         RentalStatus = "expired"
	RentalCancelled       RentalStatus = "cancelled"
)

type Rental struct {
	ID                    uuid.UUID    `json:"id"`
	UserID                uuid.UUID    `json:"user_id"`
	LocationID            uuid.UUID    `json:"location_id"`
	LockerID              uuid.UUID    `json:"locker_id"`
	SlotID                uuid.UUID    `json:"slot_id"`
	Status                RentalStatus `json:"status"`
	DurationHours         int          `json:"duration_hours"`
	BasePriceSnapshot     float64      `json:"base_price_snapshot"`
	DiscountAmount        float64      `json:"discount_amount"`
	TotalAmount           float64      `json:"total_amount"`
	ReservationExpiresAt  *time.Time   `json:"reservation_expires_at,omitempty"`
	StartedAt             *time.Time   `json:"started_at,omitempty"`
	ExpiresAt             *time.Time   `json:"expires_at,omitempty"`
	EndedAt               *time.Time   `json:"ended_at,omitempty"`
	CreatedAt             time.Time    `json:"created_at"`
	UpdatedAt             time.Time    `json:"updated_at"`

	// Joined fields
	UserName     string `json:"user_name,omitempty"`
	LocationName string `json:"location_name,omitempty"`
	LockerCode   string `json:"locker_code,omitempty"`
	SlotCode     string `json:"slot_code,omitempty"`
}

type RentalQuoteRequest struct {
	SlotID        uuid.UUID `json:"slot_id"`
	DurationHours int       `json:"duration_hours"`
	PromoCode     string    `json:"promo_code,omitempty"`
}

type RentalQuoteResponse struct {
	SlotID           uuid.UUID `json:"slot_id"`
	SlotCode         string    `json:"slot_code"`
	Size             string    `json:"size"`
	DurationHours    int       `json:"duration_hours"`
	BasePricePerHour float64   `json:"base_price_per_hour"`
	Subtotal         float64   `json:"subtotal"`
	DiscountAmount   float64   `json:"discount_amount"`
	TotalAmount      float64   `json:"total_amount"`
	PromoApplied     string    `json:"promo_applied,omitempty"`
}

type ReserveRentalRequest struct {
	SlotID        uuid.UUID `json:"slot_id"`
	DurationHours int       `json:"duration_hours"`
	PromoCode     string    `json:"promo_code,omitempty"`
}

type ReserveRentalResponse struct {
	Rental               *Rental   `json:"rental"`
	ReservationExpiresAt time.Time `json:"reservation_expires_at"`
}

type OpenLockerRequest struct {
	SecurityPIN string `json:"security_pin,omitempty"`
}

type OpenLockerResponse struct {
	Success bool   `json:"success"`
	Action  string `json:"action"` // UNLOCKED
	Message string `json:"message"`
}
