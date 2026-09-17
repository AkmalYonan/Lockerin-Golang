package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"lockerin-backend/models"
	"lockerin-backend/store"
)

func EnableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func JSONResponse(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

// Auth Handlers
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	EnableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "POST" {
		JSONResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	store.GlobalStore.RLock()
	defer store.GlobalStore.RUnlock()

	for _, u := range store.GlobalStore.Users {
		if (u.Username == req.Username || u.Email == req.Username) && u.Password == req.Password {
			token := fmt.Sprintf("token-lockerin-%s-%d", u.ID, time.Now().Unix())
			JSONResponse(w, http.StatusOK, map[string]interface{}{
				"message": "Login berhasil",
				"token":   token,
				"user":    u,
			})
			return
		}
	}

	JSONResponse(w, http.StatusUnauthorized, map[string]string{"error": "Username atau password salah"})
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	EnableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "POST" {
		JSONResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	var req models.User
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	store.GlobalStore.Lock()
	defer store.GlobalStore.Unlock()

	userID := fmt.Sprintf("user-%d", len(store.GlobalStore.Users)+1)
	req.ID = userID
	store.GlobalStore.Users[userID] = req

	JSONResponse(w, http.StatusCreated, map[string]interface{}{
		"message": "Registrasi berhasil",
		"user":    req,
	})
}

// Locker Stations Handler
func GetStationsHandler(w http.ResponseWriter, r *http.Request) {
	EnableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	store.GlobalStore.RLock()
	defer store.GlobalStore.RUnlock()

	var stations []models.LockerStation
	for _, st := range store.GlobalStore.Stations {
		stations = append(stations, st)
	}

	JSONResponse(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   stations,
	})
}

// Locker Grid Handler (A1 - E5)
func GetStationGridHandler(w http.ResponseWriter, r *http.Request) {
	EnableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	stationID := r.URL.Query().Get("station_id")
	if stationID == "" {
		stationID = "st-medistra" // default fallback
	}

	store.GlobalStore.RLock()
	defer store.GlobalStore.RUnlock()

	station, ok := store.GlobalStore.Stations[stationID]
	if !ok {
		JSONResponse(w, http.StatusNotFound, map[string]string{"error": "Station tidak ditemukan"})
		return
	}

	grids, ok := store.GlobalStore.Grids[stationID]
	if !ok {
		JSONResponse(w, http.StatusNotFound, map[string]string{"error": "Grid data tidak ditemukan"})
		return
	}

	resp := models.StationGridResponse{
		StationID:   station.ID,
		StationName: station.Name,
		Grids:       grids,
	}

	JSONResponse(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   resp,
	})
}

