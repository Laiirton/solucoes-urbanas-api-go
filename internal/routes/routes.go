package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/laiirton/solucoes-urbanas-api/internal/handlers"
	"github.com/laiirton/solucoes-urbanas-api/internal/middleware"
	"github.com/laiirton/solucoes-urbanas-api/internal/repository"
	"github.com/laiirton/solucoes-urbanas-api/internal/services"
)

func Setup(
	userRepo *repository.UserRepository,
	serviceRepo *repository.ServiceRepository,
	srRepo *repository.ServiceRequestRepository,
	newsRepo *repository.NewsRepository,
	storageService services.StorageService,
	jwtSecret string,
) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"*"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	authHandler := handlers.NewAuthHandler(userRepo, jwtSecret)
	userHandler := handlers.NewUserHandler(userRepo, srRepo)
	serviceHandler := handlers.NewServiceHandler(serviceRepo)
	srHandler := handlers.NewServiceRequestHandler(srRepo, userRepo, storageService)
	geoHandler := handlers.NewGeolocationHandler()
	homeHandler := handlers.NewHomeHandler(srRepo, userRepo)
	newsHandler := handlers.NewNewsHandler(newsRepo, storageService)
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

		// Public news routes (read-only)
		r.Get("/news", newsHandler.ListNews)
		r.Get("/news/{id}", newsHandler.GetNews)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(jwtSecret))

			// Auth
			r.Get("/auth/me", userHandler.GetMe)
			r.Post("/auth/logout", authHandler.Logout)

			// Home
			r.Get("/home", homeHandler.Index)

			// Users
			r.Get("/users", userHandler.ListUsers)
			r.Post("/users", userHandler.CreateUser)
			r.Get("/users/me", userHandler.GetMe)
			r.Get("/users/{id}", userHandler.GetUser)
			r.Put("/users/{id}", userHandler.UpdateUser)
			r.Delete("/users/{id}", userHandler.DeleteUser)

			// Services (write)
			r.Post("/services", serviceHandler.CreateService)
			r.Put("/services/{id}", serviceHandler.UpdateService)
			r.Delete("/services/{id}", serviceHandler.DeleteService)

			// News (write)
			r.Post("/news", newsHandler.CreateNews)
			r.Put("/news/{id}", newsHandler.UpdateNews)
			r.Delete("/news/{id}", newsHandler.DeleteNews)

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
