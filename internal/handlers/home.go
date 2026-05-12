package handlers

import (
	"context"
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
	regionFilter := GetRegionFilterForAdmin(r.Context(), h.userRepo, userID)
	teamFilter := GetTeamFilterForUser(r.Context(), h.userRepo, userID)

	resp, err := h.srRepo.GetHomeStats(r.Context(), isAdmin, userID, regionFilter, teamFilter)
	if err != nil {
		http.Error(w, "Error computing home stats", http.StatusInternalServerError)
		return
	}

	resp.MapLocations = []models.MapLocation{}
	list, err := h.srRepo.ListServiceRequests(r.Context(), "", "", regionFilter, teamFilter, 1, 1000)
	if err == nil {
		for _, sr := range list {
			var lat, lon float64
			var geoAddr string
			found := false

			if sr.Latitude != nil && sr.Longitude != nil {
				lat = *sr.Latitude
				lon = *sr.Longitude
				found = true
				if sr.GeocodedAddress != nil {
					geoAddr = *sr.GeocodedAddress
				}
			} else {
				address := extractAddressFromRequestData(sr.RequestData)
				if address != "" {
					geoAddr = address
					go h.asyncGeocodeRequest(sr.ID, address)
				}
			}

			if found || geoAddr != "" {
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
					Found:        found,
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *HomeHandler) asyncGeocodeRequest(id int64, address string) {
	geoResult, err := h.geoService.GeocodeAddress(address)
	if err != nil || !geoResult.Found {
		return
	}

	ctx := context.Background()
	h.srRepo.SaveGeocoding(ctx, id, geoResult.Latitude, geoResult.Longitude, geoResult.DisplayName)
}
