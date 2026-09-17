package models

import "time"

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
}

type LockerStation struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	TotalGrid int     `json:"total_grid"`
	Available int     `json:"available"`
	PricePerHour float64 `json:"price_per_hour"`
	ImageUrl  string  `json:"image_url"`
}

type LockerGrid struct {
	Code     string `json:"code"`     // e.g. A1, A2, ..., E5
	Status   string `json:"status"`   // "available", "occupied", "selected"
	Size     string `json:"size"`     // "Small", "Medium", "Large"
	Occupant string `json:"occupant"` // Username/UserID if occupied
}

type StationGridResponse struct {
	StationID   string       `json:"station_id"`
	StationName string       `json:"station_name"`
	Grids       []LockerGrid `json:"grids"`
}

type Transaction struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	UserName      string    `json:"user_name"`
	StationID     string    `json:"station_id"`
	StationName   string    `json:"station_name"`
	LockerCode    string    `json:"locker_code"`
	SecurityPIN   string    `json:"security_pin"`
	CheckInTime   time.Time `json:"check_in_time"`
	CheckOutTime  time.Time `json:"check_out_time"`
	DurationHours int       `json:"duration_hours"`
	TotalAmount   float64   `json:"total_amount"`
	PaymentMethod string    `json:"payment_method"` // e.g. "DANA", "ShopeePay", "Virtual Account BCA"
	PaymentStatus string    `json:"payment_status"` // "Pending", "Paid", "Completed", "Cancelled"
	CreatedAt     time.Time `json:"created_at"`
}

type VerifyPinRequest struct {
	StationID   string `json:"station_id"`
	LockerCode  string `json:"locker_code"`
	SecurityPIN string `json:"security_pin"`
}

type VerifyPinResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Action  string `json:"action"` // "UNLOCKED" or "LOCKED"
}

type Notification struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Category  string    `json:"category"` // "Promo", "System", "Alert"
	CreatedAt time.Time `json:"created_at"`
}

type AdminStats struct {
	TotalStations    int     `json:"total_stations"`
	TotalLockers     int     `json:"total_lockers"`
	OccupiedLockers  int     `json:"occupied_lockers"`
	AvailableLockers int     `json:"available_lockers"`
	TotalRevenue     float64 `json:"total_revenue"`
	TotalUsers       int     `json:"total_users"`
	ActiveRentals    int     `json:"active_rentals"`
}
