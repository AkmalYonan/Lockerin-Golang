package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type OperatingHours struct {
	Open      string `json:"open"`
	Close     string `json:"close"`
	Is24Hours bool   `json:"is_24_hours"`
}

type Location struct {
	ID             uuid.UUID       `json:"id"`
	Code           string          `json:"code"`
	Name           string          `json:"name"`
	Address        string          `json:"address"`
	City           string          `json:"city"`
	Latitude       float64         `json:"latitude"`
	Longitude      float64         `json:"longitude"`
	Status         string          `json:"status"` // active, inactive, maintenance
	OperatingHours json.RawMessage `json:"operating_hours"`
	ImageURL       string          `json:"image_url,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`

	// Derived / Aggregated properties
	TotalLockers     int `json:"total_lockers,omitempty"`
	AvailableSlots   int `json:"available_slots,omitempty"`
	OccupiedSlots    int `json:"occupied_slots,omitempty"`
	MaintenanceSlots int `json:"maintenance_slots,omitempty"`
}

type LocationAvailability struct {
	LocationID     uuid.UUID `json:"location_id"`
	LocationName   string    `json:"location_name"`
	TotalSlots     int       `json:"total_slots"`
	AvailableSlots int       `json:"available_slots"`
	OccupiedSlots  int       `json:"occupied_slots"`
	OfflineSlots   int       `json:"offline_slots"`
	SmallSlots     int       `json:"small_slots_available"`
	MediumSlots    int       `json:"medium_slots_available"`
	LargeSlots     int       `json:"large_slots_available"`
	IsOpenNow      bool      `json:"is_open_now"`
}
