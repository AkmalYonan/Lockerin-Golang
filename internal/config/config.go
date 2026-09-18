package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv               string
	Port                 string
	SupabaseURL          string
	SupabaseAnonKey      string
	SupabaseServiceKey   string
	SupabaseDBURL        string
	JWTSecret            string
	MidtransServerKey    string
	MidtransClientKey    string
	MidtransIsProduction bool
	IoTDeviceSecretKey   string
	ReservationTTLMin    int
	PINMaxAttempts       int
	PINExpiryMin         int
}

func LoadConfig() *Config {
	// Attempt to load .env file if present
	if err := godotenv.Load(); err != nil {
		log.Println("Notice: No .env file found or unable to load, using system environment variables")
	}

	resTTL, _ := strconv.Atoi(getEnv("RESERVATION_TTL_MINUTES", "15"))
	pinAttempts, _ := strconv.Atoi(getEnv("PIN_MAX_ATTEMPTS", "5"))
	pinExpiry, _ := strconv.Atoi(getEnv("PIN_EXPIRY_MINUTES", "1440")) // 24 hours default
	midtransProd, _ := strconv.ParseBool(getEnv("MIDTRANS_IS_PRODUCTION", "false"))

	rawSupabaseURL := getEnv("SUPABASE_URL", "https://xyzcompany.supabase.co")
	rawSupabaseURL = strings.TrimRight(rawSupabaseURL, "/")
	rawSupabaseURL = strings.TrimSuffix(rawSupabaseURL, "/rest/v1")
	rawSupabaseURL = strings.TrimRight(rawSupabaseURL, "/")

	return &Config{
		AppEnv:               getEnv("APP_ENV", "development"),
		Port:                 getEnv("PORT", "8080"),
		SupabaseURL:          rawSupabaseURL,
		SupabaseAnonKey:      getEnv("SUPABASE_ANON_KEY", ""),
		SupabaseServiceKey:   getEnv("SUPABASE_SERVICE_ROLE_KEY", ""),
		SupabaseDBURL:        getEnv("SUPABASE_DB_URL", ""),
		JWTSecret:            getEnv("JWT_SECRET", "super-secret-lockerin-jwt-key-2026"),
		MidtransServerKey:    getEnv("MIDTRANS_SERVER_KEY", "SB-Mid-server-sample-key"),
		MidtransClientKey:    getEnv("MIDTRANS_CLIENT_KEY", "SB-Mid-client-sample-key"),
		MidtransIsProduction: midtransProd,
		IoTDeviceSecretKey:   getEnv("IOT_DEVICE_SECRET_KEY", "lockerin-head-locker-secret-2026"),
		ReservationTTLMin:    resTTL,
		PINMaxAttempts:       pinAttempts,
		PINExpiryMin:         pinExpiry,
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