// Transactions Handler
func CreateTransactionHandler(w http.ResponseWriter, r *http.Request) {
	EnableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "POST" {
		JSONResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	var req struct {
		UserID        string `json:"user_id"`
		StationID     string `json:"station_id"`
		LockerCode    string `json:"locker_code"`
		DurationHours int    `json:"duration_hours"`
		PaymentMethod string `json:"payment_method"`
		SecurityPIN   string `json:"security_pin"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	store.GlobalStore.Lock()
	defer store.GlobalStore.Unlock()

	station, ok := store.GlobalStore.Stations[req.StationID]
	if !ok {
		JSONResponse(w, http.StatusNotFound, map[string]string{"error": "Stasiun loker tidak ditemukan"})
		return
	}

	// Update Grid status
	grids := store.GlobalStore.Grids[req.StationID]
	gridFound := false
	for i, g := range grids {
		if g.Code == req.LockerCode {
			if g.Status == "occupied" {
				JSONResponse(w, http.StatusConflict, map[string]string{"error": "Loker sedang terisi"})
				return
			}
			grids[i].Status = "occupied"
			grids[i].Occupant = req.UserID
			gridFound = true
			break
		}
	}
	if !gridFound {
		JSONResponse(w, http.StatusBadRequest, map[string]string{"error": "Kode loker tidak valid"})
		return
	}
	store.GlobalStore.Grids[req.StationID] = grids

	// Decrease station available count
	station.Available -= 1
	store.GlobalStore.Stations[req.StationID] = station

	trxID := fmt.Sprintf("trx-%d", time.Now().UnixNano()%100000)
	totalAmt := float64(req.DurationHours) * station.PricePerHour
	pin := req.SecurityPIN
	if pin == "" {
		pin = "123456"
	}

	trx := models.Transaction{
		ID:            trxID,
		UserID:        req.UserID,
		UserName:      "User Lockerin",
		StationID:     req.StationID,
		StationName:   station.Name,
		LockerCode:    req.LockerCode,
		SecurityPIN:   pin,
		CheckInTime:   time.Now(),
		CheckOutTime:  time.Now().Add(time.Duration(req.DurationHours) * time.Hour),
		DurationHours: req.DurationHours,
		TotalAmount:   totalAmt,
		PaymentMethod: req.PaymentMethod,
		PaymentStatus: "Paid",
		CreatedAt:     time.Now(),
	}

	store.GlobalStore.Transactions[trxID] = trx

	JSONResponse(w, http.StatusCreated, map[string]interface{}{
		"status":  "success",
		"message": "Transaksi berhasil dibuat dan dibayar",
		"data":    trx,
	})
}

func GetTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	EnableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	store.GlobalStore.RLock()
	defer store.GlobalStore.RUnlock()

	var trxs []models.Transaction
	for _, t := range store.GlobalStore.Transactions {
		trxs = append(trxs, t)
	}

	JSONResponse(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   trxs,
	})
}

// Verify PIN Handler (IoT Head Locker)
func VerifyPinHandler(w http.ResponseWriter, r *http.Request) {
	EnableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "POST" {
		JSONResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	var req models.VerifyPinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	store.GlobalStore.RLock()
	defer store.GlobalStore.RUnlock()

	for _, trx := range store.GlobalStore.Transactions {
		if (req.StationID == "" || trx.StationID == req.StationID) &&
			(req.LockerCode == "" || strings.EqualFold(trx.LockerCode, req.LockerCode)) {
			if trx.SecurityPIN == req.SecurityPIN {
				JSONResponse(w, http.StatusOK, models.VerifyPinResponse{
					Success: true,
					Message: fmt.Sprintf("PIN Benar! Solenoid Loker %s Berhasil Terbuka.", trx.LockerCode),
					Action:  "UNLOCKED",
				})
				return
			}
		}
	}

	// Also check default fallback PIN for quick test
	if req.SecurityPIN == "123456" || req.SecurityPIN == "000000" {
		JSONResponse(w, http.StatusOK, models.VerifyPinResponse{
			Success: true,
			Message: "PIN Valid! Loker Berhasil Terbuka.",
			Action:  "UNLOCKED",
		})
		return
	}

	JSONResponse(w, http.StatusUnauthorized, models.VerifyPinResponse{
		Success: false,
		Message: "PIN Keamanan Salah! Mohon periksa kembali kode Anda.",
		Action:  "LOCKED",
	})
}

// Notification Handlers
func GetNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	EnableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	store.GlobalStore.RLock()
	defer store.GlobalStore.RUnlock()

	JSONResponse(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   store.GlobalStore.Notifications,
	})
}

func CreateNotificationHandler(w http.ResponseWriter, r *http.Request) {
	EnableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "POST" {
		JSONResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	var req struct {
		Title    string `json:"title"`
		Message  string `json:"message"`
		Category string `json:"category"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	store.GlobalStore.Lock()
	defer store.GlobalStore.Unlock()

	notif := models.Notification{
		ID:        fmt.Sprintf("notif-%d", time.Now().Unix()),
		Title:     req.Title,
		Message:   req.Message,
		Category:  req.Category,
		CreatedAt: time.Now(),
	}

	store.GlobalStore.Notifications = append([]models.Notification{notif}, store.GlobalStore.Notifications...)

	JSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Notifikasi/Promo berhasil dikirim",
		"data":    notif,
	})
}

// Admin Stats Handler
func GetAdminStatsHandler(w http.ResponseWriter, r *http.Request) {
	EnableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	store.GlobalStore.RLock()
	defer store.GlobalStore.RUnlock()

	totalStations := len(store.GlobalStore.Stations)
	totalUsers := len(store.GlobalStore.Users)
	totalLockers := 0
	occupied := 0
	available := 0

	for _, grids := range store.GlobalStore.Grids {
		totalLockers += len(grids)
		for _, g := range grids {
			if g.Status == "occupied" {
				occupied++
			} else {
				available++
			}
		}
	}

	totalRev := 0.0
	activeRentals := 0
	for _, t := range store.GlobalStore.Transactions {
		totalRev += t.TotalAmount
		if t.PaymentStatus == "Paid" {
			activeRentals++
		}
	}

	stats := models.AdminStats{
		TotalStations:    totalStations,
		TotalLockers:     totalLockers,
		OccupiedLockers:  occupied,
		AvailableLockers: available,
		TotalRevenue:     totalRev,
		TotalUsers:       totalUsers,
		ActiveRentals:    activeRentals,
	}

	JSONResponse(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   stats,
	})
}
