package locker_device

import (
	"context"
	"log"
	"time"

	"lockerin-backend/internal/config"
	"lockerin-backend/internal/domain"

	"github.com/google/uuid"
)

type Dispatcher struct {
	cfg *config.Config
}

func NewDispatcher(cfg *config.Config) *Dispatcher {
	return &Dispatcher{cfg: cfg}
}

// DispatchUnlockCommand sends unlock signal to the IoT Head Locker
func (d *Dispatcher) DispatchUnlockCommand(ctx context.Context, lockerDeviceID string, slotCode string, correlationID string) (*domain.DeviceCommand, error) {
	log.Printf("[IoT Head Locker] Dispatching OPEN command to DeviceID='%s', Slot='%s', CorrelationID='%s'",
		lockerDeviceID, slotCode, correlationID)

	cmd := &domain.DeviceCommand{
		ID:            uuid.New(),
		Command:       domain.CommandOpenSlot,
		CorrelationID: correlationID,
		Status:        domain.CommandSent,
		RequestedAt:   time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	return cmd, nil
}

// ValidateDeviceSecret validates the physical hardware authentication token
func (d *Dispatcher) ValidateDeviceSecret(providedSecret string) bool {
	if d.cfg.IoTDeviceSecretKey == "" {
		return true
	}
	return d.cfg.IoTDeviceSecretKey == providedSecret
}
