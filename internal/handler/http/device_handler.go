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

type DeviceHandler struct {
	deviceService *service.DeviceService
}

func NewDeviceHandler(deviceService *service.DeviceService) *DeviceHandler {
	return &DeviceHandler{deviceService: deviceService}
}

func (h *DeviceHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	var req domain.HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON", nil, reqID)
		return
	}

	if err := h.deviceService.Heartbeat(r.Context(), &req); err != nil {
		domain.WriteError(w, http.StatusBadRequest, "HEARTBEAT_FAILED", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Heartbeat acknowledged", nil, reqID)
}

func (h *DeviceHandler) Status(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	var req domain.DeviceStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON", nil, reqID)
		return
	}

	if err := h.deviceService.SyncStatus(r.Context(), &req); err != nil {
		domain.WriteError(w, http.StatusBadRequest, "STATUS_SYNC_FAILED", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Status synced successfully", nil, reqID)
}

func (h *DeviceHandler) CommandAck(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	idStr := chi.URLParam(r, "id")
	cmdID, _ := uuid.Parse(idStr)

	var req domain.CommandAckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON", nil, reqID)
		return
	}

	if err := h.deviceService.AcknowledgeCommand(r.Context(), cmdID, &req); err != nil {
		domain.WriteError(w, http.StatusBadRequest, "ACK_FAILED", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Command ACK processed", nil, reqID)
}
