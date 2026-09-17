package worker

import (
	"context"
	"log"
	"time"

	"lockerin-backend/internal/repository"
)

type ExpiryWorker struct {
	rentalRepo repository.RentalRepository
	slotRepo   repository.SlotRepository
	interval   time.Duration
	stopChan   chan struct{}
}

func NewExpiryWorker(rentalRepo repository.RentalRepository, slotRepo repository.SlotRepository, interval time.Duration) *ExpiryWorker {
	if interval <= 0 {
		interval = 1 * time.Minute
	}
	return &ExpiryWorker{
		rentalRepo: rentalRepo,
		slotRepo:   slotRepo,
		interval:   interval,
		stopChan:   make(chan struct{}),
	}
}

func (w *ExpiryWorker) Start() {
	ticker := time.NewTicker(w.interval)
	go func() {
		log.Printf("[ExpiryWorker] Started reservation cleaner routine (interval: %v)", w.interval)
		for {
			select {
			case <-ticker.C:
				w.cleanupOverdueReservations()
			case <-w.stopChan:
				ticker.Stop()
				log.Println("[ExpiryWorker] Stopped reservation cleaner routine")
				return
			}
		}
	}()
}

func (w *ExpiryWorker) Stop() {
	close(w.stopChan)
}

func (w *ExpiryWorker) cleanupOverdueReservations() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	expired, err := w.rentalRepo.ExpireOverdueReservations(ctx, time.Now())
	if err != nil {
		log.Printf("[ExpiryWorker] Error checking expired reservations: %v", err)
		return
	}

	if len(expired) > 0 {
		log.Printf("[ExpiryWorker] Expired %d overdue reservation(s) and freed slots.", len(expired))
	}
}
