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

	// 5. Get Mobile Categories
	mobileCategories, _ := h.repo.GetMobileCategories(ctx)
	if mobileCategories == nil {
		mobileCategories = []string{}
	}

	// 6. Get Mobile Services
	mobileServices, _ := h.repo.GetMobileServices(ctx)
	if mobileServices == nil {
		mobileServices = []int64{}
	}

	// Filter featured content by mobile (only show what's actually available)
	if len(mobileCategories) > 0 {
		mobileCatSet := make(map[string]bool, len(mobileCategories))
		for _, c := range mobileCategories {
			mobileCatSet[c] = true
		}
		filtered := make([]models.CategorySummary, 0, len(featuredCategories))
		for _, fc := range featuredCategories {
			if mobileCatSet[fc.Name] {
				filtered = append(filtered, fc)
			}
		}
		featuredCategories = filtered
	}

	if len(mobileServices) > 0 {
		mobileSvcSet := make(map[int64]bool, len(mobileServices))
		for _, id := range mobileServices {
			mobileSvcSet[id] = true
		}
		filtered := make([]models.ServiceSummary, 0, len(featuredServices))
		for _, fs := range featuredServices {
			if mobileSvcSet[fs.ID] {
				filtered = append(filtered, fs)
			}
		}
		featuredServices = filtered
	}

	// Build the response
	response := models.MobileHomeResponse{
		LogoURL: logoURL,
		Banners: banners,
		Sections: []models.Section{
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
		MobileCategories: mobileCategories,
		MobileServices:   mobileServices,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AppConfigHandler) UpdateSetting(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	var value interface{}
	if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
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
		respondError(w, http.StatusInternalServerError, "error updating setting")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AppConfigHandler) CreateBanner(w http.ResponseWriter, r *http.Request) {
	var banner models.AppBanner
	if err := json.NewDecoder(r.Body).Decode(&banner); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.repo.CreateBanner(r.Context(), &banner); err != nil {
		respondError(w, http.StatusInternalServerError, "error creating banner")
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
		respondError(w, http.StatusBadRequest, "invalid request body")
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
		respondError(w, http.StatusInternalServerError, "error updating banner")
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
		respondError(w, http.StatusInternalServerError, "error deleting banner")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AppConfigHandler) ListBanners(w http.ResponseWriter, r *http.Request) {
	banners, err := h.repo.GetBanners(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "error fetching banners")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(banners)
}

func (h *AppConfigHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "unable to parse form")
		return
	}

	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		respondError(w, http.StatusBadRequest, "image is required")
		return
	}
	defer file.Close()

	ext := filepath.Ext(fileHeader.Filename)
	filename := "app_config/" + uuid.New().String() + ext

	if h.storage == nil {
		respondError(w, http.StatusInternalServerError, "storage service not configured")
		return
	}

	imageURL, uploadErr := h.storage.UploadFile(file, filename, fileHeader.Header.Get("Content-Type"))
	if uploadErr != nil {
		respondError(w, http.StatusInternalServerError, "failed to upload image")
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
