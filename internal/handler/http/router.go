package http

import (
	"net/http"
	"time"

	"lockerin-backend/internal/auth"
	"lockerin-backend/internal/domain"
	"lockerin-backend/internal/middleware"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type RouterConfig struct {
	AuthHandler         *AuthHandler
	LocationHandler     *LocationHandler
	LockerHandler       *LockerHandler
	RentalHandler       *RentalHandler
	PaymentHandler      *PaymentHandler
	DeviceHandler       *DeviceHandler
	NotificationHandler *NotificationHandler
	AdminHandler        *AdminHandler
	SwaggerHandler      *SwaggerHandler
	JWTManager          *auth.JWTManager
}

func SetupRouter(cfg *RouterConfig) *chi.Mux {
	r := chi.NewRouter()

	swaggerH := cfg.SwaggerHandler
	if swaggerH == nil {
		swaggerH = NewSwaggerHandler()
	}

	// Global Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Timeout(30 * time.Second))

	// CORS Setup for Flutter & Web Admin
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "X-Device-Secret"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Rate limiter for sensitive auth and PIN routes
	authLimiter := middleware.NewRateLimiter(20, 1*time.Minute)
	pinLimiter := middleware.NewRateLimiter(10, 1*time.Minute)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		reqID := middleware.GetRequestID(r.Context())
		domain.WriteSuccess(w, http.StatusOK, "Lockerin Golang API is running healthy", map[string]string{
			"status":  "healthy",
			"service": "Lockerin Modular Monolith Go API",
			"version": "2.0.0",
		}, reqID)
	})

	// Swagger API Documentation
	r.Get("/docs/openapi.yaml", swaggerH.ServeOpenAPISpec)
	r.Get("/swagger/openapi.yaml", swaggerH.ServeOpenAPISpec)
	r.Get("/swagger", swaggerH.ServeUI)
	r.Get("/swagger/*", swaggerH.ServeUI)
	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger", http.StatusMovedPermanently)
	})

	// API v1 Routes
	r.Route("/api/v1", func(r chi.Router) {

		// 1. Auth Routes
		r.Route("/auth", func(r chi.Router) {
			r.With(authLimiter.Limit()).Post("/register", cfg.AuthHandler.Register)
			r.With(authLimiter.Limit()).Post("/login", cfg.AuthHandler.Login)
			r.With(middleware.AuthMiddleware(cfg.JWTManager)).Post("/logout", cfg.AuthHandler.Logout)
			r.With(middleware.AuthMiddleware(cfg.JWTManager)).Post("/refresh", cfg.AuthHandler.Refresh)
		})

		r.With(middleware.AuthMiddleware(cfg.JWTManager)).Get("/me", cfg.AuthHandler.Me)

		// 2. Locations Routes
		r.Route("/locations", func(r chi.Router) {
			r.Get("/", cfg.LocationHandler.ListLocations)
			r.Get("/{id}", cfg.LocationHandler.GetLocation)
			r.Get("/{id}/availability", cfg.LocationHandler.GetAvailability)
		})

		// 3. Lockers & Slots Routes
		r.Route("/lockers", func(r chi.Router) {
			r.Get("/{id}", cfg.LockerHandler.GetLocker)
			r.Get("/{id}/slots", cfg.LockerHandler.GetSlots)
		})

		// 4. Rental Routes
		r.Route("/rentals", func(r chi.Router) {
			r.Post("/quote", cfg.RentalHandler.Quote)

			// Security code verification (can be called by physical locker or user)
			r.With(pinLimiter.Limit()).Post("/{id}/security-code/verify", cfg.RentalHandler.VerifySecurityCode)

			// Authenticated Rental Actions
			r.Group(func(r chi.Router) {
				r.Use(middleware.AuthMiddleware(cfg.JWTManager))
				r.Get("/", cfg.RentalHandler.List)
				r.Get("/{id}", cfg.RentalHandler.GetByID)
				r.Post("/reserve", cfg.RentalHandler.Reserve)
				r.Post("/{id}/start", cfg.RentalHandler.Start)
				r.Post("/{id}/open", cfg.RentalHandler.Open)
				r.Post("/{id}/close", cfg.RentalHandler.Close)
			})
		})

		// 5. Payment Routes
		r.Route("/payments", func(r chi.Router) {
			// Webhook from Midtrans (Provider callback)
			r.Post("/webhook/midtrans", cfg.PaymentHandler.MidtransWebhook)

			r.Group(func(r chi.Router) {
				r.Use(middleware.AuthMiddleware(cfg.JWTManager))
				r.Post("/", cfg.PaymentHandler.CreatePayment)
				r.Get("/{id}", cfg.PaymentHandler.GetPaymentByID)
			})
		})

		// 6. Notifications & FCM Devices
		r.Route("/devices", func(r chi.Router) {
			r.With(middleware.AuthMiddleware(cfg.JWTManager)).Post("/fcm-token", cfg.NotificationHandler.RegisterFCMToken)
		})

		r.Route("/notifications", func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(cfg.JWTManager))
			r.Get("/", cfg.NotificationHandler.GetNotifications)
			r.Patch("/{id}/read", cfg.NotificationHandler.MarkRead)
		})

		// 7. IoT Head Locker Gateway
		r.Route("/device", func(r chi.Router) {
			r.Post("/heartbeat", cfg.DeviceHandler.Heartbeat)
			r.Post("/status", cfg.DeviceHandler.Status)
			r.Post("/commands/{id}/ack", cfg.DeviceHandler.CommandAck)
		})

		// 8. Admin Routes (Protected by RBAC: Admin / Operator)
		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(cfg.JWTManager))
			r.Use(middleware.RequireRole(domain.RoleAdmin, domain.RoleOperator))

			r.Get("/dashboard", cfg.AdminHandler.Dashboard)
			r.Get("/lockers", cfg.AdminHandler.ListLockers)
			r.Get("/users", cfg.AdminHandler.ListUsers)
			r.Get("/users/{id}", cfg.AdminHandler.GetUserDetail)
			r.Patch("/users/{id}/status", cfg.AdminHandler.UpdateUserStatus)

			// Pricing
			r.Get("/pricing", cfg.AdminHandler.GetPricing)
			r.Put("/pricing", cfg.AdminHandler.UpdatePricing)

			// Locations CRUD
			r.Post("/locations", cfg.AdminHandler.CreateLocation)
			r.Put("/locations/{id}", cfg.AdminHandler.UpdateLocation)
			r.Patch("/locations/{id}/status", cfg.AdminHandler.UpdateLocationStatus)
			r.Delete("/locations/{id}", cfg.AdminHandler.DeleteLocation)

			// Promos CRUD
			r.Post("/promos", cfg.AdminHandler.CreatePromo)
			r.Get("/promos", cfg.AdminHandler.ListPromos)
			r.Patch("/promos/{id}/status", cfg.AdminHandler.UpdatePromoStatus)

			// Device Remote Control / Emergency Commands
			r.Post("/device/commands", cfg.AdminHandler.DispatchDeviceCommand)

			// Audits & Transactions
			r.Get("/transactions", cfg.AdminHandler.ListTransactions)
			r.Get("/transactions/{id}", cfg.AdminHandler.GetTransactionDetail)
			r.Get("/audit-logs", cfg.AdminHandler.ListAuditLogs)
		})
	})

	return r
}
