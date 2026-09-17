package domain

import (
	"time"

	"github.com/google/uuid"
)

type LockerStatus string

const (
	LockerOnline      LockerStatus = "online"
	LockerOffline     LockerStatus = "offline"
	LockerMaintenance LockerStatus = "maintenance"
)

type SlotStatus string

const (
	SlotAvailable   SlotStatus = "available"
	SlotReserved    SlotStatus = "reserved"
	SlotOccupied    SlotStatus = "occupied"
	SlotMaintenance SlotStatus = "maintenance"
	SlotOffline     SlotStatus = "offline"
)

type SlotSize string

const (
	SizeSmall  SlotSize = "Small"
	SizeMedium SlotSize = "Medium"
	SizeLarge  SlotSize = "Large"
	SizeXL     SlotSize = "ExtraLarge"
)

type Locker struct {
	ID              uuid.UUID    `json:"id"`
	LocationID      uuid.UUID    `json:"location_id"`
	Code            string       `json:"code"`
	Name            string       `json:"name"`
	DeviceID        string       `json:"device_id"`
	Status          LockerStatus `json:"status"`
	FirmwareVersion string       `json:"firmware_version"`
	LastSeenAt      time.Time    `json:"last_seen_at"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`

	Location *Location    `json:"location,omitempty"`
	Slots    []LockerSlot `json:"slots,omitempty"`
}

type LockerSlot struct {
	ID                uuid.UUID  `json:"id"`
	LockerID          uuid.UUID  `json:"locker_id"`
	SlotCode          string     `json:"slot_code"` // e.g. A1, A2, B3, E5
	Size              SlotSize   `json:"size"`
	Status            SlotStatus `json:"status"`
	BasePricePerHour  float64    `json:"base_price_per_hour"`
	CurrentRentalID   *uuid.UUID `json:"current_rental_id,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`

	LockerCode string `json:"locker_code,omitempty"`
}
