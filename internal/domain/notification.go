package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID        uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"user_id"`
	Type      string          `json:"type"` // general, promo, payment, rental, alert
	Title     string          `json:"title"`
	Body      string          `json:"body"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
	ReadAt    *time.Time      `json:"read_at,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

type FCMDevice struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	Token      string    `json:"token"`
	Platform   string    `json:"platform"` // android, ios, web
	LastSeenAt time.Time `json:"last_seen_at"`
	CreatedAt  time.Time `json:"created_at"`
}

type RegisterFCMTokenRequest struct {
	Token    string `json:"token"`
	Platform string `json:"platform"`
}

type Promo struct {
	ID                uuid.UUID  `json:"id"`
	Code              string     `json:"code"`
	Title             string     `json:"title"`
	Description       string     `json:"description"`
	DiscountType      string     `json:"discount_type"` // percentage, fixed
	DiscountValue     float64    `json:"discount_value"`
	MinOrderAmount    float64    `json:"min_order_amount"`
	MaxDiscountAmount *float64   `json:"max_discount_amount,omitempty"`
	StartsAt          time.Time  `json:"starts_at"`
	EndsAt            time.Time  `json:"ends_at"`
	Status            string     `json:"status"` // active, inactive, expired
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type CreatePromoRequest struct {
	Code              string    `json:"code"`
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	DiscountType      string    `json:"discount_type"` // percentage, fixed
	DiscountValue     float64   `json:"discount_value"`
	DiscountPercent   float64   `json:"discount_percent,omitempty"`
	MaxDiscountAmount *float64  `json:"max_discount_amount,omitempty"`
	MaxDiscount       *float64  `json:"max_discount,omitempty"`
	MinOrderAmount    float64   `json:"min_order_amount,omitempty"`
	UsageLimit        int       `json:"usage_limit,omitempty"`
	Quota             int       `json:"quota,omitempty"`
	StartsAt          time.Time `json:"starts_at"`
	EndsAt            time.Time `json:"ends_at"`
	ValidUntil        time.Time `json:"valid_until,omitempty"`
	Status            string    `json:"status,omitempty"`
}

