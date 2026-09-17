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

type NotificationHandler struct {
	notifService *service.NotificationService
}

func NewNotificationHandler(notifService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifService: notifService}
}

func (h *NotificationHandler) RegisterFCMToken(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		domain.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil, reqID)
		return
	}

	var req domain.RegisterFCMTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil, reqID)
		return
	}

	if err := h.notifService.RegisterFCMToken(r.Context(), claims.UserID, &req); err != nil {
		domain.WriteError(w, http.StatusInternalServerError, "REGISTRATION_FAILED", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Token perangkat FCM berhasil didaftarkan", nil, reqID)
}

func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		domain.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil, reqID)
		return
	}

	notifs, err := h.notifService.GetNotifications(r.Context(), claims.UserID)
	if err != nil {
		domain.WriteError(w, http.StatusInternalServerError, "FETCH_ERROR", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Daftar notifikasi ditemukan", notifs, reqID)
}

func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	claims := middleware.GetUserClaims(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid UUID", nil, reqID)
		return
	}

	if err := h.notifService.MarkRead(r.Context(), id, claims.UserID); err != nil {
		domain.WriteError(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Notifikasi telah ditandai dibaca", nil, reqID)
}
