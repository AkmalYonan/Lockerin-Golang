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

type RentalHandler struct {
	rentalService *service.RentalService
	pinService    *service.PINService
}

func NewRentalHandler(rentalService *service.RentalService, pinService *service.PINService) *RentalHandler {
	return &RentalHandler{
		rentalService: rentalService,
		pinService:    pinService,
	}
}

func (h *RentalHandler) Quote(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	var req domain.RentalQuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil, reqID)
		return
	}

	quote, err := h.rentalService.Quote(r.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			domain.WriteError(w, appErr.StatusCode, appErr.Code, appErr.Message, nil, reqID)
			return
		}
		domain.WriteError(w, http.StatusInternalServerError, "QUOTE_ERROR", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Estimasi biaya sewa berhasil dihitung", quote, reqID)
}

func (h *RentalHandler) Reserve(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		domain.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil, reqID)
		return
	}

	var req domain.ReserveRentalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil, reqID)
		return
	}

	resp, err := h.rentalService.Reserve(r.Context(), claims.UserID, &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			domain.WriteError(w, appErr.StatusCode, appErr.Code, appErr.Message, nil, reqID)
			return
		}
		domain.WriteError(w, http.StatusInternalServerError, "RESERVATION_FAILED", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusCreated, "Slot loker berhasil direservasi. Silahkan lakukan pembayaran.", resp, reqID)
}

func (h *RentalHandler) Start(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	claims := middleware.GetUserClaims(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid UUID", nil, reqID)
		return
	}

	rental, err := h.rentalService.StartRental(r.Context(), id, claims.UserID)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			domain.WriteError(w, appErr.StatusCode, appErr.Code, appErr.Message, nil, reqID)
			return
		}
		domain.WriteError(w, http.StatusInternalServerError, "START_FAILED", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Sewa loker berhasil dimulai", rental, reqID)
}

func (h *RentalHandler) Open(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	claims := middleware.GetUserClaims(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid UUID", nil, reqID)
		return
	}

	var req domain.OpenLockerRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	resp, err := h.rentalService.OpenLocker(r.Context(), id, claims.UserID, req.SecurityPIN)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			domain.WriteError(w, appErr.StatusCode, appErr.Code, appErr.Message, nil, reqID)
			return
		}
		domain.WriteError(w, http.StatusInternalServerError, "OPEN_FAILED", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, resp.Message, resp, reqID)
}

func (h *RentalHandler) Close(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	claims := middleware.GetUserClaims(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid UUID", nil, reqID)
		return
	}

	rental, err := h.rentalService.CloseLocker(r.Context(), id, claims.UserID)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			domain.WriteError(w, appErr.StatusCode, appErr.Code, appErr.Message, nil, reqID)
			return
		}
		domain.WriteError(w, http.StatusInternalServerError, "CLOSE_FAILED", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Sewa loker telah selesai. Terima kasih!", rental, reqID)
}

func (h *RentalHandler) List(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		domain.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil, reqID)
		return
	}

	rentals, err := h.rentalService.GetRentalsByUser(r.Context(), claims.UserID)
	if err != nil {
		domain.WriteError(w, http.StatusInternalServerError, "FETCH_ERROR", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Daftar riwayat sewa ditemukan", rentals, reqID)
}

func (h *RentalHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	claims := middleware.GetUserClaims(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid UUID", nil, reqID)
		return
	}

	rental, err := h.rentalService.GetRentalByID(r.Context(), id, claims.UserID)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			domain.WriteError(w, appErr.StatusCode, appErr.Code, appErr.Message, nil, reqID)
			return
		}
		domain.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Rental not found", nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Detail sewa ditemukan", rental, reqID)
}

func (h *RentalHandler) VerifySecurityCode(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	idStr := chi.URLParam(r, "id")
	rentalID, err := uuid.Parse(idStr)
	if err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid UUID", nil, reqID)
		return
	}

	var req domain.VerifyPINRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON", nil, reqID)
		return
	}

	err = h.pinService.VerifyPIN(r.Context(), rentalID, req.SecurityPIN)
	if err != nil {
		domain.WriteError(w, http.StatusUnauthorized, "PIN_VERIFICATION_FAILED", err.Error(), domain.VerifyPINResponse{
			Success: false,
			Action:  "LOCKED",
			Message: "PIN Salah atau telah terkunci!",
		}, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "PIN Benar! Solenoid terbuka.", domain.VerifyPINResponse{
		Success: true,
		Action:  "UNLOCKED",
		Message: "PIN Valid! Loker berhasil terbuka.",
	}, reqID)
}
