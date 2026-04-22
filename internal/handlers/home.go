package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/laiirton/solucoes-urbanas-api/internal/middleware"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
	"github.com/laiirton/solucoes-urbanas-api/internal/repository"
	"github.com/laiirton/solucoes-urbanas-api/internal/services"
)

type HomeHandler struct {
	srRepo     *repository.ServiceRequestRepository
	userRepo   *repository.UserRepository
	geoService *services.GeocodingService
}

func NewHomeHandler(srRepo *repository.ServiceRequestRepository, userRepo *repository.UserRepository, geoService *services.GeocodingService) *HomeHandler {
	return &HomeHandler{
		srRepo:     srRepo,
		userRepo:   userRepo,
		geoService: geoService,
	}
}

func (h *HomeHandler) Index(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userRepo.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	isAdmin := user.Type != nil && *user.Type == "admin"
	var categoryFilter string
	if isAdmin && user.Team != nil {
		categoryFilter = user.Team.ServiceCategory
	}

	resp, err := h.srRepo.GetHomeStats(r.Context(), isAdmin, userID, categoryFilter)
	if err != nil {
		http.Error(w, "Error computing home stats", http.StatusInternalServerError)
		return
	}

	resp.MapLocations = []models.MapLocation{}
	list, err := h.srRepo.ListServiceRequests(r.Context(), "", categoryFilter, 1, 1000)
	if err == nil {
		for _, sr := range list {
			var lat, lon float64
			var found bool
			var geoAddr string

			// Use persisted coordinates if available
			if sr.Latitude != nil && sr.Longitude != nil {
				lat = *sr.Latitude
				lon = *sr.Longitude
				found = true
				if sr.GeocodedAddress != nil {
					geoAddr = *sr.GeocodedAddress
				}
			} else {
				// Geocode and save to DB
				address := extractAddressFromRequestData(sr.RequestData)
				if address != "" {
					geoResult, _ := h.geoService.GeocodeAddress(address)
					if geoResult.Found {
						lat = geoResult.Latitude
						lon = geoResult.Longitude
						found = true
						geoAddr = geoResult.DisplayName
						// Save to DB
						h.srRepo.SaveGeocoding(r.Context(), sr.ID, lat, lon, geoAddr)
					}
				}
			}

			if found {
				icon := ""
				if sr.ServiceID != nil {
					icon = models.GetServiceIcon(*sr.ServiceID)
				}

				resp.MapLocations = append(resp.MapLocations, models.MapLocation{
					ID:           sr.ID,
					Address:      geoAddr,
					Latitude:     lat,
					Longitude:    lon,
					ServiceTitle: sr.ServiceTitle,
					Status:       sr.Status,
					Icon:         icon,
					Found:        true,
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
