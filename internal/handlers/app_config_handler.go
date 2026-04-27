package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
	"github.com/laiirton/solucoes-urbanas-api/internal/repository"
)

type AppConfigHandler struct {
	repo *repository.AppConfigRepository
}

func NewAppConfigHandler(repo *repository.AppConfigRepository) *AppConfigHandler {
	return &AppConfigHandler{repo: repo}
}

func (h *AppConfigHandler) GetMobileConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Get Logo
	var logoURL string
	if err := h.repo.GetSetting(ctx, "logo_url", &logoURL); err != nil {
		logoURL = "" // Default or empty
	}

	// 2. Get Banners
	banners, err := h.repo.GetBanners(ctx)
	if err != nil {
		banners = []models.AppBanner{}
	}

	// 3. Get Featured Services
	featuredServices, err := h.repo.GetFeaturedServices(ctx)
	if err != nil {
		featuredServices = []models.ServiceSummary{}
	}

	// 4. Get Featured Categories
	featuredCategories, err := h.repo.GetFeaturedCategories(ctx)
	if err != nil {
		featuredCategories = []models.CategorySummary{}
	}

	// Build the response
	response := models.MobileHomeResponse{
		LogoURL: logoURL,
		Banners: banners,
		Sections: []models.Section{
			{
				Type:  "banners",
				Title: "Destaques",
				Data:  banners,
			},
			{
				Type:  "categories",
				Title: "Categorias",
				Data:  featuredCategories,
			},
			{
				Type:  "services",
				Title: "Serviços em Destaque",
				Data:  featuredServices,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AppConfigHandler) UpdateSetting(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	var value interface{}
	if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.repo.UpdateSetting(r.Context(), key, value); err != nil {
		http.Error(w, "Error updating setting", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AppConfigHandler) CreateBanner(w http.ResponseWriter, r *http.Request) {
	var banner models.AppBanner
	if err := json.NewDecoder(r.Body).Decode(&banner); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.repo.CreateBanner(r.Context(), &banner); err != nil {
		http.Error(w, "Error creating banner", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(banner)
}

func (h *AppConfigHandler) UpdateBanner(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	var banner models.AppBanner
	if err := json.NewDecoder(r.Body).Decode(&banner); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	banner.ID = id

	if err := h.repo.UpdateBanner(r.Context(), &banner); err != nil {
		http.Error(w, "Error updating banner", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AppConfigHandler) DeleteBanner(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	if err := h.repo.DeleteBanner(r.Context(), id); err != nil {
		http.Error(w, "Error deleting banner", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AppConfigHandler) ListBanners(w http.ResponseWriter, r *http.Request) {
	banners, err := h.repo.GetBanners(r.Context())
	if err != nil {
		http.Error(w, "Error fetching banners", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(banners)
}
