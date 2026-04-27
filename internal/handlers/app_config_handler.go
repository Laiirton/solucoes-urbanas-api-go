package handlers

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
	"github.com/laiirton/solucoes-urbanas-api/internal/repository"
	"github.com/laiirton/solucoes-urbanas-api/internal/services"
)

type AppConfigHandler struct {
	repo    *repository.AppConfigRepository
	storage services.StorageService
}

func NewAppConfigHandler(repo *repository.AppConfigRepository, storage services.StorageService) *AppConfigHandler {
	return &AppConfigHandler{repo: repo, storage: storage}
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

	// Image cleanup for logo_url
	if key == "logo_url" {
		var oldLogoURL string
		if err := h.repo.GetSetting(r.Context(), "logo_url", &oldLogoURL); err == nil && oldLogoURL != "" {
			newLogoURL, ok := value.(string)
			if ok && newLogoURL != oldLogoURL {
				h.deleteFileIfInternal(oldLogoURL)
			}
		}
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

	// Image cleanup
	if existing, err := h.repo.GetBannerByID(r.Context(), id); err == nil {
		if existing.ImageURL != banner.ImageURL {
			h.deleteFileIfInternal(existing.ImageURL)
		}
	}

	if err := h.repo.UpdateBanner(r.Context(), &banner); err != nil {
		http.Error(w, "Error updating banner", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AppConfigHandler) DeleteBanner(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	// Image cleanup
	if existing, err := h.repo.GetBannerByID(r.Context(), id); err == nil {
		h.deleteFileIfInternal(existing.ImageURL)
	}

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

func (h *AppConfigHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Image is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := filepath.Ext(fileHeader.Filename)
	filename := "app_config/" + uuid.New().String() + ext

	if h.storage == nil {
		http.Error(w, "Storage service not configured", http.StatusInternalServerError)
		return
	}

	imageURL, uploadErr := h.storage.UploadFile(file, filename, fileHeader.Header.Get("Content-Type"))
	if uploadErr != nil {
		http.Error(w, "Failed to upload image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": imageURL})
}

func (h *AppConfigHandler) deleteFileIfInternal(url string) {
	if h.storage == nil || url == "" {
		return
	}

	// Simple check to see if the URL belongs to our Supabase storage
	// You can make this more robust if needed by checking the domain
	if strings.Contains(url, "/storage/v1/object/public/") {
		_ = h.storage.DeleteFile(url)
	}
}
