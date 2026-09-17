package service

import (
	"context"
	"time"

	"lockerin-backend/internal/domain"
	"lockerin-backend/internal/repository"

	"github.com/google/uuid"
)

type NotificationService struct {
	notifRepo repository.NotificationRepository
}

func NewNotificationService(notifRepo repository.NotificationRepository) *NotificationService {
	return &NotificationService{notifRepo: notifRepo}
}

func (s *NotificationService) RegisterFCMToken(ctx context.Context, userID uuid.UUID, req *domain.RegisterFCMTokenRequest) error {
	dev := &domain.FCMDevice{
		ID:         uuid.New(),
		UserID:     userID,
		Token:      req.Token,
		Platform:   req.Platform,
		LastSeenAt: time.Now(),
		CreatedAt:  time.Now(),
	}
	return s.notifRepo.RegisterFCMToken(ctx, dev)
}

func (s *NotificationService) GetNotifications(ctx context.Context, userID uuid.UUID) ([]domain.Notification, error) {
	return s.notifRepo.ListByUser(ctx, userID)
}

func (s *NotificationService) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	return s.notifRepo.MarkRead(ctx, id, userID)
}
