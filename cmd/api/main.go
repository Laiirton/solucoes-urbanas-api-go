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

	if err := database.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db.Pool)
	serviceRepo := repository.NewServiceRepository(db.Pool)
	srRepo := repository.NewServiceRequestRepository(db.Pool)
	newsRepo := repository.NewNewsRepository(db.Pool)
	teamRepo := repository.NewTeamRepository(db.Pool)
	regionRepo := repository.NewRegionRepository(db.Pool)
	pushTokenRepo := repository.NewPushTokenRepository(db.Pool)
	sysNotifRepo := repository.NewSystemNotificationRepository(db.Pool)
	appConfigRepo := repository.NewAppConfigRepository(db.Pool)
	serviceRatingRepo := repository.NewServiceRatingRepository(db.Pool)
	attendanceRepo := repository.NewServiceAttendanceRepository(db.Pool)
	categoryRepo := repository.NewCategoryRepository(db.Pool)

	storageService := services.NewSupabaseStorageService(cfg.SupabaseURL, cfg.SupabaseKey, cfg.SupabaseBucket)

	router := routes.Setup(userRepo, serviceRepo, srRepo, newsRepo, teamRepo, regionRepo, pushTokenRepo, sysNotifRepo, appConfigRepo, serviceRatingRepo, attendanceRepo, categoryRepo, storageService, cfg.JWTSecret)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server starting on %s", addr)

	// Graceful shutdown
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Println("Server stopped.")
}
