package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"lockerin-backend/internal/domain"
	"lockerin-backend/internal/integration/fcm"
	"lockerin-backend/internal/integration/midtrans"
	"lockerin-backend/internal/repository"

	"github.com/google/uuid"
)

type PaymentService struct {
	payRepo    repository.PaymentRepository
	rentalRepo repository.RentalRepository
	profRepo   repository.ProfileRepository
	notifRepo  repository.NotificationRepository
	slotRepo   repository.SlotRepository
	midtrans   *midtrans.Gateway
	pinService *PINService
	fcm        *fcm.Notifier
}

func NewPaymentService(
	payRepo repository.PaymentRepository,
	rentalRepo repository.RentalRepository,
	profRepo repository.ProfileRepository,
	notifRepo repository.NotificationRepository,
	slotRepo repository.SlotRepository,
	midtrans *midtrans.Gateway,
	pinService *PINService,
	fcm *fcm.Notifier,
) *PaymentService {
	return &PaymentService{
		payRepo:    payRepo,
		rentalRepo: rentalRepo,
		profRepo:   profRepo,
		notifRepo:  notifRepo,
		slotRepo:   slotRepo,
		midtrans:   midtrans,
		pinService: pinService,
		fcm:        fcm,
	}
}

func (s *PaymentService) CreatePayment(ctx context.Context, userID uuid.UUID, req *domain.CreatePaymentRequest) (*domain.CreatePaymentResponse, error) {
	rent, err := s.rentalRepo.GetByID(ctx, req.RentalID)
	if err != nil {
		return nil, domain.NewAppError(404, "RENTAL_NOT_FOUND", "Rental not found", err)
	}

	if rent.UserID != userID {
		return nil, domain.NewAppError(403, "FORBIDDEN", "Unauthorized access to rental", domain.ErrForbidden)
	}

	if rent.Status != domain.RentalReserved && rent.Status != domain.RentalAwaitingPayment {
		return nil, domain.NewAppError(400, "INVALID_STATE", fmt.Sprintf("Cannot create payment for rental in '%s' status", rent.Status), domain.ErrInvalidRentalState)
	}

	user, _ := s.profRepo.GetByID(ctx, userID)

	orderID := fmt.Sprintf("LOCKERIN-%s-%d", rent.ID.String()[:8], time.Now().Unix())
	itemName := fmt.Sprintf("Sewa Loker %s (%d Jam)", rent.SlotCode, rent.DurationHours)

	snapResp, err := s.midtrans.CreateSnapTransaction(ctx, orderID, rent.TotalAmount, user, itemName)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate payment: %w", err)
	}

	paymentID := uuid.New()
	payment := &domain.Payment{
		ID:              paymentID,
		RentalID:        rent.ID,
		Provider:        "midtrans",
		OrderID:         orderID,
		Amount:          rent.TotalAmount,
		Status:          domain.PaymentPending,
		PaymentType:     req.PaymentMethod,
		SnapToken:       snapResp.Token,
		SnapRedirectURL: snapResp.RedirectURL,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.payRepo.Create(ctx, payment); err != nil {
		return nil, err
	}

	// Update rental status to awaiting_payment
	_ = s.rentalRepo.UpdateStatus(ctx, rent.ID, domain.RentalAwaitingPayment)

	return &domain.CreatePaymentResponse{
		PaymentID:       paymentID,
		OrderID:         orderID,
		Amount:          rent.TotalAmount,
		SnapToken:       snapResp.Token,
		SnapRedirectURL: snapResp.RedirectURL,
		Status:          string(domain.PaymentPending),
	}, nil
}

func (s *PaymentService) ProcessMidtransWebhook(ctx context.Context, notif *domain.MidtransWebhookNotification) error {
	// 1. Verify Signature
	if !s.midtrans.VerifyWebhookSignature(notif.OrderID, notif.StatusCode, notif.GrossAmount, notif.SignatureKey) {
		log.Printf("[Webhook] Invalid signature key for OrderID: %s", notif.OrderID)
		return domain.ErrInvalidSignature
	}

	// 2. Find Payment Record
	payment, err := s.payRepo.GetByOrderID(ctx, notif.OrderID)
	if err != nil {
		log.Printf("[Webhook] Payment record not found for OrderID: %s", notif.OrderID)
		return domain.ErrNotFound
	}

	// 3. Idempotency check: if payment already settled, return success
	if payment.Status == domain.PaymentSettlement {
		log.Printf("[Webhook] OrderID %s is already settled. Skipping duplicate event.", notif.OrderID)
		return nil
	}

	// 4. Map provider status
	newStatus := s.midtrans.MapPaymentStatus(notif.TransactionStatus, notif.FraudStatus)
	payment.Status = newStatus
	payment.ProviderTransactionID = notif.TransactionID
	payment.PaymentType = notif.PaymentType
	raw, _ := json.Marshal(notif)
	payment.RawResponse = raw

	now := time.Now()
	if newStatus == domain.PaymentSettlement {
		payment.PaidAt = &now
	}

	if err := s.payRepo.Update(ctx, payment); err != nil {
		return err
	}

	// 5. Handle Business state transitions based on settlement
	if newStatus == domain.PaymentSettlement {
		// Update rental to paid
		_ = s.rentalRepo.UpdateStatus(ctx, payment.RentalID, domain.RentalPaid)

		// Generate Security PIN for the user
		pin, err := s.pinService.GeneratePIN(ctx, payment.RentalID)
		if err == nil {
			log.Printf("[Security PIN] One-time PIN generated for RentalID: %s (hash stored)", payment.RentalID)
		}

		// Create in-app notification
		rent, _ := s.rentalRepo.GetByID(ctx, payment.RentalID)
		if rent != nil {
			notifMsg := fmt.Sprintf("Pembayaran sewa loker %s berhasil! PIN Keamanan Anda adalah: %s", rent.SlotCode, pin)
			notifObj := &domain.Notification{
				ID:        uuid.New(),
				UserID:    rent.UserID,
				Type:      "payment",
				Title:     "Pembayaran Berhasil",
				Body:      notifMsg,
				CreatedAt: time.Now(),
			}
			_ = s.notifRepo.Create(ctx, notifObj)

			// Push FCM
			tokens, _ := s.notifRepo.GetFCMTokensByUser(ctx, rent.UserID)
			_ = s.fcm.SendPushNotification(ctx, tokens, "Pembayaran Berhasil", notifMsg, nil)
		}
	} else if newStatus == domain.PaymentExpire || newStatus == domain.PaymentCancel || newStatus == domain.PaymentDeny {
		_ = s.rentalRepo.UpdateStatus(ctx, payment.RentalID, domain.RentalCancelled)
		// Free slot
		rent, _ := s.rentalRepo.GetByID(ctx, payment.RentalID)
		if rent != nil {
			_ = s.slotRepo.UpdateStatus(ctx, rent.SlotID, domain.SlotAvailable, nil)
		}
	}

	return nil
}

func (s *PaymentService) GetPaymentByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	return s.payRepo.GetByID(ctx, id)
}
