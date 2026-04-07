package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
	"github.com/laiirton/solucoes-urbanas-api/internal/repository"
)

type ServiceHandler struct {
	serviceRepo *repository.ServiceRepository
	srRepo      *repository.ServiceRequestRepository
}

func NewServiceHandler(serviceRepo *repository.ServiceRepository, srRepo *repository.ServiceRequestRepository) *ServiceHandler {
	return &ServiceHandler{serviceRepo: serviceRepo, srRepo: srRepo}
}

// GET /services
func (h *ServiceHandler) ListServices(w http.ResponseWriter, r *http.Request) {
	onlyActive := r.URL.Query().Get("all") != "true"
	search := r.URL.Query().Get("search")
	page, limit := parsePagination(r)

	services, err := h.serviceRepo.ListServices(r.Context(), onlyActive, search, page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list services")
		return
	}
	respondJSON(w, http.StatusOK, services)
}

// GET /services/{id}
func (h *ServiceHandler) GetService(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid service id")
		return
	}
	svc, err := h.serviceRepo.GetServiceByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "service not found")
		return
	}

	// Fetch metrics
	stats, _ := h.srRepo.GetServiceStatusStats(r.Context(), id)
	avgTime, _ := h.srRepo.GetAverageServiceTime(r.Context(), id)
	recent, _ := h.srRepo.ListServiceRequestDetailsByService(r.Context(), id, 1, 5)

	resp := models.ServiceDetailResponse{
		Service:            svc,
		AverageServiceTime: avgTime,
		StatusStats:        stats,
		RecentRequests:     recent,
	}

	respondJSON(w, http.StatusOK, resp)
}

// POST /services
func (h *ServiceHandler) CreateService(w http.ResponseWriter, r *http.Request) {
	var req models.CreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Title == "" || req.Category == "" {
		respondError(w, http.StatusBadRequest, "title and category are required")
		return
	}
	svc, err := h.serviceRepo.CreateService(r.Context(), &req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create service")
		return
	}
	respondJSON(w, http.StatusCreated, svc)
}

// PUT /services/{id}
func (h *ServiceHandler) UpdateService(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid service id")
		return
	}
	var req models.UpdateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	svc, err := h.serviceRepo.UpdateService(r.Context(), id, &req)
	if err != nil {
		respondError(w, http.StatusNotFound, "service not found or update failed")
		return
	}
	respondJSON(w, http.StatusOK, svc)
}

// DELETE /services/{id}
func (h *ServiceHandler) DeleteService(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid service id")
		return
	}
	if err := h.serviceRepo.DeleteService(r.Context(), id); err != nil {
		if strings.Contains(err.Error(), "service not found") {
			respondError(w, http.StatusNotFound, "service not found")
			return
		}
		// Probably a foreign key constraint (if service_requests exist)
		if strings.Contains(err.Error(), "violates foreign key constraint") {
			respondError(w, http.StatusConflict, "cannot delete service because it has associated service requests. try deactivating it instead.")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to delete service: "+err.Error())
		return
	}
	respondJSON(w, http.StatusOK, models.MessageResponse{Message: "service deleted successfully"})
}

