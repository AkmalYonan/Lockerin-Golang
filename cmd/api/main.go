package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lockerin-backend/internal/auth"
	"lockerin-backend/internal/config"
	handlerhttp "lockerin-backend/internal/handler/http"
	"lockerin-backend/internal/integration/fcm"
	"lockerin-backend/internal/integration/locker_device"
	"lockerin-backend/internal/integration/midtrans"
	"lockerin-backend/internal/integration/supabase"
	"lockerin-backend/internal/repository"
	"lockerin-backend/internal/repository/inmemory"
	"lockerin-backend/internal/repository/postgres"
	"lockerin-backend/internal/service"
	"lockerin-backend/internal/worker"
)

func main() {
	cfg := config.LoadConfig()

	log.Println("==================================================")
	log.Println("  LOCKERIN MODULAR-MONOLITH BACKEND API (GOLANG)")
	log.Printf("  Environment : %s\n", cfg.AppEnv)
	log.Printf("  Port        : %s\n", cfg.Port)
	log.Println("==================================================")

	// 1. Initialize Repositories (PostgreSQL / Supabase or In-Memory fallback)
	var (
		profRepo   repository.ProfileRepository
		locRepo    repository.LocationRepository
		lockerRepo repository.LockerRepository
		slotRepo   repository.SlotRepository
		rentRepo   repository.RentalRepository
		secRepo    repository.SecurityCodeRepository
		payRepo    repository.PaymentRepository
		devRepo    repository.DeviceRepository
		notifRepo  repository.NotificationRepository
		promoRepo   repository.PromoRepository
		auditRepo   repository.AuditRepository
		pricingRepo repository.PricingRepository
	)

	if cfg.SupabaseDBURL != "" {
		pgDB, err := postgres.NewPostgresDB(cfg.SupabaseDBURL)
		if err != nil {
			log.Fatalf("Failed to connect to Supabase PostgreSQL: %v", err)
		}
		defer pgDB.Close()

		profRepo = postgres.NewProfileRepo(pgDB)
		locRepo = postgres.NewLocationRepo(pgDB)
		lockerRepo = postgres.NewLockerRepo(pgDB)
		slotRepo = postgres.NewSlotRepo(pgDB)
		rentRepo = postgres.NewRentalRepo(pgDB)
		secRepo = postgres.NewSecurityCodeRepo(pgDB)
		payRepo = postgres.NewPaymentRepo(pgDB)
		devRepo = postgres.NewDeviceRepo(pgDB)
		notifRepo = postgres.NewNotificationRepo(pgDB)
		promoRepo = postgres.NewPromoRepo(pgDB)
		auditRepo = postgres.NewAuditRepo(pgDB)
		pricingRepo = postgres.NewPricingRepo(pgDB)
		log.Println("[Database] Using Supabase PostgreSQL connection pool.")
	} else {
		log.Println("[Database] SUPABASE_DB_URL not set. Running with In-Memory Repository (Demo/Test Mode).")
		memStore := inmemory.NewInMemoryStore()
		repos := inmemory.NewRepositories(memStore)

		profRepo = repos.Profiles
		locRepo = repos.Locations
		lockerRepo = repos.Lockers
		slotRepo = repos.Slots
		rentRepo = repos.Rentals
		secRepo = repos.SecurityCodes
		payRepo = repos.Payments
		devRepo = repos.Devices
		notifRepo = repos.Notifications
		promoRepo = repos.Promos
		auditRepo = repos.Audit
		pricingRepo = repos.Pricing
	}

	// 2. Initialize Integrations
	supabaseClient := supabase.NewClient(cfg)
	midtransGateway := midtrans.NewGateway(cfg)
	fcmNotifier := fcm.NewNotifier(cfg)
	deviceDispatcher := locker_device.NewDispatcher(cfg)
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, 24*time.Hour)

	// 3. Initialize Services
	authService := service.NewAuthService(profRepo, jwtManager, supabaseClient)
	locationService := service.NewLocationService(locRepo, lockerRepo, slotRepo)
	lockerService := service.NewLockerService(lockerRepo, slotRepo)
	pinService := service.NewPINService(secRepo, cfg)
	rentalService := service.NewRentalService(rentRepo, slotRepo, lockerRepo, promoRepo, pinService, deviceDispatcher, cfg)
	paymentService := service.NewPaymentService(payRepo, rentRepo, profRepo, notifRepo, slotRepo, midtransGateway, pinService, fcmNotifier)
	deviceService := service.NewDeviceService(devRepo, lockerRepo, slotRepo)
	notifService := service.NewNotificationService(notifRepo)
	adminService := service.NewAdminService(locRepo, lockerRepo, slotRepo, profRepo, rentRepo, promoRepo, payRepo, devRepo, auditRepo, pricingRepo, deviceDispatcher)

	// 4. Start Background Worker for Expirations
	expiryWorker := worker.NewExpiryWorker(rentRepo, slotRepo, 1*time.Minute)
	expiryWorker.Start()
	defer expiryWorker.Stop()

	// 5. Setup Router & Handlers
	routerCfg := &handlerhttp.RouterConfig{
		AuthHandler:         handlerhttp.NewAuthHandler(authService),
		LocationHandler:     handlerhttp.NewLocationHandler(locationService),
		LockerHandler:       handlerhttp.NewLockerHandler(lockerService),
		RentalHandler:       handlerhttp.NewRentalHandler(rentalService, pinService),
		PaymentHandler:      handlerhttp.NewPaymentHandler(paymentService),
		DeviceHandler:       handlerhttp.NewDeviceHandler(deviceService),
		NotificationHandler: handlerhttp.NewNotificationHandler(notifService),
		AdminHandler:        handlerhttp.NewAdminHandler(adminService),
		SwaggerHandler:      handlerhttp.NewSwaggerHandler(),
		JWTManager:          jwtManager,
	}

	router := handlerhttp.SetupRouter(routerCfg)

	// 6. Start HTTP Server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server listening on http://localhost:%s\n", cfg.Port)
		log.Printf("Swagger UI documentation available at: http://localhost:%s/swagger\n", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited cleanly.")
}
