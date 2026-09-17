package http

import (
	"encoding/json"
	"net/http"

	"lockerin-backend/internal/domain"
	"lockerin-backend/internal/middleware"
	"lockerin-backend/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	var req domain.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Failed to parse JSON body", nil, reqID)
		return
	}

	resp, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			domain.WriteError(w, appErr.StatusCode, appErr.Code, appErr.Message, nil, reqID)
			return
		}
		domain.WriteError(w, http.StatusInternalServerError, "REGISTRATION_FAILED", err.Error(), nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusCreated, "Registrasi berhasil", resp, reqID)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		domain.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Failed to parse JSON body", nil, reqID)
		return
	}

	resp, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			domain.WriteError(w, appErr.StatusCode, appErr.Code, appErr.Message, nil, reqID)
			return
		}
		domain.WriteError(w, http.StatusUnauthorized, "LOGIN_FAILED", "Authentication failed", nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Login berhasil", resp, reqID)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		domain.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil, reqID)
		return
	}

	profile, err := h.authService.GetProfile(r.Context(), claims.UserID)
	if err != nil {
		domain.WriteError(w, http.StatusNotFound, "USER_NOT_FOUND", "Profile not found", nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Profil pengguna ditemukan", profile, reqID)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	// In stateless JWT, logout is client-side clearance or server-side blacklist
	domain.WriteSuccess(w, http.StatusOK, "Logout berhasil", nil, reqID)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		domain.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Valid session required", nil, reqID)
		return
	}

	profile, err := h.authService.GetProfile(r.Context(), claims.UserID)
	if err != nil {
		domain.WriteError(w, http.StatusNotFound, "USER_NOT_FOUND", "Profile not found", nil, reqID)
		return
	}

	domain.WriteSuccess(w, http.StatusOK, "Session refreshed", profile, reqID)
}
