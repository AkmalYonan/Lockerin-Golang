package service

import (
	"context"
	"time"

	"lockerin-backend/internal/domain"
	"lockerin-backend/internal/repository"

	"github.com/google/uuid"
)

type DeviceService struct {
	deviceRepo repository.DeviceRepository
	lockerRepo repository.LockerRepository
	slotRepo   repository.SlotRepository
}

func NewDeviceService(deviceRepo repository.DeviceRepository, lockerRepo repository.LockerRepository, slotRepo repository.SlotRepository) *DeviceService {
	return &DeviceService{
		deviceRepo: deviceRepo,
		lockerRepo: lockerRepo,
		slotRepo:   slotRepo,
	}
}

func (s *DeviceService) Heartbeat(ctx context.Context, req *domain.HeartbeatRequest) error {
	if req.DeviceID == "" {
		return domain.NewAppError(400, "INVALID_DEVICE_ID", "device_id is required", nil)
	}

	return s.deviceRepo.UpdateLockerHeartbeat(ctx, req.DeviceID, req.FirmwareVersion, time.Now())
}

func (s *DeviceService) SyncStatus(ctx context.Context, req *domain.DeviceStatusRequest) error {
	locker, err := s.lockerRepo.GetByDeviceID(ctx, req.DeviceID)
	if err != nil {
		return domain.NewAppError(404, "DEVICE_NOT_FOUND", "Locker hardware not registered", err)
	}

	for slotCode, status := range req.Slots {
		slot, err := s.slotRepo.GetByLockerAndCode(ctx, locker.ID, slotCode)
		if err == nil {
			// Update status if not in active rental
			if slot.Status != domain.SlotReserved && slot.Status != domain.SlotOccupied {
				_ = s.slotRepo.UpdateStatus(ctx, slot.ID, status, nil)
			}
		}
	}

	return nil
}

func (s *DeviceService) AcknowledgeCommand(ctx context.Context, commandID uuid.UUID, req *domain.CommandAckRequest) error {
	now := time.Now()
	status := domain.CommandAcknowledged
	if req.Status == "FAILED" {
		status = domain.CommandFailed
	}

	corrID := req.CorrelationID
	if corrID == "" {
		cmd, err := s.deviceRepo.GetCommandByID(ctx, commandID)
		if err == nil && cmd != nil {
			corrID = cmd.CorrelationID
		}
	}

	return s.deviceRepo.UpdateCommandStatus(ctx, corrID, status, &now, req.ErrorMessage)
}
