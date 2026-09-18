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

	"lockerin-backend/internal/app"
)

func main() {
	application, cleanup, err := app.NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize Lockerin application: %v", err)
	}
	defer cleanup()

	cfg := application.Config

	log.Println("==================================================")
	log.Println("  LOCKERIN MODULAR-MONOLITH BACKEND API (GOLANG)")
	log.Printf("  Environment : %s\n", cfg.AppEnv)
	log.Printf("  Port        : %s\n", cfg.Port)
	log.Println("==================================================")

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      application.Router,
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
