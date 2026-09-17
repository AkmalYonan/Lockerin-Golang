package http

import (
	"encoding/json"
	"net/http"

	"lockerin-backend/internal/domain"
	"lockerin-backend/internal/middleware"
	"lockerin-backend/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type PaymentHandler struct {
	payService *service.PaymentService
}

func NewPaymentHandler(payService *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{payService: payService}
}

func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		domain.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil, reqID)
		return
	}

	var req domain.CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil, reqID)
		return
	}

	resp, err := h.payService.CreatePayment(r.Context(), claims.UserID, &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			domain.WriteError(w, appErr.StatusCode, appErr.Code, appErr.Message, nil, reqID)
			return
		}
		domain.WriteError(w, http.StatusInternalServerError, "PAYMENT_INITIATION_FAILED", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusCreated, "Transaksi pembayaran berhasil dibuat", resp, reqID)
}

func (h *PaymentHandler) GetPaymentByID(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid UUID", nil, reqID)
		return
	}

	payment, err := h.payService.GetPaymentByID(r.Context(), id)
	if err != nil {
		domain.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Payment not found", nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Detail pembayaran ditemukan", payment, reqID)
}

func (h *PaymentHandler) MidtransWebhook(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	var notif domain.MidtransWebhookNotification
	if err := json.NewDecoder(r.Body).Decode(&notif); err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Invalid JSON body", nil, reqID)
		return
	}

	err := h.payService.ProcessMidtransWebhook(r.Context(), &notif)
	if err != nil {
		domain.WriteError(w, http.StatusBadRequest, "WEBHOOK_FAILED", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Webhook processed successfully", nil, reqID)
}
