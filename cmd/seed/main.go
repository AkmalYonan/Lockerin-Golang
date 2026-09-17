package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"lockerin-backend/internal/config"
	"lockerin-backend/internal/repository/postgres"
)

func main() {
	cfg := config.LoadConfig()

	log.Println("==================================================")
	log.Println("   LOCKERIN SUPABASE MIGRATION & SEEDER RUNNER    ")
	log.Println("==================================================")

	if cfg.SupabaseDBURL == "" {
		log.Fatal("ERROR: SUPABASE_DB_URL is not set in .env")
	}

	log.Println("Connecting to Supabase PostgreSQL Database...")
	pgDB, err := postgres.NewPostgresDB(cfg.SupabaseDBURL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer pgDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Ensure check constraint allows lowercase roles
	log.Println("Aligning constraints...")
	_, _ = pgDB.Pool.Exec(ctx, `
		ALTER TABLE profiles DROP CONSTRAINT IF EXISTS profiles_role_check;
		ALTER TABLE profiles ADD CONSTRAINT profiles_role_check CHECK (role IN ('user', 'admin', 'operator', 'viewer', 'USER', 'ADMIN', 'OPERATOR', 'VIEWER'));
	`)

	// Migration files in order
	migrationFiles := []string{
		"001_profiles.sql",
		"002_locations.sql",
		"003_lockers.sql",
		"004_rentals.sql",
		"005_security_codes.sql",
		"006_payments.sql",
		"007_devices.sql",
		"008_notifications_promos_audit.sql",
		"009_seed_data.sql",
		"010_pricing_tiers.sql",
	}

	for _, file := range migrationFiles {
		filePath := filepath.Join("migrations", file)
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", file, err)
		}

		log.Printf("Executing %s ...", file)
		_, err = pgDB.Pool.Exec(ctx, string(content))
		if err != nil {
			log.Fatalf("Migration failed on %s: %v", file, err)
		}
		log.Printf("✔ %s executed successfully!", file)
	}

	log.Println("--------------------------------------------------")
	log.Println("Verifying seeded data in Supabase...")

	var totalUsers, totalLocations, totalLockers, totalSlots, totalPromos, totalRentals int
	_ = pgDB.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM profiles").Scan(&totalUsers)
	_ = pgDB.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM locations").Scan(&totalLocations)
	_ = pgDB.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM lockers").Scan(&totalLockers)
	_ = pgDB.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM locker_slots").Scan(&totalSlots)
	_ = pgDB.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM promos").Scan(&totalPromos)
	_ = pgDB.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM rentals").Scan(&totalRentals)

	fmt.Printf("\n🎉 Supabase Database Seeded Successfully!\n")
	fmt.Printf("   - Profiles/Users : %d\n", totalUsers)
	fmt.Printf("   - Locations      : %d\n", totalLocations)
	fmt.Printf("   - Lockers        : %d\n", totalLockers)
	fmt.Printf("   - Locker Slots   : %d\n", totalSlots)
	fmt.Printf("   - Promos         : %d\n", totalPromos)
	fmt.Printf("   - Rentals        : %d\n", totalRentals)
	fmt.Println("==================================================")
}
