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

type RegionHandler struct {
	regionRepo *repository.RegionRepository
	userRepo   *repository.UserRepository
}

func NewRegionHandler(regionRepo *repository.RegionRepository, userRepo *repository.UserRepository) *RegionHandler {
	return &RegionHandler{regionRepo: regionRepo, userRepo: userRepo}
}

// POST /regions
func (h *RegionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRegionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}
	if len(req.Neighborhoods) == 0 {
		req.Neighborhoods = json.RawMessage("[]")
	}

	region, err := h.regionRepo.Create(r.Context(), &req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, region)
}

// GET /regions
func (h *RegionHandler) List(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	page, limit := parsePagination(r)

	regions, err := h.regionRepo.List(r.Context(), search, page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list regions")
		return
	}
	respondJSON(w, http.StatusOK, regions)
}

// GET /regions/{id}
func (h *RegionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid region id")
		return
	}
	region, err := h.regionRepo.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "region not found")
		return
	}
	respondJSON(w, http.StatusOK, region)
}

// PUT /regions/{id}
func (h *RegionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid region id")
		return
	}
	var req models.UpdateRegionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	region, err := h.regionRepo.Update(r.Context(), id, &req)
	if err != nil {
		respondError(w, http.StatusNotFound, "region not found")
		return
	}
	respondJSON(w, http.StatusOK, region)
}

// DELETE /regions/{id}
func (h *RegionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid region id")
		return
	}
	// Check auth
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.userRepo.GetUserByID(r.Context(), userID)
	if err != nil || user.Type == nil || *user.Type != "admin" {
		respondError(w, http.StatusForbidden, "only admins can delete regions")
		return
	}

	if err := h.regionRepo.Delete(r.Context(), id); err != nil {
		respondError(w, http.StatusNotFound, "region not found")
		return
	}
	respondJSON(w, http.StatusOK, models.MessageResponse{Message: "region deleted successfully"})
}

// GET /regions/bairro/{bairro} - Find region by neighborhood name
func (h *RegionHandler) FindByBairro(w http.ResponseWriter, r *http.Request) {
	bairro := chi.URLParam(r, "bairro")
	if bairro == "" {
		respondError(w, http.StatusBadRequest, "neighborhood name is required")
		return
	}
	region, err := h.regionRepo.FindByNeighborhood(r.Context(), bairro)
	if err != nil {
		respondError(w, http.StatusNotFound, "no region found for this neighborhood")
		return
	}
	respondJSON(w, http.StatusOK, region)
}
