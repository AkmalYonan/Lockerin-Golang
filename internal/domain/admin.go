package domain

import (
	"time"

	"github.com/google/uuid"
)

type LockerAdminItem struct {
	ID              uuid.UUID    `json:"id"`
	DeviceID        string       `json:"device_id"`
	Code            string       `json:"code"`
	Name            string       `json:"name"`
	LocationID      uuid.UUID    `json:"location_id"`
	LocationName    string       `json:"location_name"`
	Status          LockerStatus `json:"status"`
	FirmwareVersion string       `json:"firmware_version"`
	BatteryLevel    int          `json:"battery_level"`
	SignalStrength  int          `json:"signal_strength"`
	LastHeartbeat   time.Time    `json:"last_heartbeat"`
	TotalSlots      int          `json:"total_slots"`
	AvailableSlots  int          `json:"available_slots"`
	OccupiedSlots   int          `json:"occupied_slots"`
}

type PricingTier struct {
	Size             SlotSize `json:"size"`
	PricePerHour     float64  `json:"price_per_hour"`
	Description      string   `json:"description,omitempty"`
	RecommendedUsage string   `json:"recommended_usage,omitempty"`
}

type PricingTierItem struct {
	ID            uuid.UUID `json:"id,omitempty"`
	SlotSize      string    `json:"slot_size"` // small, medium, large
	HourlyRate    float64   `json:"hourly_rate"`
	DepositAmount float64   `json:"deposit_amount"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
}

type UpdatePricingRequest struct {
	Pricing []PricingTierItem    `json:"pricing,omitempty"`
	Prices  map[SlotSize]float64 `json:"prices,omitempty"` // fallback for legacy
}

type UpdateUserStatusRequest struct {
	Status string `json:"status"` // active, suspended, deactivated
}

type UpdatePromoStatusRequest struct {
	IsActive *bool  `json:"is_active,omitempty"`
	Status   string `json:"status,omitempty"` // active, inactive, expired
}

type UpdateLocationStatusRequest struct {
	Status string `json:"status"` // active, inactive, maintenance
}

type AdminDeviceCommandRequest struct {
	DeviceID    string                 `json:"device_id"`
	CommandType string                 `json:"command_type"` // SOLENOID_UNLOCK, REBOOT, STATUS_CHECK
	Payload     map[string]interface{} `json:"payload"`      // e.g. {"slot_code": "A2", "reason": "Maintenance override"}
}

type AdminUserDetailResponse struct {
	Profile  *Profile  `json:"profile"`
	Rentals  []Rental  `json:"rentals"`
	Payments []Payment `json:"payments"`
}

type AdminAuditEvent struct {
	Step   string `json:"step"`
	Time   string `json:"time"`
	Detail string `json:"detail"`
}

type AdminTransactionItem struct {
	ID              uuid.UUID         `json:"id"`
	OrderID         string            `json:"order_id"`
	Amount          float64           `json:"amount"`
	PaymentMethod   string            `json:"payment_method"`
	Status          string            `json:"status"`
	MidtransID      string            `json:"midtrans_id,omitempty"`
	SnapToken       string            `json:"snap_token,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	SettledAt       *time.Time        `json:"settled_at,omitempty"`
	UserName        string            `json:"user_name"`
	UserEmail       string            `json:"user_email"`
	UserPhone       string            `json:"user_phone"`
	LocationName    string            `json:"location_name"`
	LocationAddress string            `json:"location_address"`
	LockerCode      string            `json:"locker_code"`
	SlotCode        string            `json:"slot_code"`
	SlotSize        string            `json:"slot_size"`
	SecurityPIN     string            `json:"security_pin,omitempty"`
	AuditEvents     []AdminAuditEvent `json:"audit_events"`
}
