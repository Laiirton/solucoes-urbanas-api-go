package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/laiirton/solucoes-urbanas-api/internal/middleware"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
)

// GET /service-requests
func (h *ServiceRequestHandler) ListServiceRequests(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	search := r.URL.Query().Get("search")
	status := r.URL.Query().Get("status")
	page, limit := parsePagination(r)

	user, err := h.userRepo.GetUserByID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	regionFilter := GetRegionFilterForUser(user)
	teamFilter := GetTeamFilterForUserForUser(user)
	userRole := GetUserRoleForUser(user)

	var list []*models.ServiceRequest

	if teamFilter != nil && userRole != nil && (*userRole == "attendant" || *userRole == "secretary") {
		list, err = h.srRepo.ListServiceRequestsByTeam(r.Context(), *teamFilter, search, status, page, limit)
	} else if r.URL.Query().Get("all") == "true" {
		list, err = h.srRepo.ListServiceRequests(r.Context(), search, status, regionFilter, teamFilter, nil, nil, page, limit)
	} else {
		list, err = h.srRepo.ListServiceRequestsByUser(r.Context(), userID, search, status, regionFilter, page, limit)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list service requests")
		return
	}
	respondJSON(w, http.StatusOK, list)
}

// GET /service-requests/{id}/geocode - Geocodifica o endereço do service request
func (h *ServiceRequestHandler) GeocodeServiceRequest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid service request id")
		return
	}

	sr, err := h.srRepo.GetServiceRequestByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "service request not found")
		return
	}

	address := extractAddressFromRequestData(sr.RequestData)
	if address == "" {
		respondError(w, http.StatusBadRequest, "no address found in service request data")
		return
	}

	geoResult, err := h.geoService.GeocodeAddress(address)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to geocode address")
		return
	}

	response := map[string]interface{}{
		"service_request_id": id,
		"address":            address,
		"latitude":           geoResult.Latitude,
		"longitude":          geoResult.Longitude,
		"display_name":       geoResult.DisplayName,
		"found":              geoResult.Found,
	}

	respondJSON(w, http.StatusOK, response)
}

// GET /service-requests/geocode-all - Retorna todos os service requests com coordenadas para o mapa
func (h *ServiceRequestHandler) GeocodeAllServiceRequests(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	search := r.URL.Query().Get("search")
	page, limit := parsePagination(r)

	user, err := h.userRepo.GetUserByID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	regionFilter := GetRegionFilterForUser(user)
	teamFilter := GetTeamFilterForUserForUser(user)

	var list []*models.ServiceRequest
	if r.URL.Query().Get("all") == "true" {
		list, err = h.srRepo.ListServiceRequests(r.Context(), search, "", regionFilter, teamFilter, nil, nil, page, limit)
	} else {
		list, err = h.srRepo.ListServiceRequestsByUser(r.Context(), userID, search, "", regionFilter, page, limit)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list service requests")
		return
	}

	// Use stored coordinates from database instead of re-geocoding
	var locations []models.MapLocation
	for _, sr := range list {
		if sr.Latitude == nil || sr.Longitude == nil {
			continue
		}
		icon := ""
		if sr.ServiceID != nil {
			icon = models.GetServiceIcon(*sr.ServiceID)
		}
		address := ""
		if sr.GeocodedAddress != nil {
			address = *sr.GeocodedAddress
		}
		locations = append(locations, models.MapLocation{
			ID:           sr.ID,
			Address:      address,
			Latitude:     *sr.Latitude,
			Longitude:    *sr.Longitude,
			ServiceTitle: sr.ServiceTitle,
			Status:       sr.Status,
			Icon:         icon,
			Found:        true,
		})
	}

	respondJSON(w, http.StatusOK, locations)
}
