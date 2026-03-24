package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/laiirton/solucoes-urbanas-api/internal/middleware"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
	"github.com/laiirton/solucoes-urbanas-api/internal/repository"
)

type ServiceRequestHandler struct {
	srRepo *repository.ServiceRequestRepository
}

func NewServiceRequestHandler(srRepo *repository.ServiceRequestRepository) *ServiceRequestHandler {
	return &ServiceRequestHandler{srRepo: srRepo}
}

// POST /service-requests
func (h *ServiceRequestHandler) CreateServiceRequest(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.CreateServiceRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ServiceID == 0 || req.ServiceTitle == "" {
		respondError(w, http.StatusBadRequest, "service_id and service_title are required")
		return
	}
	if len(req.RequestData) == 0 {
		req.RequestData = []byte("{}")
	}

	sr, err := h.srRepo.CreateServiceRequest(r.Context(), &userID, &req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, sr)
}

// GET /service-requests
func (h *ServiceRequestHandler) ListServiceRequests(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// By default, list only the authenticated user's requests
	// Pass ?all=true for admins to see all (simple check — no role system yet)
	var list []*models.ServiceRequest
	var err error
	if r.URL.Query().Get("all") == "true" {
		list, err = h.srRepo.ListServiceRequests(r.Context())
	} else {
		list, err = h.srRepo.ListServiceRequestsByUser(r.Context(), userID)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list service requests")
		return
	}
	respondJSON(w, http.StatusOK, list)
}

// GET /service-requests/{id}
func (h *ServiceRequestHandler) GetServiceRequest(w http.ResponseWriter, r *http.Request) {
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
	respondJSON(w, http.StatusOK, sr)
}

// PATCH /service-requests/{id}/status
func (h *ServiceRequestHandler) UpdateServiceRequestStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid service request id")
		return
	}

	var req models.UpdateServiceRequestStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Status == "" {
		respondError(w, http.StatusBadRequest, "status is required")
		return
	}

	sr, err := h.srRepo.UpdateServiceRequestStatus(r.Context(), id, req.Status)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, sr)
}

// DELETE /service-requests/{id}
func (h *ServiceRequestHandler) DeleteServiceRequest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid service request id")
		return
	}
	if err := h.srRepo.DeleteServiceRequest(r.Context(), id); err != nil {
		respondError(w, http.StatusNotFound, "service request not found")
		return
	}
	respondJSON(w, http.StatusOK, models.MessageResponse{Message: "service request deleted successfully"})
}
