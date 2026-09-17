package store

import (
	"fmt"
	"sync"
	"time"

	"lockerin-backend/models"
)

type Store struct {
	sync.RWMutex
	Users         map[string]models.User
	Stations      map[string]models.LockerStation
	Grids         map[string][]models.LockerGrid // station_id -> grids
	Transactions  map[string]models.Transaction
	Notifications []models.Notification
}

var GlobalStore *Store

func InitStore() {
	s := &Store{
		Users:         make(map[string]models.User),
		Stations:      make(map[string]models.LockerStation),
		Grids:         make(map[string][]models.LockerGrid),
		Transactions:  make(map[string]models.Transaction),
		Notifications: []models.Notification{},
	}

	// Seed User
	s.Users["user1"] = models.User{
		ID:       "user1",
		Username: "akmal",
		Email:    "akmal@lockerin.id",
		Password: "password123",
		Name:     "Akmal Maindata",
		Phone:    "081234567890",
	}

	// Seed Stations (Peta & Penentuan Lokasi Loker)
	s.Stations["st-medistra"] = models.LockerStation{
		ID:           "st-medistra",
		Name:         "Lockerin RS Medistra",
		Address:      "Jl. Jend. Gatot Subroto No.Kav. 59, Jakarta Selatan",
		Latitude:     -6.2372,
		Longitude:    106.8335,
		TotalGrid:    25,
		Available:    18,
		PricePerHour: 7000,
		ImageUrl:     "https://images.unsplash.com/photo-1590381105924-c72589b9ef3f?w=600",
	}

	s.Stations["st-paramadina"] = models.LockerStation{
		ID:           "st-paramadina",
		Name:         "Lockerin Univ. Paramadina",
		Address:      "Jl. Gatot Subroto No.Kav. 97, Mampang Prapatan, Jakarta Selatan",
		Latitude:     -6.2415,
		Longitude:    106.8329,
		TotalGrid:    25,
		Available:    21,
		PricePerHour: 5000,
		ImageUrl:     "https://images.unsplash.com/photo-1541829070764-84a7d30dd3f3?w=600",
	}

	// Seed Grid Lockers (A1 - E5)
	rows := []string{"A", "B", "C", "D", "E"}
	for _, stationID := range []string{"st-medistra", "st-paramadina"} {
		var grids []models.LockerGrid
		for _, r := range rows {
			for c := 1; c <= 5; c++ {
				code := fmt.Sprintf("%s%d", r, c)
				status := "available"
				size := "Medium"
				if c == 1 {
					size = "Small"
				} else if c == 5 {
					size = "Large"
				}
				// Mock some occupied lockers
				if (r == "A" && c == 2) || (r == "C" && c == 4) || (r == "E" && c == 1) {
					status = "occupied"
				}

				grids = append(grids, models.LockerGrid{
					Code:     code,
					Status:   status,
					Size:     size,
					Occupant: "",
				})
			}
		}
		s.Grids[stationID] = grids
	}

	// Seed Sample Transaction
	s.Transactions["trx-001"] = models.Transaction{
		ID:            "trx-001",
		UserID:        "user1",
		UserName:      "Akmal Maindata",
		StationID:     "st-medistra",
		StationName:   "Lockerin RS Medistra",
		LockerCode:    "B3",
		SecurityPIN:   "123456",
		CheckInTime:   time.Now().Add(-2 * time.Hour),
		CheckOutTime:  time.Now().Add(2 * time.Hour),
		DurationHours: 4,
		TotalAmount:   28000,
		PaymentMethod: "DANA",
		PaymentStatus: "Paid",
		CreatedAt:     time.Now().Add(-2 * time.Hour),
	}

	// Seed Notifications
	s.Notifications = []models.Notification{
		{
			ID:        "notif-1",
			Title:     "Promo Menarik!",
			Message:   "Gunakan voucher LOCKERINNEW untuk diskon 50% pengguna baru!",
			Category:  "Promo",
			CreatedAt: time.Now().Add(-24 * time.Hour),
		},
		{
			ID:        "notif-2",
			Title:     "Peringatan Durasi Sewa",
			Message:   "Sewa loker B3 di RS Medistra tersisa 30 menit lagi.",
			Category:  "Alert",
			CreatedAt: time.Now().Add(-10 * time.Minute),
		},
	}

	GlobalStore = s
}
