package domain

import (
	"time"

	"github.com/google/uuid"
)

type SecurityCode struct {
	ID          uuid.UUID  `json:"id"`
	RentalID    uuid.UUID  `json:"rental_id"`
	PINHash     string     `json:"-"` // Never expose PIN hash directly
	Salt        string     `json:"-"`
	Attempts    int        `json:"attempts"`
	MaxAttempts int        `json:"max_attempts"`
	LockedUntil *time.Time `json:"locked_until,omitempty"`
	ExpiresAt   time.Time  `json:"expires_at"`
	UsedAt      *time.Time `json:"used_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type VerifyPINRequest struct {
	StationID   string `json:"station_id,omitempty"`
	LockerCode  string `json:"locker_code,omitempty"`
	SecurityPIN string `json:"security_pin"`
}

type VerifyPINResponse struct {
	Success bool   `json:"success"`
	Action  string `json:"action"` // UNLOCKED or LOCKED
	Message string `json:"message"`
}
