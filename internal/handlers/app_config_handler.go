package handlers

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

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

	var (
		logoURL            string
		banners            []models.AppBanner
		featuredServices   []models.ServiceSummary
		featuredCategories []models.CategorySummary
		wg                 sync.WaitGroup
	)

	wg.Add(4)
	go func() {
		defer wg.Done()
		if err := h.repo.GetSetting(ctx, "logo_url", &logoURL); err != nil {
			logoURL = ""
		}
	}()
	go func() {
		defer wg.Done()
		b, e := h.repo.GetBanners(ctx, true)
		if e != nil {
			banners = []models.AppBanner{}
		} else {
			banners = b
		}
	}()
	go func() {
		defer wg.Done()
		fs, e := h.repo.GetFeaturedServices(ctx)
		if e != nil {
			featuredServices = []models.ServiceSummary{}
		} else {
			featuredServices = fs
		}
	}()
	go func() {
		defer wg.Done()
		fc, e := h.repo.GetFeaturedCategories(ctx)
		if e != nil {
			featuredCategories = []models.CategorySummary{}
		} else {
			featuredCategories = fc
		}
	}()
	wg.Wait()

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
		Categories: featuredCategories,
		Services:   featuredServices,
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
	banners, err := h.repo.GetBanners(r.Context(), false)
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

	if strings.Contains(url, "/storage/v1/object/public/") {
		_ = h.storage.DeleteFile(url)
	}
}
