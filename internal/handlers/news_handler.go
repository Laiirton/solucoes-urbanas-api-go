package handlers

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/laiirton/solucoes-urbanas-api/internal/middleware"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
	"github.com/laiirton/solucoes-urbanas-api/internal/repository"
	"github.com/laiirton/solucoes-urbanas-api/internal/services"
)

type NewsHandler struct {
	repo    *repository.NewsRepository
	storage services.StorageService
}

func NewNewsHandler(repo *repository.NewsRepository, storage services.StorageService) *NewsHandler {
	return &NewsHandler{repo: repo, storage: storage}
}

func (h *NewsHandler) CreateNews(w http.ResponseWriter, r *http.Request) {
	// Parse multi-part form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")

	if title == "" || content == "" {
		http.Error(w, "Title and content are required", http.StatusBadRequest)
		return
	}

	var n models.News
	n.Title = title
	n.Content = content

	// Tenta capturar o User ID do contexto (se autenticado)
	var currentUserID int64
	if userID := r.Context().Value(middleware.UserIDKey); userID != nil {
		if id, ok := userID.(int64); ok {
			currentUserID = id
			n.AuthorID = &currentUserID
		} else if id, ok := userID.(float64); ok {
			currentUserID = int64(id)
			n.AuthorID = &currentUserID
		}
	}

	// Handle optional multiple images upload
	var imageURLs []string
	files := r.MultipartForm.File["images"] // Use "images" para suportar multiplos uploads

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			continue // Pula o arquivo se houver erro ao abrir
		}

		ext := filepath.Ext(fileHeader.Filename)
		userIdStr := strconv.FormatInt(currentUserID, 10)
		filename := "news_images/" + userIdStr + "/" + uuid.New().String() + ext

		// Usar o service de storage injetado
		if h.storage != nil {
			imageURL, uploadErr := h.storage.UploadFile(file, filename, fileHeader.Header.Get("Content-Type"))
			if uploadErr == nil {
				imageURLs = append(imageURLs, imageURL)
			}
		}
		file.Close()
	}

	n.ImageURLs = imageURLs

	news, err := h.repo.CreateNews(r.Context(), &n)
	if err != nil {
		http.Error(w, "Failed to create news", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(news)
}

func (h *NewsHandler) ListNews(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	page, limit := parsePagination(r)

	newsList, err := h.repo.ListNews(r.Context(), search, page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list news")
		return
	}

	if newsList == nil {
		newsList = []*models.News{}
	}
	respondJSON(w, http.StatusOK, newsList)
}

func (h *NewsHandler) GetNews(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	n, err := h.repo.GetNews(r.Context(), id)
	if err != nil {
		http.Error(w, "News not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(n)
}

func (h *NewsHandler) UpdateNews(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")

	if title == "" || content == "" {
		http.Error(w, "Title and content are required", http.StatusBadRequest)
		return
	}

	var n models.News
	n.Title = title
	n.Content = content

	// Tenta capturar o User ID do contexto
	var currentUserID int64
	if userID := r.Context().Value(middleware.UserIDKey); userID != nil {
		if idVal, ok := userID.(int64); ok {
			currentUserID = idVal
		} else if idVal, ok := userID.(float64); ok {
			currentUserID = int64(idVal)
		}
	}

	// Handle optional multiple images upload
	var imageURLs []string
	files := r.MultipartForm.File["images"]

	if len(files) > 0 {
		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				continue
			}

			ext := filepath.Ext(fileHeader.Filename)
			userIdStr := strconv.FormatInt(currentUserID, 10)
			filename := "news_images/" + userIdStr + "/" + uuid.New().String() + ext

			if h.storage != nil {
				imageURL, uploadErr := h.storage.UploadFile(file, filename, fileHeader.Header.Get("Content-Type"))
				if uploadErr == nil {
					imageURLs = append(imageURLs, imageURL)
				}
			}
			file.Close()
		}
		n.ImageURLs = imageURLs
	} else {
		// Keep the existing image urls if no new images were uploaded
		existingNews, getErr := h.repo.GetNews(r.Context(), id)
		if getErr == nil {
			n.ImageURLs = existingNews.ImageURLs
		}
	}

	news, err := h.repo.UpdateNews(r.Context(), id, &n)
	if err != nil {
		http.Error(w, "Failed to update news", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(news)
}

func (h *NewsHandler) DeleteNews(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.repo.DeleteNews(r.Context(), id); err != nil {
		http.Error(w, "Failed to delete news", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
