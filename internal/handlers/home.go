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
	regionFilter := GetRegionFilterForUser(user)
	teamFilter := GetTeamFilterForUserForUser(user)

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	var startDatePtr, endDatePtr *string
	if startDate != "" {
		startDatePtr = &startDate
	}
	if endDate != "" {
		endDatePtr = &endDate
	}

	resp, err := h.srRepo.GetHomeStats(r.Context(), isAdmin, userID, regionFilter, teamFilter, startDatePtr, endDatePtr)
	if err != nil {
		http.Error(w, "Error computing home stats", http.StatusInternalServerError)
		return
	}

	resp.MapLocations = []models.MapLocation{}
	list, err := h.srRepo.ListMapLocations(r.Context(), regionFilter, teamFilter, startDatePtr, endDatePtr, 1000)
	if err == nil {
		for _, loc := range list {
			if loc.Found {
				resp.MapLocations = append(resp.MapLocations, loc)
			} else if loc.Address != "" {
				go h.asyncGeocodeRequest(loc.ID, loc.Address)
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
