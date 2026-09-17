package http

import (
	"net/http"
	"strconv"

	"lockerin-backend/internal/domain"
	"lockerin-backend/internal/middleware"
	"lockerin-backend/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type LocationHandler struct {
	locService *service.LocationService
}

func NewLocationHandler(locService *service.LocationService) *LocationHandler {
	return &LocationHandler{locService: locService}
}

func (h *LocationHandler) ListLocations(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	locations, err := h.locService.ListLocations(r.Context(), limit, offset)
	if err != nil {
		domain.WriteError(w, http.StatusInternalServerError, "FETCH_ERROR", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Daftar lokasi ditemukan", locations, reqID)
}

func (h *LocationHandler) GetLocation(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid UUID format", nil, reqID)
		return
	}

	loc, err := h.locService.GetLocation(r.Context(), id)
	if err != nil {
		domain.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Location not found", nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Detail lokasi ditemukan", loc, reqID)
}

func (h *LocationHandler) GetAvailability(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid UUID format", nil, reqID)
		return
	}

	avail, err := h.locService.GetLocationAvailability(r.Context(), id)
	if err != nil {
		domain.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Location availability not found", nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Ketersediaan loker ditemukan", avail, reqID)
}

// LockerHandler
type LockerHandler struct {
	lockerService *service.LockerService
}

func NewLockerHandler(lockerService *service.LockerService) *LockerHandler {
	return &LockerHandler{lockerService: lockerService}
}

func (h *LockerHandler) GetLocker(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid UUID format", nil, reqID)
		return
	}

	locker, err := h.lockerService.GetLocker(r.Context(), id)
	if err != nil {
		domain.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Locker not found", nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Detail unit loker ditemukan", locker, reqID)
}

func (h *LockerHandler) GetSlots(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid UUID format", nil, reqID)
		return
	}

	slots, err := h.lockerService.GetLockerSlots(r.Context(), id)
	if err != nil {
		domain.WriteError(w, http.StatusInternalServerError, "FETCH_ERROR", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Daftar slot loker ditemukan", slots, reqID)
}
