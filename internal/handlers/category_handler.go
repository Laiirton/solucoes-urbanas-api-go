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
}

func NewCategoryHandler(repo *repository.CategoryRepository, serviceRepo *repository.ServiceRepository) *CategoryHandler {
	return &CategoryHandler{repo: repo, serviceRepo: serviceRepo}
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
