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

type ServiceRatingHandler struct {
	ratingRepo *repository.ServiceRatingRepository
	srRepo     *repository.ServiceRequestRepository
}

func NewServiceRatingHandler(ratingRepo *repository.ServiceRatingRepository, srRepo *repository.ServiceRequestRepository) *ServiceRatingHandler {
	return &ServiceRatingHandler{
		ratingRepo: ratingRepo,
		srRepo:     srRepo,
	}
}

// POST /ratings
func (h *ServiceRatingHandler) CreateRating(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.CreateServiceRatingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Stars < 1 || req.Stars > 5 {
		respondError(w, http.StatusBadRequest, "stars must be between 1 and 5")
		return
	}

	// Verify service request belongs to user and is completed
	sr, err := h.srRepo.GetServiceRequestByID(r.Context(), req.ServiceRequestID)
	if err != nil {
		respondError(w, http.StatusNotFound, "service request not found")
		return
	}

	if sr.UserID == nil || *sr.UserID != userID {
		respondError(w, http.StatusForbidden, "you can only rate your own service requests")
		return
	}

	if sr.Status != "completed" {
		respondError(w, http.StatusBadRequest, "you can only rate completed service requests")
		return
	}

	if sr.ServiceID == nil {
		respondError(w, http.StatusBadRequest, "service request has no associated service")
		return
	}

	// Check if already rated
	existing, _ := h.ratingRepo.GetByRequestID(r.Context(), req.ServiceRequestID)
	if existing != nil {
		respondError(w, http.StatusConflict, "this service request has already been rated")
		return
	}

	rating := &models.ServiceRating{
		ServiceRequestID: req.ServiceRequestID,
		ServiceID:        *sr.ServiceID,
		UserID:           userID,
		Stars:            req.Stars,
		Comment:          req.Comment,
	}

	if err := h.ratingRepo.Create(r.Context(), rating); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, rating)
}

// GET /services/{id}/ratings
func (h *ServiceRatingHandler) ListRatingsByService(w http.ResponseWriter, r *http.Request) {
	serviceID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid service id")
		return
	}

	page, limit := parsePagination(r)
	offset := (page - 1) * limit

	ratings, err := h.ratingRepo.ListByServiceID(r.Context(), serviceID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list ratings")
		return
	}

	respondJSON(w, http.StatusOK, ratings)
}

// GET /services/{id}/rating-stats
func (h *ServiceRatingHandler) GetRatingStats(w http.ResponseWriter, r *http.Request) {
	serviceID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid service id")
		return
	}

	stats, err := h.ratingRepo.GetStatsByServiceID(r.Context(), serviceID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get stats")
		return
	}

	respondJSON(w, http.StatusOK, stats)
}
