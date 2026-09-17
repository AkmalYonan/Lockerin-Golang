package tests

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"testing"
	"time"

	"lockerin-backend/internal/auth"
	"lockerin-backend/internal/config"
	"lockerin-backend/internal/domain"
	"lockerin-backend/internal/integration/midtrans"
	"lockerin-backend/internal/integration/supabase"
	"lockerin-backend/internal/repository/inmemory"
	"lockerin-backend/internal/service"
)

func TestAuthFlow(t *testing.T) {
	memStore := inmemory.NewInMemoryStore()
	repos := inmemory.NewRepositories(memStore)
	cfg := &config.Config{JWTSecret: "test-jwt-secret"}

	jwtManager := auth.NewJWTManager(cfg.JWTSecret, 1*time.Hour)
	supabaseClient := supabase.NewClient(cfg)
	authService := service.NewAuthService(repos.Profiles, jwtManager, supabaseClient)

	ctx := context.Background()

	// 1. Test Register
	regReq := &domain.RegisterRequest{
		Email:    "newuser@lockerin.id",
		Username: "newuser",
		Password: "SecurePassword123!",
		Name:     "New User Testing",
		Phone:    "08123456789",
	}

	regResp, err := authService.Register(ctx, regReq)
	if err != nil {
		t.Fatalf("Failed to register: %v", err)
	}

	if regResp.AccessToken == "" {
		t.Fatal("Expected JWT token upon registration")
	}

	// 2. Test Login
	loginReq := &domain.LoginRequest{
		EmailOrUsername: "newuser@lockerin.id",
		Password:        "SecurePassword123!",
	}

	loginResp, err := authService.Login(ctx, loginReq)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	if loginResp.Profile.Email != "newuser@lockerin.id" {
		t.Fatalf("Expected email 'newuser@lockerin.id', got: %s", loginResp.Profile.Email)
	}

	// 3. Test JWT Token Verification
	claims, err := jwtManager.VerifyToken(loginResp.AccessToken)
	if err != nil {
		t.Fatalf("Failed to verify token: %v", err)
	}

	if claims.Email != "newuser@lockerin.id" {
		t.Fatalf("Expected claims email 'newuser@lockerin.id', got: %s", claims.Email)
	}
}

func TestPINSecurityAndLockout(t *testing.T) {
	memStore := inmemory.NewInMemoryStore()
	repos := inmemory.NewRepositories(memStore)
	cfg := &config.Config{
		PINMaxAttempts: 3,
		PINExpiryMin:   60,
	}

	pinService := service.NewPINService(repos.SecurityCodes, cfg)
	ctx := context.Background()

	// Generate PIN
	rentalID := domain.Profile{}.ID // dummy
	plainPIN, err := pinService.GeneratePIN(ctx, rentalID)
	if err != nil {
		t.Fatalf("Failed to generate PIN: %v", err)
	}

	if len(plainPIN) != 6 {
		t.Fatalf("Expected 6-digit PIN, got: %s", plainPIN)
	}

	// Verify Wrong PIN increments attempts
	err = pinService.VerifyPIN(ctx, rentalID, "000000")
	if err == nil {
		t.Fatal("Expected error on wrong PIN")
	}

	// Attempt 2 Wrong
	_ = pinService.VerifyPIN(ctx, rentalID, "111111")
	// Attempt 3 Wrong -> triggers Lockout
	err = pinService.VerifyPIN(ctx, rentalID, "222222")
	if err != domain.ErrPINLocked {
		t.Fatalf("Expected ErrPINLocked after max attempts, got: %v", err)
	}

	// Verify Correct PIN is now blocked because of lockout
	err = pinService.VerifyPIN(ctx, rentalID, plainPIN)
	if err != domain.ErrPINLocked {
		t.Fatalf("Expected ErrPINLocked even with correct PIN during lockout, got: %v", err)
	}
}

func TestPaymentWebhookIdempotencyAndFlow(t *testing.T) {
	memStore := inmemory.NewInMemoryStore()
	repos := inmemory.NewRepositories(memStore)
	cfg := &config.Config{
		MidtransServerKey: "test-server-key",
		PINMaxAttempts:    5,
		PINExpiryMin:      1440,
	}

	pinService := service.NewPINService(repos.SecurityCodes, cfg)
	midtransGateway := midtrans.NewGateway(cfg)
	rentalService := service.NewRentalService(
		repos.Rentals,
		repos.Slots,
		repos.Lockers,
		repos.Promos,
		pinService,
		nil,
		cfg,
	)

	paymentService := service.NewPaymentService(
		repos.Payments,
		repos.Rentals,
		repos.Profiles,
		repos.Notifications,
		repos.Slots,
		midtransGateway,
		pinService,
		nil,
	)

	ctx := context.Background()

	// 1. Reserve slot
	var targetSlot domain.LockerSlot
	for _, slot := range memStore.Slots {
		if slot.Status == domain.SlotAvailable {
			targetSlot = slot
			break
		}
	}

	userID := memStore.Profiles[domain.Profile{}.ID].ID
	for _, u := range memStore.Profiles {
		userID = u.ID
		break
	}

	res, err := rentalService.Reserve(ctx, userID, &domain.ReserveRentalRequest{
		SlotID:        targetSlot.ID,
		DurationHours: 2,
	})
	if err != nil {
		t.Fatalf("Failed to reserve: %v", err)
	}

	// 2. Create Payment
	payResp, err := paymentService.CreatePayment(ctx, userID, &domain.CreatePaymentRequest{
		RentalID:      res.Rental.ID,
		PaymentMethod: "qris",
	})
	if err != nil {
		t.Fatalf("Failed to create payment: %v", err)
	}

	// 3. Simulate Webhook Settlement with correct SHA-512 signature
	rawSig := payResp.OrderID + "200" + "12000.00" + cfg.MidtransServerKey
	hashSig := sha512.Sum512([]byte(rawSig))
	validSignature := hex.EncodeToString(hashSig[:])

	webhookNotif := &domain.MidtransWebhookNotification{
		OrderID:           payResp.OrderID,
		TransactionStatus: "settlement",
		StatusCode:        "200",
		GrossAmount:       "12000.00",
		PaymentType:       "qris",
		SignatureKey:      validSignature,
	}

	err = paymentService.ProcessMidtransWebhook(ctx, webhookNotif)
	if err != nil {
		t.Fatalf("Failed to process webhook: %v", err)
	}

	// Check Rental transitioned to 'paid'
	updatedRent, _ := repos.Rentals.GetByID(ctx, res.Rental.ID)
	if updatedRent.Status != domain.RentalPaid {
		t.Fatalf("Expected status 'paid', got: %s", updatedRent.Status)
	}

	// Check Security Code generated
	secCode, err := repos.SecurityCodes.GetByRentalID(ctx, res.Rental.ID)
	if err != nil || secCode == nil {
		t.Fatal("Expected security code to be generated upon settlement")
	}

	// 4. Test Idempotency: Duplicate webhook does not error
	err = paymentService.ProcessMidtransWebhook(ctx, webhookNotif)
	if err != nil {
		t.Fatalf("Duplicate webhook processing failed: %v", err)
	}
}
