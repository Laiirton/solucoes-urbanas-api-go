package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/laiirton/solucoes-urbanas-api/internal/config"
	"github.com/laiirton/solucoes-urbanas-api/internal/database"
	"github.com/laiirton/solucoes-urbanas-api/internal/repository"
	"github.com/laiirton/solucoes-urbanas-api/internal/routes"
	"github.com/laiirton/solucoes-urbanas-api/internal/services"
)

func main() {
	cfg := config.Load()

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}
	if cfg.SupabaseURL == "" || cfg.SupabaseKey == "" {
		log.Println("Warning: SUPABASE_URL or SUPABASE_KEY not provided. File uploads may fail.")
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	userRepo := repository.NewUserRepository(db.Pool)
	serviceRepo := repository.NewServiceRepository(db.Pool)
	srRepo := repository.NewServiceRequestRepository(db.Pool)
	newsRepo := repository.NewNewsRepository(db.Pool)
	teamRepo := repository.NewTeamRepository(db.Pool)
	pushTokenRepo := repository.NewPushTokenRepository(db.Pool)

	storageService := services.NewSupabaseStorageService(cfg.SupabaseURL, cfg.SupabaseKey, cfg.SupabaseBucket)

	router := routes.Setup(userRepo, serviceRepo, srRepo, newsRepo, teamRepo, pushTokenRepo, storageService, cfg.JWTSecret)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server starting on %s", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
