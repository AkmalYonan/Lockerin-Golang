package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"lockerin-backend/internal/domain"
	"lockerin-backend/internal/middleware"
	"lockerin-backend/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AdminHandler struct {
	adminService *service.AdminService
}

func NewAdminHandler(adminService *service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	stats, err := h.adminService.GetDashboardStats(r.Context())
	if err != nil {
		domain.WriteError(w, http.StatusInternalServerError, "STATS_ERROR", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Statistik admin dashboard ditemukan", stats, reqID)
}

func (h *AdminHandler) ListLockers(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	lockers, err := h.adminService.ListLockersWithStats(r.Context())
	if err != nil {
		domain.WriteError(w, http.StatusInternalServerError, "FETCH_ERROR", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Daftar unit loker hardware ditemukan", lockers, reqID)
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	users, err := h.adminService.ListUsers(r.Context(), limit, offset)
	if err != nil {
		domain.WriteError(w, http.StatusInternalServerError, "FETCH_ERROR", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Daftar pengguna ditemukan", users, reqID)
}

func (h *AdminHandler) GetUserDetail(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid UUID", nil, reqID)
		return
	}

	detail, err := h.adminService.GetUserDetail(r.Context(), id)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			domain.WriteError(w, appErr.StatusCode, appErr.Code, appErr.Message, nil, reqID)
			return
		}
		domain.WriteError(w, http.StatusNotFound, "USER_NOT_FOUND", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Detail pengguna ditemukan", detail, reqID)
}

func (h *AdminHandler) UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	claims := middleware.GetUserClaims(r.Context())
	actorID := "admin"
	if claims != nil {
		actorID = claims.UserID.String()
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid UUID", nil, reqID)
		return
	}

	var req domain.UpdateUserStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON body", nil, reqID)
		return
	}

	p, err := h.adminService.UpdateUserStatus(r.Context(), actorID, id, req.Status)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			domain.WriteError(w, appErr.StatusCode, appErr.Code, appErr.Message, nil, reqID)
			return
		}
		domain.WriteError(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error(), nil, reqID)
		return
	}

	respData := map[string]interface{}{
		"id":         p.ID,
		"email":      p.Email,
		"status":     p.Status,
		"updated_at": p.UpdatedAt,
	}

	domain.WriteSuccess(w, http.StatusOK, "Status pengguna berhasil diperbarui", respData, reqID)
}

func (h *AdminHandler) GetPricing(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	pricing, err := h.adminService.GetPricing(r.Context())
	if err != nil {
		domain.WriteError(w, http.StatusInternalServerError, "FETCH_ERROR", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Tarif sewa loker aktif ditemukan", pricing, reqID)
}

func (h *AdminHandler) UpdatePricing(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	claims := middleware.GetUserClaims(r.Context())
	actorID := "admin"
	if claims != nil {
		actorID = claims.UserID.String()
	}

	var req domain.UpdatePricingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON body", nil, reqID)
		return
	}

	tiers := req.Pricing
	if len(tiers) == 0 && len(req.Prices) > 0 {
		for sz, prc := range req.Prices {
			tiers = append(tiers, domain.PricingTierItem{
				SlotSize:      string(sz),
				HourlyRate:    prc,
				DepositAmount: 10000,
			})
		}
	}

	if err := h.adminService.UpdatePricing(r.Context(), actorID, tiers); err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			domain.WriteError(w, appErr.StatusCode, appErr.Code, appErr.Message, nil, reqID)
			return
		}
		domain.WriteError(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Tarif sewa loker berhasil diupdate", nil, reqID)
}

func (h *AdminHandler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	var loc domain.Location
	if err := json.NewDecoder(r.Body).Decode(&loc); err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON", nil, reqID)
		return
	}

	loc.ID = uuid.New()
	loc.CreatedAt = time.Now()
	loc.UpdatedAt = time.Now()

	if err := h.adminService.CreateLocation(r.Context(), &loc); err != nil {
		domain.WriteError(w, http.StatusInternalServerError, "CREATE_FAILED", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusCreated, "Stasiun loker dan unit 25 slot berhasil dibuat otomatis", loc, reqID)
}

func (h *AdminHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid UUID", nil, reqID)
		return
	}

	var loc domain.Location
	if err := json.NewDecoder(r.Body).Decode(&loc); err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON", nil, reqID)
		return
	}

	loc.ID = id
	if err := h.adminService.UpdateLocation(r.Context(), &loc); err != nil {
		domain.WriteError(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Lokasi loker berhasil diupdate", loc, reqID)
}

func (h *AdminHandler) UpdateLocationStatus(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	claims := middleware.GetUserClaims(r.Context())
	actorID := "admin"
	if claims != nil {
		actorID = claims.UserID.String()
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid UUID", nil, reqID)
		return
	}

	var req domain.UpdateLocationStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON", nil, reqID)
		return
	}

	if err := h.adminService.UpdateLocationStatus(r.Context(), actorID, id, req.Status); err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			domain.WriteError(w, appErr.StatusCode, appErr.Code, appErr.Message, nil, reqID)
			return
		}
		domain.WriteError(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Status lokasi berhasil diupdate", map[string]interface{}{
		"id":     id,
		"status": req.Status,
	}, reqID)
}

