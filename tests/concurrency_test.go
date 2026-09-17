package tests

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"lockerin-backend/internal/config"
	"lockerin-backend/internal/domain"
	"lockerin-backend/internal/integration/locker_device"
	"lockerin-backend/internal/repository/inmemory"
	"lockerin-backend/internal/service"

	"github.com/google/uuid"
)

// TestMandatorySlotConcurrency verifies that when multiple users attempt to reserve
// the EXACT SAME locker slot at the exact same millisecond, only ONE reservation succeeds,
// and all other concurrent requests are rejected (409 Conflict).
func TestMandatorySlotConcurrency(t *testing.T) {
	memStore := inmemory.NewInMemoryStore()
	repos := inmemory.NewRepositories(memStore)
	cfg := &config.Config{
		ReservationTTLMin: 15,
		PINMaxAttempts:    5,
		PINExpiryMin:      1440,
	}

	pinService := service.NewPINService(repos.SecurityCodes, cfg)
	deviceDisp := locker_device.NewDispatcher(cfg)
	rentalService := service.NewRentalService(
		repos.Rentals,
		repos.Slots,
		repos.Lockers,
		repos.Promos,
		pinService,
		deviceDisp,
		cfg,
	)

	// Pick an available slot
	var targetSlot domain.LockerSlot
	for _, slot := range memStore.Slots {
		if slot.Status == domain.SlotAvailable {
			targetSlot = slot
			break
		}
	}

	if targetSlot.ID == uuid.Nil {
		t.Fatal("No available slot found in store")
	}

	concurrentUsers := 10
	var successCount int32
	var failureCount int32
	var wg sync.WaitGroup

	startBarrier := make(chan struct{})

	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		userUUID := uuid.New()

		go func(uid uuid.UUID) {
			defer wg.Done()
			<-startBarrier // wait for all goroutines to sync

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			req := &domain.ReserveRentalRequest{
				SlotID:        targetSlot.ID,
				DurationHours: 3,
			}

			_, err := rentalService.Reserve(ctx, uid, req)
			if err == nil {
				atomic.AddInt32(&successCount, 1)
			} else {
				atomic.AddInt32(&failureCount, 1)
			}
		}(userUUID)
	}

	// Release all goroutines simultaneously
	close(startBarrier)
	wg.Wait()

	t.Logf("Concurrency Result: %d Succeeded, %d Failed out of %d concurrent requests", successCount, failureCount, concurrentUsers)

	if successCount != 1 {
		t.Fatalf("CRITICAL: Expected exactly 1 success for single slot reservation, got: %d", successCount)
	}

	if failureCount != int32(concurrentUsers-1) {
		t.Fatalf("Expected %d failures, got: %d", concurrentUsers-1, failureCount)
	}
}
