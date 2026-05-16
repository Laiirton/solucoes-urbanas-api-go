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
	regionRepo *repository.RegionRepository,
	pushTokenRepo *repository.PushTokenRepository,
	sysNotifRepo *repository.SystemNotificationRepository,
	appConfigRepo *repository.AppConfigRepository,
	ratingRepo *repository.ServiceRatingRepository,
	attendanceRepo *repository.ServiceAttendanceRepository,
	categoryRepo *repository.CategoryRepository,
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
	serviceHandler := handlers.NewServiceHandler(serviceRepo, categoryRepo, srRepo, ratingRepo, appConfigRepo)
	uploadService := services.NewUploadService(storageService)
	geoService := services.NewGeocodingService()
	pushService := services.NewExpoPushService()
	srHandler := handlers.NewServiceRequestHandler(srRepo, userRepo, regionRepo, teamRepo, sysNotifRepo, pushTokenRepo, pushService, uploadService, geoService, ratingRepo, attendanceRepo)
	geoHandler := handlers.NewGeolocationHandler()
	homeHandler := handlers.NewHomeHandler(srRepo, userRepo, geoService)
	newsHandler := handlers.NewNewsHandler(newsRepo, pushTokenRepo, sysNotifRepo, pushService, storageService)
	notificationHandler := handlers.NewNotificationHandler(pushTokenRepo, sysNotifRepo)
	teamHandler := handlers.NewTeamHandler(teamRepo, userRepo, srRepo)
	regionHandler := handlers.NewRegionHandler(regionRepo, userRepo)
	appConfigHandler := handlers.NewAppConfigHandler(appConfigRepo, storageService)
	ratingHandler := handlers.NewServiceRatingHandler(ratingRepo, srRepo)
	attendanceHandler := handlers.NewServiceAttendanceHandler(attendanceRepo, srRepo, uploadService, srHandler)
	categoryHandler := handlers.NewCategoryHandler(categoryRepo, serviceRepo, srRepo, teamRepo, ratingRepo, userRepo)

	// Routes under /api
	r.Route("/api", func(r chi.Router) {

		// Health check
		r.Head("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":    "ok",
				"timestamp": time.Now().UTC(),
			})
		})

		// Public auth routes
		r.Post("/auth/login", authHandler.Login)

		// Geolocation route
		r.Get("/geolocation", geoHandler.Search)

		// Public service routes (read-only)
		r.Get("/services", serviceHandler.ListServices)
		r.Get("/services/categories", serviceHandler.ListCategories)
		r.Get("/services/category/{category}", serviceHandler.ListServicesByCategory)
		r.Get("/services/category/id/{id}", serviceHandler.ListServicesByCategoryID)
		r.Get("/services/{id}", serviceHandler.GetService)
		r.Get("/services/{id}/ratings", ratingHandler.ListRatingsByService)
		r.Get("/services/{id}/rating-stats", ratingHandler.GetRatingStats)

		// Public news routes (read-only)
		r.Get("/news", newsHandler.ListNews)
		r.Get("/news/{id}", newsHandler.GetNews)

		// App Config route (public)
		r.Get("/app/config", appConfigHandler.GetMobileConfig)

		// Categories (public read-only)
		r.Get("/categories", categoryHandler.ListCategories)
		r.Get("/categories/{id}", categoryHandler.GetCategory)

		// Protected routes — any authenticated user
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(jwtSecret))

			// Auth
			r.Get("/auth/me", userHandler.GetMe)
			r.Post("/auth/logout", authHandler.Logout)
			r.Put("/auth/password", authHandler.ChangePassword)

			// Home
			r.Get("/home", homeHandler.Index)

			// My Team — current user's team info
			r.Get("/my-team", teamHandler.GetMyTeam)

			// Team members (admin or secretary of the team — checked in handler)
			r.Get("/teams/{id}/members", teamHandler.ListTeamMembers)
			r.Post("/teams/{id}/members", teamHandler.AddTeamMember)
			r.Delete("/teams/{id}/members/{userId}", teamHandler.RemoveTeamMember)

			// Service Requests — all authenticated users can create and view
			r.Post("/service-requests", srHandler.CreateServiceRequest)
			r.Get("/service-requests", srHandler.ListServiceRequests)
			r.Route("/service-requests/{id}", func(r chi.Router) {
				r.Get("/", srHandler.GetServiceRequest)
				r.Put("/", srHandler.UpdateServiceRequest)
				r.Patch("/status", srHandler.UpdateServiceRequestStatus)
				r.Delete("/", srHandler.DeleteServiceRequest)

				// Attendance (Handling)
				r.Post("/attendances", attendanceHandler.CreateAttendance)
				r.Get("/attendances", attendanceHandler.ListAttendances)
			})

			// Service Ratings
			r.Post("/ratings", ratingHandler.CreateRating)

			// Category Dashboard (admin OR secretary)
			r.Get("/categories/{id}/dashboard", categoryHandler.GetCategoryDetails)

			// Geocoding
			r.Get("/geocode-service-requests", srHandler.GeocodeAllServiceRequests)
			r.Get("/geocode-service-requests/{id}", srHandler.GeocodeServiceRequest)

			// Notifications
			r.Post("/notifications/push-tokens", notificationHandler.RegisterPushToken)
			r.Get("/notifications", notificationHandler.ListSystemNotifications)
			r.Post("/notifications", notificationHandler.CreateSystemNotification)
			r.Get("/notifications/{id}", notificationHandler.GetSystemNotification)
			r.Put("/notifications/{id}", notificationHandler.UpdateSystemNotification)
			r.Patch("/notifications/{id}/read", notificationHandler.MarkSystemNotificationAsRead)
			r.Patch("/notifications/read-all", notificationHandler.MarkAllSystemNotificationsAsRead)
			r.Delete("/notifications/{id}", notificationHandler.DeleteSystemNotification)

			// News write — Admin OR Marketing
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(userRepo, "admin", "marketing"))
				r.Post("/news", newsHandler.CreateNews)
				r.Post("/news/upload-image", newsHandler.UploadImage)
				r.Put("/news/{id}", newsHandler.UpdateNews)
				r.Delete("/news/{id}", newsHandler.DeleteNews)
			})

			// Users creation and listing (Admin and Secretary)
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(userRepo, "admin", "secretary"))
				r.Post("/users", userHandler.CreateUser)
				r.Get("/users", userHandler.ListUsers)
				r.Get("/users/{id}", userHandler.GetUser)
			})

			// Admin-only routes
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(userRepo, "admin"))

				// Users (Write ops except create)
				r.Put("/users/{id}", userHandler.UpdateUser)
				r.Delete("/users/{id}", userHandler.DeleteUser)
				r.Post("/users/{id}/profile-image", userHandler.UploadProfileImage)
				r.Delete("/users/{id}/profile-image", userHandler.DeleteProfileImage)

				// Regions
				r.Get("/regions", regionHandler.List)
				r.Post("/regions", regionHandler.Create)
				r.Get("/regions/{id}", regionHandler.Get)
				r.Put("/regions/{id}", regionHandler.Update)
				r.Delete("/regions/{id}", regionHandler.Delete)
				r.Get("/regions/bairro/{bairro}", regionHandler.FindByBairro)

				// Teams (CRUD — members are in the general group above for secretary access)
				r.Get("/teams", teamHandler.ListTeams)
				r.Post("/teams", teamHandler.CreateTeam)
				r.Get("/teams/{id}", teamHandler.GetTeam)
				r.Put("/teams/{id}", teamHandler.UpdateTeam)
				r.Delete("/teams/{id}", teamHandler.DeleteTeam)
				r.Get("/teams/{id}/stats", teamHandler.GetTeamDashboard)

				// Services (write)
				r.Post("/services", serviceHandler.CreateService)
				r.Put("/services/{id}", serviceHandler.UpdateService)
				r.Delete("/services/{id}", serviceHandler.DeleteService)

				// App Configuration
				r.Put("/app/settings/{key}", appConfigHandler.UpdateSetting)
				r.Post("/app/upload-image", appConfigHandler.UploadImage)
				r.Get("/app/banners", appConfigHandler.ListBanners)
				r.Post("/app/banners", appConfigHandler.CreateBanner)
				r.Put("/app/banners/{id}", appConfigHandler.UpdateBanner)
				r.Delete("/app/banners/{id}", appConfigHandler.DeleteBanner)

				// Categories (read-only dashboard for secretary too)
				r.Post("/categories", categoryHandler.CreateCategory)
				r.Put("/categories/{id}", categoryHandler.UpdateCategory)
				r.Delete("/categories/{id}", categoryHandler.DeleteCategory)
			})
		})
	})

	return r
}
