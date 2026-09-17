package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type PaymentStatus string

const (
	PaymentPending    PaymentStatus = "pending"
	PaymentSettlement PaymentStatus = "settlement"
	PaymentCapture    PaymentStatus = "capture"
	PaymentDeny       PaymentStatus = "deny"
	PaymentCancel     PaymentStatus = "cancel"
	PaymentExpire     PaymentStatus = "expire"
	PaymentFailure    PaymentStatus = "failure"
	PaymentRefund     PaymentStatus = "refund"
)

type Payment struct {
	ID                    uuid.UUID       `json:"id"`
	RentalID              uuid.UUID       `json:"rental_id"`
	Provider              string          `json:"provider"` // midtrans
	OrderID               string          `json:"order_id"`
	ProviderTransactionID string          `json:"provider_transaction_id,omitempty"`
	Amount                float64         `json:"amount"`
	Status                PaymentStatus   `json:"status"`
	PaymentType           string          `json:"payment_type,omitempty"`
	SnapToken             string          `json:"snap_token,omitempty"`
	SnapRedirectURL       string          `json:"snap_redirect_url,omitempty"`
	RawResponse           json.RawMessage `json:"raw_response,omitempty"`
	PaidAt                *time.Time      `json:"paid_at,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

type CreatePaymentRequest struct {
	RentalID      uuid.UUID `json:"rental_id"`
	PaymentMethod string    `json:"payment_method,omitempty"` // qris, gopay, bank_transfer, bca_va
}

type CreatePaymentResponse struct {
	PaymentID       uuid.UUID `json:"payment_id"`
	OrderID         string    `json:"order_id"`
	Amount          float64   `json:"amount"`
	SnapToken       string    `json:"snap_token,omitempty"`
	SnapRedirectURL string    `json:"snap_redirect_url,omitempty"`
	Status          string    `json:"status"`
}

type MidtransWebhookNotification struct {
	TransactionTime   string `json:"transaction_time"`
	TransactionStatus string `json:"transaction_status"`
	TransactionID     string `json:"transaction_id"`
	StatusMessage     string `json:"status_message"`
	StatusCode        string `json:"status_code"`
	SignatureKey      string `json:"signature_key"`
	PaymentType       string `json:"payment_type"`
	OrderID           string `json:"order_id"`
	GrossAmount       string `json:"gross_amount"`
	FraudStatus       string `json:"fraud_status,omitempty"`
	Currency          string `json:"currency,omitempty"`
}
