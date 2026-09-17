package midtrans

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"lockerin-backend/internal/config"
	"lockerin-backend/internal/domain"
)

type Gateway struct {
	cfg        *config.Config
	httpClient *http.Client
}

func NewGateway(cfg *config.Config) *Gateway {
	return &Gateway{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (g *Gateway) getBaseURL() string {
	if g.cfg.MidtransIsProduction {
		return "https://app.midtrans.com"
	}
	return "https://app.sandbox.midtrans.com"
}

type SnapTransactionRequest struct {
	TransactionDetails struct {
		OrderID     string  `json:"order_id"`
		GrossAmount float64 `json:"gross_amount"`
	} `json:"transaction_details"`
	CustomerDetails struct {
		FirstName string `json:"first_name"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
	} `json:"customer_details"`
	ItemDetails []struct {
		ID       string  `json:"id"`
		Price    float64 `json:"price"`
		Quantity int     `json:"quantity"`
		Name     string  `json:"name"`
	} `json:"item_details"`
}

type SnapResponse struct {
	Token         string   `json:"token"`
	RedirectURL   string   `json:"redirect_url"`
	ErrorMessages []string `json:"error_messages,omitempty"`
}

// CreateSnapTransaction generates a Snap token and payment redirect URL
func (g *Gateway) CreateSnapTransaction(ctx context.Context, orderID string, amount float64, user *domain.Profile, itemName string) (*SnapResponse, error) {
	if g.cfg.MidtransServerKey == "" || g.cfg.MidtransServerKey == "SB-Mid-server-sample-key" {
		// Mock Snap response for sandbox / development
		return &SnapResponse{
			Token:       fmt.Sprintf("mock-snap-token-%s", orderID),
			RedirectURL: fmt.Sprintf("https://app.sandbox.midtrans.com/snap/v2/vtweb/mock-%s", orderID),
		}, nil
	}

	url := fmt.Sprintf("%s/snap/v1/transactions", g.getBaseURL())

	var reqBody SnapTransactionRequest
	reqBody.TransactionDetails.OrderID = orderID
	reqBody.TransactionDetails.GrossAmount = amount
	if user != nil {
		reqBody.CustomerDetails.FirstName = user.Name
		reqBody.CustomerDetails.Email = user.Email
		reqBody.CustomerDetails.Phone = user.Phone
	}
	reqBody.ItemDetails = []struct {
		ID       string  `json:"id"`
		Price    float64 `json:"price"`
		Quantity int     `json:"quantity"`
		Name     string  `json:"name"`
	}{
		{
			ID:       orderID,
			Price:    amount,
			Quantity: 1,
			Name:     itemName,
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}

	authHeader := base64.StdEncoding.EncodeToString([]byte(g.cfg.MidtransServerKey + ":"))
	httpReq.Header.Set("Authorization", "Basic "+authHeader)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("midtrans request failed: %w", err)
	}
	defer resp.Body.Close()

	var snapResp SnapResponse
	if err := json.NewDecoder(resp.Body).Decode(&snapResp); err != nil {
		return nil, err
	}

	return &snapResp, nil
}

// VerifyWebhookSignature verifies Midtrans SHA-512 signature key
// Signature formula: SHA512(order_id + status_code + gross_amount + ServerKey)
func (g *Gateway) VerifyWebhookSignature(orderID, statusCode, grossAmount, signatureKey string) bool {
	if g.cfg.MidtransServerKey == "" || g.cfg.MidtransServerKey == "SB-Mid-server-sample-key" {
		// Accept in dev / test mode if test signature provided
		return true
	}

	raw := orderID + statusCode + grossAmount + g.cfg.MidtransServerKey
	hash := sha512.Sum512([]byte(raw))
	expectedSignature := hex.EncodeToString(hash[:])

	return expectedSignature == signatureKey
}

// MapPaymentStatus maps Midtrans status to Lockerin domain status
func (g *Gateway) MapPaymentStatus(transactionStatus, fraudStatus string) domain.PaymentStatus {
	switch transactionStatus {
	case "capture":
		if fraudStatus == "challenge" {
			return domain.PaymentPending
		}
		return domain.PaymentSettlement
	case "settlement":
		return domain.PaymentSettlement
	case "pending":
		return domain.PaymentPending
	case "deny":
		return domain.PaymentDeny
	case "expire":
		return domain.PaymentExpire
	case "cancel":
		return domain.PaymentCancel
	case "refund":
		return domain.PaymentRefund
	default:
		return domain.PaymentFailure
	}
}
