package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
	"github.com/laiirton/solucoes-urbanas-api/internal/repository"
)

type ServiceHandler struct {
	serviceRepo *repository.ServiceRepository
}

func NewServiceHandler(serviceRepo *repository.ServiceRepository) *ServiceHandler {
	return &ServiceHandler{serviceRepo: serviceRepo}
}

// GET /services
func (h *ServiceHandler) ListServices(w http.ResponseWriter, r *http.Request) {
	onlyActive := r.URL.Query().Get("all") != "true"
	services, err := h.serviceRepo.ListServices(r.Context(), onlyActive)
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
	respondJSON(w, http.StatusOK, svc)
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
		respondError(w, http.StatusNotFound, "service not found")
		return
	}
	respondJSON(w, http.StatusOK, models.MessageResponse{Message: "service deleted successfully"})
}
