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

	// Buscar localizações geocodificadas
	list, err := h.srRepo.ListServiceRequests(r.Context(), "", categoryFilter, 1, 1000)
	if err == nil {
		for _, sr := range list {
			address := extractAddressFromRequestData(sr.RequestData)
			if address != "" {
				geoResult, _ := h.geoService.GeocodeAddress(address)
				if geoResult.Found {
					resp.MapLocations = append(resp.MapLocations, models.MapLocation{
						ID:           sr.ID,
						Address:      address,
						Latitude:     geoResult.Latitude,
						Longitude:    geoResult.Longitude,
						ServiceTitle: sr.ServiceTitle,
						Status:       sr.Status,
						Found:        geoResult.Found,
					})
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
