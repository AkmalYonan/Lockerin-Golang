package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"lockerin-backend/internal/config"
	"lockerin-backend/internal/domain"
	"lockerin-backend/internal/repository"

	"github.com/google/uuid"
)

type PINService struct {
	secRepo repository.SecurityCodeRepository
	cfg     *config.Config
}

func NewPINService(secRepo repository.SecurityCodeRepository, cfg *config.Config) *PINService {
	return &PINService{
		secRepo: secRepo,
		cfg:     cfg,
	}
}

// GeneratePIN creates a random 6-digit PIN and hashes it with a secure salt
func (s *PINService) GeneratePIN(ctx context.Context, rentalID uuid.UUID) (plaintextPIN string, err error) {
	// Generate random 6-digit PIN
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}
	pin := fmt.Sprintf("%06d", n.Int64()+100000)

	// Generate salt
	saltBytes := make([]byte, 16)
	_, _ = rand.Read(saltBytes)
	salt := hex.EncodeToString(saltBytes)

	// Hash PIN + Salt
	hash := s.hashPIN(pin, salt)

	expiryMinutes := s.cfg.PINExpiryMin
	if expiryMinutes <= 0 {
		expiryMinutes = 1440
	}
	expiresAt := time.Now().Add(time.Duration(expiryMinutes) * time.Minute)

	secCode := &domain.SecurityCode{
		ID:          uuid.New(),
		RentalID:    rentalID,
		PINHash:     hash,
		Salt:        salt,
		Attempts:    0,
		MaxAttempts: s.cfg.PINMaxAttempts,
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.secRepo.Create(ctx, secCode); err != nil {
		return "", err
	}

	return pin, nil
}

func (s *PINService) VerifyPIN(ctx context.Context, rentalID uuid.UUID, plainPIN string) error {
	secCode, err := s.secRepo.GetByRentalID(ctx, rentalID)
	if err != nil {
		return domain.ErrInvalidPIN
	}

	now := time.Now()

	// Check if expired
	if now.After(secCode.ExpiresAt) {
		return domain.ErrPINExpired
	}

	// Check if currently locked out
	if secCode.LockedUntil != nil && now.Before(*secCode.LockedUntil) {
		return domain.ErrPINLocked
	}

	// Verify hash
	expectedHash := s.hashPIN(plainPIN, secCode.Salt)
	if expectedHash != secCode.PINHash {
		// Increment attempts
		_ = s.secRepo.IncrementAttempts(ctx, rentalID)

		if secCode.Attempts+1 >= secCode.MaxAttempts {
			// Lockout for 15 minutes
			lockoutUntil := now.Add(15 * time.Minute)
			_ = s.secRepo.SetLockout(ctx, rentalID, lockoutUntil)
			return domain.ErrPINLocked
		}
		return domain.ErrInvalidPIN
	}

	// Mark as used/verified
	_ = s.secRepo.MarkUsed(ctx, rentalID, now)
	return nil
}

func (s *PINService) hashPIN(pin, salt string) string {
	sum := sha256.Sum256([]byte(pin + ":" + salt))
	return hex.EncodeToString(sum[:])
}
