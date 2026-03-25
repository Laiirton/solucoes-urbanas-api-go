package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/laiirton/solucoes-urbanas-api/internal/handlers"
	"github.com/laiirton/solucoes-urbanas-api/internal/middleware"
	"github.com/laiirton/solucoes-urbanas-api/internal/repository"
	"github.com/laiirton/solucoes-urbanas-api/internal/services"
)

func Setup(
	userRepo *repository.UserRepository,
	serviceRepo *repository.ServiceRepository,
	srRepo *repository.ServiceRequestRepository,
	storageService services.StorageService,
	jwtSecret string,
) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.RequestID)

	authHandler := handlers.NewAuthHandler(userRepo, jwtSecret)
	userHandler := handlers.NewUserHandler(userRepo)
	serviceHandler := handlers.NewServiceHandler(serviceRepo)
	srHandler := handlers.NewServiceRequestHandler(srRepo, storageService)
	geoHandler := handlers.NewGeolocationHandler()
	homeHandler := handlers.NewHomeHandler(srRepo, userRepo)
	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "ok",
			"timestamp": time.Now().UTC(),
		})
	})

	// Routes under /api
	r.Route("/api", func(r chi.Router) {
		// Public auth routes
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)

		// Geolocation route
		r.Get("/geolocation", geoHandler.Search)

		// Public service routes (read-only)
		r.Get("/services", serviceHandler.ListServices)
		r.Get("/services/{id}", serviceHandler.GetService)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(jwtSecret))

			// Auth
			r.Post("/auth/logout", authHandler.Logout)

			// Home
			r.Get("/home", homeHandler.Index)

			// Users
			r.Get("/users", userHandler.ListUsers)
			r.Get("/users/me", userHandler.GetMe)
			r.Get("/users/{id}", userHandler.GetUser)
			r.Put("/users/{id}", userHandler.UpdateUser)
			r.Delete("/users/{id}", userHandler.DeleteUser)

			// Services (write)
			r.Post("/services", serviceHandler.CreateService)
			r.Put("/services/{id}", serviceHandler.UpdateService)
			r.Delete("/services/{id}", serviceHandler.DeleteService)

			// Service Requests
			r.Post("/service-requests", srHandler.CreateServiceRequest)
			r.Get("/service-requests", srHandler.ListServiceRequests)
			r.Get("/service-requests/{id}", srHandler.GetServiceRequest)
			r.Patch("/service-requests/{id}/status", srHandler.UpdateServiceRequestStatus)
			r.Delete("/service-requests/{id}", srHandler.DeleteServiceRequest)
		})
	})

	return r
}
