package app

import (
	"log"
	"net/http"
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

type App struct {
	Config *config.Config
	Router http.Handler
}

// NewApp initializes all database connections, services, integrations, and HTTP handlers.
func NewApp() (*App, func(), error) {
	cfg := config.LoadConfig()

	var (
		profRepo    repository.ProfileRepository
		locRepo     repository.LocationRepository
		lockerRepo  repository.LockerRepository
		slotRepo    repository.SlotRepository
		rentRepo    repository.RentalRepository
		secRepo     repository.SecurityCodeRepository
		payRepo     repository.PaymentRepository
		devRepo     repository.DeviceRepository
		notifRepo   repository.NotificationRepository
		promoRepo   repository.PromoRepository
		auditRepo   repository.AuditRepository
		pricingRepo repository.PricingRepository
		cleanupFunc = func() {}
	)

	if cfg.SupabaseDBURL != "" {
		pgDB, err := postgres.NewPostgresDB(cfg.SupabaseDBURL)
		if err != nil {
			log.Printf("[Database ERROR] Failed to connect to Supabase PostgreSQL: %v", err)
			return nil, cleanupFunc, err
		}
		cleanupFunc = func() {
			pgDB.Close()
		}

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

	// Integrations
	supabaseClient := supabase.NewClient(cfg)
	midtransGateway := midtrans.NewGateway(cfg)
	fcmNotifier := fcm.NewNotifier(cfg)
	deviceDispatcher := locker_device.NewDispatcher(cfg)
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, 24*time.Hour)

	// Services
	authService := service.NewAuthService(profRepo, jwtManager, supabaseClient)
	locationService := service.NewLocationService(locRepo, lockerRepo, slotRepo)
	lockerService := service.NewLockerService(lockerRepo, slotRepo)
	pinService := service.NewPINService(secRepo, cfg)
	rentalService := service.NewRentalService(rentRepo, slotRepo, lockerRepo, promoRepo, pinService, deviceDispatcher, cfg)
	paymentService := service.NewPaymentService(payRepo, rentRepo, profRepo, notifRepo, slotRepo, midtransGateway, pinService, fcmNotifier)
	deviceService := service.NewDeviceService(devRepo, lockerRepo, slotRepo)
	notifService := service.NewNotificationService(notifRepo)
	adminService := service.NewAdminService(locRepo, lockerRepo, slotRepo, profRepo, rentRepo, promoRepo, payRepo, devRepo, auditRepo, pricingRepo, deviceDispatcher)

	// Background Worker for Expirations
	expiryWorker := worker.NewExpiryWorker(rentRepo, slotRepo, 1*time.Minute)
	expiryWorker.Start()
	oldCleanup := cleanupFunc
	cleanupFunc = func() {
		expiryWorker.Stop()
		oldCleanup()
	}

	// Setup Router & Handlers
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

	return &App{
		Config: cfg,
		Router: router,
	}, cleanupFunc, nil
}
