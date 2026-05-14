package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
	"github.com/laiirton/solucoes-urbanas-api/internal/repository"
)

type CategoryHandler struct {
	repo        *repository.CategoryRepository
	serviceRepo *repository.ServiceRepository
	srRepo      *repository.ServiceRequestRepository
	teamRepo    *repository.TeamRepository
	ratingRepo  *repository.ServiceRatingRepository
}

func NewCategoryHandler(
	repo *repository.CategoryRepository,
	serviceRepo *repository.ServiceRepository,
	srRepo *repository.ServiceRequestRepository,
	teamRepo *repository.TeamRepository,
	ratingRepo *repository.ServiceRatingRepository,
) *CategoryHandler {
	return &CategoryHandler{
		repo:        repo,
		serviceRepo: serviceRepo,
		srRepo:      srRepo,
		teamRepo:    teamRepo,
		ratingRepo:  ratingRepo,
	}
}

// GET /categories
func (h *CategoryHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	onlyActive := r.URL.Query().Get("all") != "true"

	categories, err := h.repo.List(r.Context(), onlyActive)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list categories")
		return
	}
	respondJSON(w, http.StatusOK, categories)
}

// GET /categories/{id}
func (h *CategoryHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid category id")
		return
	}

	cat, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "category not found")
		return
	}

	onlyActive := r.URL.Query().Get("all") != "true"
	services, err := h.serviceRepo.ListServicesByCategoryID(r.Context(), id, onlyActive, nil)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list category services")
		return
	}

	resp := models.CategoryDetailResponse{
		Category: cat,
		Services: services,
	}
	respondJSON(w, http.StatusOK, resp)
}

// GET /categories/{id}/dashboard
func (h *CategoryHandler) GetCategoryDetails(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid category id")
		return
	}

	// 1. Get Category
	cat, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "category not found")
		return
	}

	// 2. Get Services
	services, err := h.serviceRepo.ListServicesByCategoryID(r.Context(), id, false, nil)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list category services")
		return
	}

	// 3. Get Requests
	requests, err := h.srRepo.ListServiceRequestsByCategory(r.Context(), cat.Name, 0)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list category requests")
		return
	}

	// 4. Calculate KPIs & Status Distribution
	totalRequests := len(requests)
	statusDist := models.CategoryStatusDistribution{}
	var totalResolutionDays float64
	var completedCount int

	for _, req := range requests {
		switch req.Status {
		case "pending":
			statusDist.Pending++
		case "in_progress":
			statusDist.InProgress++
		case "completed":
			statusDist.Completed++
			completedCount++
			totalResolutionDays += req.UpdatedAt.Sub(req.CreatedAt).Hours() / 24
		case "cancelled":
			statusDist.Cancelled++
		}
	}

	avgResolutionDays := 0.0
	if completedCount > 0 {
		avgResolutionDays = totalResolutionDays / float64(completedCount)
	}

	// 5. Calculate Service Details
	serviceDetails := make([]models.CategoryServiceDetail, len(services))
	for i, s := range services {
		serviceRequests := 0
		serviceCompleted := 0
		for _, req := range requests {
			if req.ServiceID != nil && *req.ServiceID == s.ID {
				serviceRequests++
				if req.Status == "completed" {
					serviceCompleted++
				}
			}
		}
		
		avgRating := 0.0
		stats, err := h.ratingRepo.GetStatsByServiceID(r.Context(), s.ID)
		if err == nil {
			avgRating = stats.Average
		}

		serviceDetails[i] = models.CategoryServiceDetail{
			ID:                s.ID,
			Title:             s.Title,
			IsActive:          s.IsActive,
			TotalRequests:     serviceRequests,
			CompletedRequests: serviceCompleted,
			AverageRating:     avgRating,
		}
	}

	// 6. Get Teams (by work_area of secretaries)
	teamsData, err := h.teamRepo.ListTeamsByWorkArea(r.Context(), cat.Name)
	teams := []models.CategoryTeamDetail{}
	if err == nil {
		for _, t := range teamsData {
			teams = append(teams, models.CategoryTeamDetail{
				ID:          t.ID,
				Name:        t.Name,
				Description: t.Description,
			})
		}
	}

	// 7. Recent Requests
	recentRequests := []models.CategoryRecentRequest{}
	limit := 10
	if len(requests) < limit {
		limit = len(requests)
	}
	for i := 0; i < limit; i++ {
		req := requests[i]
		recentRequests = append(recentRequests, models.CategoryRecentRequest{
			ID:           req.ID,
			ServiceTitle: req.ServiceTitle,
			Status:       req.Status,
			Address:      req.GeocodedAddress,
			CreatedAt:    req.CreatedAt,
		})
	}

	categoryAvgRating, _ := h.ratingRepo.GetAverageRatingByCategoryID(r.Context(), id)

	resp := models.CategoryDashboardResponse{
		Category: models.CategoryInfo{
			Name: cat.Name,
			Icon: cat.Icon,
		},
		KPIs: models.CategoryKPIs{
			TotalServices:         len(services),
			TotalTeams:            len(teams),
			TotalRequests:         totalRequests,
			AverageRating:         categoryAvgRating,
			AverageResolutionDays: avgResolutionDays,
		},
		StatusDistribution: statusDist,
		Services:           serviceDetails,
		Teams:              teams,
		RecentRequests:     recentRequests,
	}

	respondJSON(w, http.StatusOK, resp)
}

// POST /categories
func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req models.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}

	cat, err := h.repo.Create(r.Context(), &req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create category: "+err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, cat)
}

// PUT /categories/{id}
func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid category id")
		return
	}

	var req models.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cat, err := h.repo.Update(r.Context(), id, &req)
	if err != nil {
		respondError(w, http.StatusNotFound, "category not found or update failed")
		return
	}

	if req.IsActive != nil && !*req.IsActive {
		err := h.serviceRepo.DeactivateServicesByCategoryID(r.Context(), id)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "category updated but failed to deactivate services: "+err.Error())
			return
		}
	}

	respondJSON(w, http.StatusOK, cat)
}

// DELETE /categories/{id}
func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid category id")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if err.Error() == "category not found" {
			respondError(w, http.StatusNotFound, "category not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to delete category: "+err.Error())
		return
	}
	respondJSON(w, http.StatusOK, models.MessageResponse{Message: "category deleted successfully"})
}
