package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
	"github.com/laiirton/solucoes-urbanas-api/internal/repository"
)

type ServiceHandler struct {
	serviceRepo   *repository.ServiceRepository
	srRepo        *repository.ServiceRequestRepository
	ratingRepo    *repository.ServiceRatingRepository
	appConfigRepo *repository.AppConfigRepository
}

func NewServiceHandler(serviceRepo *repository.ServiceRepository, srRepo *repository.ServiceRequestRepository, ratingRepo *repository.ServiceRatingRepository, appConfigRepo *repository.AppConfigRepository) *ServiceHandler {
	return &ServiceHandler{serviceRepo: serviceRepo, srRepo: srRepo, ratingRepo: ratingRepo, appConfigRepo: appConfigRepo}
}

func (h *ServiceHandler) getAllowedCategories(r *http.Request) []string {
	if h.appConfigRepo == nil {
		return nil
	}
	showAll := r.URL.Query().Get("all") == "true"
	if showAll {
		return nil
	}
	categories, _ := h.appConfigRepo.GetMobileCategories(r.Context())
	return categories
}

func (h *ServiceHandler) getAllowedServices(r *http.Request) []int64 {
	if h.appConfigRepo == nil {
		return nil
	}
	showAll := r.URL.Query().Get("all") == "true"
	if showAll {
		return nil
	}
	services, _ := h.appConfigRepo.GetMobileServices(r.Context())
	return services
}

// GET /services
func (h *ServiceHandler) ListServices(w http.ResponseWriter, r *http.Request) {
	onlyActive := r.URL.Query().Get("all") != "true"
	search := r.URL.Query().Get("search")
	page, limit := parsePagination(r)

	services, err := h.serviceRepo.ListServices(r.Context(), onlyActive, search, page, limit, h.getAllowedCategories(r), h.getAllowedServices(r))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list services")
		return
	}
	respondJSON(w, http.StatusOK, services)
}

// GET /services/category/{category}
func (h *ServiceHandler) ListServicesByCategory(w http.ResponseWriter, r *http.Request) {
	category := chi.URLParam(r, "category")
	if category == "" {
		respondError(w, http.StatusBadRequest, "category is required")
		return
	}

	allowed := h.getAllowedCategories(r)
	if len(allowed) > 0 {
		found := false
		for _, c := range allowed {
			if c == category {
				found = true
				break
			}
		}
		if !found {
			respondError(w, http.StatusNotFound, "category not available")
			return
		}
	}

	onlyActive := r.URL.Query().Get("all") != "true"

	services, err := h.serviceRepo.ListServicesByCategory(r.Context(), category, onlyActive, nil)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list services by category")
		return
	}
	respondJSON(w, http.StatusOK, services)
}

// GET /services/category/id/{id}
func (h *ServiceHandler) ListServicesByCategoryID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid category id")
		return
	}

	onlyActive := r.URL.Query().Get("all") != "true"

	services, err := h.serviceRepo.ListServicesByCategoryID(r.Context(), id, onlyActive, nil)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list services by category id")
		return
	}
	respondJSON(w, http.StatusOK, services)
}

// GET /services/categories
func (h *ServiceHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	onlyActive := r.URL.Query().Get("all") != "true"

	categories, err := h.serviceRepo.ListCategories(r.Context(), onlyActive, h.getAllowedCategories(r))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list categories")
		return
	}

	// Add icons to categories
	var categoriesWithIcon []models.CategorySummary
	for _, cat := range categories {
		categoriesWithIcon = append(categoriesWithIcon, models.CategorySummary{
			Name: cat,
			Icon: models.GetCategoryIcon(cat),
		})
	}

	respondJSON(w, http.StatusOK, categoriesWithIcon)
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

	// Fetch metrics (log errors but don't fail the whole request)
	stats, err := h.srRepo.GetServiceStatusStats(r.Context(), id)
	if err != nil {
		log.Printf("Warning: failed to fetch status stats for service %d: %v", id, err)
		stats = []models.StatusStat{}
	}
	avgTime, err := h.srRepo.GetAverageServiceTime(r.Context(), id)
	if err != nil {
		log.Printf("Warning: failed to fetch avg service time for service %d: %v", id, err)
	}
	recent, err := h.srRepo.ListServiceRequestDetailsByService(r.Context(), id, 1, 5)
	if err != nil {
		log.Printf("Warning: failed to fetch recent requests for service %d: %v", id, err)
		recent = []*models.ServiceRequestDetailResponse{}
	}
	ratingStats, err := h.ratingRepo.GetStatsByServiceID(r.Context(), id)
	if err != nil {
		log.Printf("Warning: failed to fetch rating stats for service %d: %v", id, err)
	}
	recentRatings, err := h.ratingRepo.ListByServiceID(r.Context(), id, 3, 0)
	if err != nil {
		log.Printf("Warning: failed to fetch recent ratings for service %d: %v", id, err)
		recentRatings = []*models.ServiceRatingResponse{}
	}

	resp := models.ServiceDetailResponse{
		Service:            svc,
		AverageServiceTime: avgTime,
		RatingStats:        ratingStats,
		StatusStats:        stats,
		RecentRequests:     recent,
		RecentRatings:      recentRatings,
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
