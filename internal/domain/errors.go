package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrNotFound            = errors.New("resource not found")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrBadRequest          = errors.New("bad request")
	ErrSlotUnavailable     = errors.New("slot is already reserved or occupied")
	ErrInvalidPIN          = errors.New("invalid security PIN")
	ErrPINLocked           = errors.New("PIN entry locked due to too many attempts")
	ErrPINExpired          = errors.New("PIN has expired")
	ErrInvalidRentalState  = errors.New("invalid rental state transition")
	ErrPaymentAlreadyPaid  = errors.New("payment is already settled")
	ErrInvalidSignature    = errors.New("invalid signature verification")
	ErrDeviceOffline       = errors.New("locker device is offline")
	ErrConflict            = errors.New("resource conflict")
	ErrInternalServerError = errors.New("internal server error")
)

type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type AppError struct {
	StatusCode int
	Code       string
	Message    string
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func NewAppError(status int, code, msg string, err error) *AppError {
	return &AppError{
		StatusCode: status,
		Code:       code,
		Message:    msg,
		Err:        err,
	}
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func WriteSuccess(w http.ResponseWriter, status int, message string, data interface{}, requestID string) {
	WriteJSON(w, status, APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		RequestID: requestID,
	})
}

func WriteError(w http.ResponseWriter, status int, code, message string, details interface{}, requestID string) {
	WriteJSON(w, status, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
		RequestID: requestID,
	})
}
