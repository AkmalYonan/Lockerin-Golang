package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type DeviceCommandAction string

const (
	CommandOpenSlot    DeviceCommandAction = "OPEN_SLOT"
	CommandLockSlot    DeviceCommandAction = "LOCK_SLOT"
	CommandReboot      DeviceCommandAction = "REBOOT"
	CommandStatusCheck DeviceCommandAction = "STATUS_CHECK"
)

type DeviceCommandStatus string

const (
	CommandPending      DeviceCommandStatus = "PENDING"
	CommandSent         DeviceCommandStatus = "SENT"
	CommandAcknowledged DeviceCommandStatus = "ACKNOWLEDGED"
	CommandFailed       DeviceCommandStatus = "FAILED"
	CommandTimeout      DeviceCommandStatus = "TIMEOUT"
)

type DeviceCommand struct {
	ID             uuid.UUID           `json:"id"`
	LockerID       uuid.UUID           `json:"locker_id"`
	SlotID         *uuid.UUID          `json:"slot_id,omitempty"`
	Command        DeviceCommandAction `json:"command"`
	CorrelationID  string              `json:"correlation_id"`
	Payload        json.RawMessage     `json:"payload,omitempty"`
	Status         DeviceCommandStatus `json:"status"`
	RequestedAt    time.Time           `json:"requested_at"`
	AcknowledgedAt *time.Time          `json:"acknowledged_at,omitempty"`
	ErrorMessage   string              `json:"error_message,omitempty"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

type HeartbeatRequest struct {
	DeviceID        string `json:"device_id"`
	FirmwareVersion string `json:"firmware_version"`
	BatteryLevel    int    `json:"battery_level,omitempty"`
	SignalStrength  int    `json:"signal_strength,omitempty"`
}

type DeviceStatusRequest struct {
	DeviceID string                 `json:"device_id"`
	Slots    map[string]SlotStatus  `json:"slots"` // map of slot_code -> status
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type CommandAckRequest struct {
	CorrelationID string `json:"correlation_id"`
	Status        string `json:"status"` // SUCCESS, FAILED
	ErrorMessage  string `json:"error_message,omitempty"`
}