func (h *AdminHandler) DeleteLocation(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid UUID", nil, reqID)
		return
	}

	if err := h.adminService.DeleteLocation(r.Context(), id); err != nil {
		domain.WriteError(w, http.StatusInternalServerError, "DELETE_FAILED", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Lokasi loker berhasil dihapus", nil, reqID)
}

func (h *AdminHandler) CreatePromo(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	var req domain.CreatePromoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON", nil, reqID)
		return
	}

	promo, err := h.adminService.CreatePromo(r.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			domain.WriteError(w, appErr.StatusCode, appErr.Code, appErr.Message, nil, reqID)
			return
		}
		domain.WriteError(w, http.StatusInternalServerError, "CREATE_FAILED", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusCreated, "Promo berhasil dibuat", promo, reqID)
}

func (h *AdminHandler) ListPromos(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	promos, err := h.adminService.ListPromos(r.Context())
	if err != nil {
		domain.WriteError(w, http.StatusInternalServerError, "FETCH_ERROR", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Daftar promo ditemukan", promos, reqID)
}

func (h *AdminHandler) UpdatePromoStatus(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid UUID", nil, reqID)
		return
	}

	var req domain.UpdatePromoStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON body", nil, reqID)
		return
	}

	if err := h.adminService.UpdatePromoStatus(r.Context(), id, &req); err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			domain.WriteError(w, appErr.StatusCode, appErr.Code, appErr.Message, nil, reqID)
			return
		}
		domain.WriteError(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Status promo berhasil diupdate", nil, reqID)
}

func (h *AdminHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	trxs, err := h.adminService.ListTransactions(r.Context(), limit, offset)
	if err != nil {
		domain.WriteError(w, http.StatusInternalServerError, "FETCH_ERROR", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Daftar transaksi pembayaran dan traceability ditemukan", trxs, reqID)
}

func (h *AdminHandler) GetTransactionDetail(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid UUID", nil, reqID)
		return
	}

	trx, err := h.adminService.GetTransactionDetail(r.Context(), id)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			domain.WriteError(w, appErr.StatusCode, appErr.Code, appErr.Message, nil, reqID)
			return
		}
		domain.WriteError(w, http.StatusNotFound, "TRANSACTION_NOT_FOUND", "Transaction not found", nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Detail transaksi ditemukan", trx, reqID)
}

func (h *AdminHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	logs, err := h.adminService.ListAuditLogs(r.Context(), limit, offset)
	if err != nil {
		domain.WriteError(w, http.StatusInternalServerError, "FETCH_ERROR", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Daftar audit logs ditemukan", logs, reqID)
}

func (h *AdminHandler) DispatchDeviceCommand(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	claims := middleware.GetUserClaims(r.Context())
	actorID := "admin"
	if claims != nil {
		actorID = claims.UserID.String()
	}

	var req domain.AdminDeviceCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON body", nil, reqID)
		return
	}

	cmd, err := h.adminService.DispatchAdminCommand(r.Context(), actorID, &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			domain.WriteError(w, appErr.StatusCode, appErr.Code, appErr.Message, nil, reqID)
			return
		}
		domain.WriteError(w, http.StatusInternalServerError, "COMMAND_DISPATCH_FAILED", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Perintah kontrol perangkat berhasil dikirim", cmd, reqID)
}
