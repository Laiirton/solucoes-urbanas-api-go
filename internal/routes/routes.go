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
	teamRepo *repository.TeamRepository,
	pushTokenRepo *repository.PushTokenRepository,
	sysNotifRepo *repository.SystemNotificationRepository,
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
	userHandler := handlers.NewUserHandler(userRepo, srRepo, storageService)
	serviceHandler := handlers.NewServiceHandler(serviceRepo, srRepo)
	uploadService := services.NewUploadService(storageService)
	geoService := services.NewGeocodingService()
	pushService := services.NewExpoPushService()
	srHandler := handlers.NewServiceRequestHandler(srRepo, userRepo, sysNotifRepo, pushTokenRepo, pushService, uploadService, geoService)
	geoHandler := handlers.NewGeolocationHandler()
	homeHandler := handlers.NewHomeHandler(srRepo, userRepo, geoService)
	newsHandler := handlers.NewNewsHandler(newsRepo, pushTokenRepo, sysNotifRepo, pushService, storageService)
	notificationHandler := handlers.NewNotificationHandler(pushTokenRepo, sysNotifRepo)
	teamHandler := handlers.NewTeamHandler(teamRepo)
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
			r.Post("/users/{id}/profile-image", userHandler.UploadProfileImage)
			r.Delete("/users/{id}/profile-image", userHandler.DeleteProfileImage)

			// Services (write)
			r.Post("/services", serviceHandler.CreateService)
			r.Put("/services/{id}", serviceHandler.UpdateService)
			r.Delete("/services/{id}", serviceHandler.DeleteService)

			// News (write)
			r.Post("/news", newsHandler.CreateNews)
			r.Post("/news/upload-image", newsHandler.UploadImage)
			r.Put("/news/{id}", newsHandler.UpdateNews)
			r.Delete("/news/{id}", newsHandler.DeleteNews)

			// Notifications
			r.Post("/notifications/push-tokens", notificationHandler.RegisterPushToken)
			r.Get("/notifications", notificationHandler.ListSystemNotifications)
			r.Post("/notifications", notificationHandler.CreateSystemNotification)
			r.Get("/notifications/{id}", notificationHandler.GetSystemNotification)
			r.Put("/notifications/{id}", notificationHandler.UpdateSystemNotification)
			r.Patch("/notifications/{id}/read", notificationHandler.MarkSystemNotificationAsRead)
			r.Delete("/notifications/{id}", notificationHandler.DeleteSystemNotification)

			// Teams
			r.Get("/teams", teamHandler.ListTeams)
			r.Post("/teams", teamHandler.CreateTeam)
			r.Get("/teams/{id}", teamHandler.GetTeam)
			r.Put("/teams/{id}", teamHandler.UpdateTeam)
			r.Delete("/teams/{id}", teamHandler.DeleteTeam)

			// Service Requests
			r.Post("/service-requests", srHandler.CreateServiceRequest)
			r.Get("/service-requests", srHandler.ListServiceRequests)
			r.Route("/service-requests/{id}", func(r chi.Router) {
				r.Get("/", srHandler.GetServiceRequest)
				r.Patch("/status", srHandler.UpdateServiceRequestStatus)
				r.Delete("/", srHandler.DeleteServiceRequest)
			})

			// Geocoding
			r.Get("/geocode-service-requests", srHandler.GeocodeAllServiceRequests)
			r.Get("/geocode-service-requests/{id}", srHandler.GeocodeServiceRequest)
		})
	})

	return r
}
