package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID         uuid.UUID       `json:"id"`
	ActorType  string          `json:"actor_type"` // user, admin, system, iot_device
	ActorID    string          `json:"actor_id,omitempty"`
	Action     string          `json:"action"`
	EntityType string          `json:"entity_type"`
	EntityID   string          `json:"entity_id,omitempty"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
	IPAddress  string          `json:"ip_address,omitempty"`
	UserAgent  string          `json:"user_agent,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

type AdminDashboardStats struct {
	TotalLocations   int     `json:"total_locations"`
	TotalLockers     int     `json:"total_lockers"`
	TotalSlots       int     `json:"total_slots"`
	AvailableSlots   int     `json:"available_slots"`
	OccupiedSlots    int     `json:"occupied_slots"`
	MaintenanceSlots int     `json:"maintenance_slots"`
	ActiveRentals    int     `json:"active_rentals"`
	TotalUsers       int     `json:"total_users"`
	TotalRevenue     float64 `json:"total_revenue"`
	TodayRevenue     float64 `json:"today_revenue"`
}
