package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/laiirton/solucoes-urbanas-api/internal/middleware"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
	"github.com/laiirton/solucoes-urbanas-api/internal/repository"
	"github.com/laiirton/solucoes-urbanas-api/internal/services"
)

type NewsHandler struct {
	repo          *repository.NewsRepository
	pushTokenRepo *repository.PushTokenRepository
	pushService   *services.ExpoPushService
	storage       services.StorageService
}

func NewNewsHandler(repo *repository.NewsRepository, pushTokenRepo *repository.PushTokenRepository, pushService *services.ExpoPushService, storage services.StorageService) *NewsHandler {
	return &NewsHandler{
		repo:          repo,
		pushTokenRepo: pushTokenRepo,
		pushService:   pushService,
		storage:       storage,
	}
}

func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = regexp.MustCompile(`[^a-z0-9\s-]`).ReplaceAllString(slug, "")
	slug = regexp.MustCompile(`\s+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if len(slug) > 100 {
		slug = slug[:100]
	}
	return slug
}

func (h *NewsHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
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

	var currentUserID int64
	if userID := r.Context().Value(middleware.UserIDKey); userID != nil {
		if id, ok := userID.(int64); ok {
			currentUserID = id
		} else if id, ok := userID.(float64); ok {
			currentUserID = int64(id)
		}
	}

	ext := filepath.Ext(fileHeader.Filename)
	userIdStr := strconv.FormatInt(currentUserID, 10)
	filename := "news_content/" + userIdStr + "/" + uuid.New().String() + ext

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

func (h *NewsHandler) CreateNews(w http.ResponseWriter, r *http.Request) {
	var n models.News
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if n.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	if n.Slug == "" {
		n.Slug = generateSlug(n.Title)
	}

	if n.Status == "" {
		n.Status = "draft"
	}

	if n.Status == "published" && n.PublishedAt == nil {
		now := time.Now()
		n.PublishedAt = &now
	}

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

	news, err := h.repo.CreateNews(r.Context(), &n)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create news: %v", err), http.StatusInternalServerError)
		return
	}

	if news.Status == "published" {
		h.dispatchNewsPublished(news.ID, news.Title, news.Summary)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(news)
}

func (h *NewsHandler) ListNews(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	status := r.URL.Query().Get("status")
	page, limit := parsePagination(r)

	newsList, err := h.repo.ListNews(r.Context(), search, status, page, limit)
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
	idOrSlug := chi.URLParam(r, "id")

	var n *models.News
	var err error

	if id, parseErr := strconv.ParseInt(idOrSlug, 10, 64); parseErr == nil {
		n, err = h.repo.GetNews(r.Context(), id)
	} else {
		n, err = h.repo.GetNewsBySlug(r.Context(), idOrSlug)
	}

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

	existing, err := h.repo.GetNews(r.Context(), id)
	if err != nil {
		http.Error(w, "News not found", http.StatusNotFound)
		return
	}

	var n models.UpdateNewsRequest
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !hasNewsUpdateFields(&n) {
		http.Error(w, "At least one field is required", http.StatusBadRequest)
		return
	}

	shouldNotify := n.Status != nil && *n.Status == "published" && existing.Status != "published"
	if shouldNotify && n.PublishedAt == nil {
		now := time.Now()
		n.PublishedAt = &now
	}

	news, err := h.repo.UpdateNews(r.Context(), id, &n)
	if err != nil {
		http.Error(w, "Failed to update news", http.StatusInternalServerError)
		return
	}

	if shouldNotify {
		h.dispatchNewsPublished(news.ID, news.Title, news.Summary)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(news)
}

func extractSupabaseURLs(content []byte, imageURLs []string) []string {
	var urls []string
	urls = append(urls, imageURLs...)

	re := regexp.MustCompile(`https?://[^\s"'\\]+/storage/v1/object/public/[^\s"'\\]+`)
	matches := re.FindAllString(string(content), -1)
	urls = append(urls, matches...)

	uniqueUrls := make(map[string]bool)
	var result []string
	for _, u := range urls {
		u = strings.ReplaceAll(u, "\\/", "/")
		if !uniqueUrls[u] {
			uniqueUrls[u] = true
			result = append(result, u)
		}
	}
	return result
}

func (h *NewsHandler) DeleteNews(w http.ResponseWriter, r *http.Request) {
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

	if err := h.repo.DeleteNews(r.Context(), id); err != nil {
		http.Error(w, "Failed to delete news", http.StatusInternalServerError)
		return
	}

	if h.storage != nil {
		urls := extractSupabaseURLs(n.Content, n.ImageURLs)
		for _, u := range urls {
			_ = h.storage.DeleteFile(u)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func hasNewsUpdateFields(req *models.UpdateNewsRequest) bool {
	return req.Title != nil ||
		req.Slug != nil ||
		req.Summary != nil ||
		req.Content != nil ||
		req.ImageURLs != nil ||
		req.Status != nil ||
		req.Category != nil ||
		req.Tags != nil ||
		req.PublishedAt != nil
}

func (h *NewsHandler) dispatchNewsPublished(newsID int64, title, summary string) {
	if h.pushTokenRepo == nil || h.pushService == nil {
		return
	}

	go func(id int64, t, s string) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		tokens, err := h.pushTokenRepo.ListTokens(ctx)
		if err != nil {
			log.Printf("warning: failed to list push tokens for news %d: %v", id, err)
			return
		}

		if len(tokens) == 0 {
			return
		}

		if err := h.pushService.SendNewsPublished(ctx, tokens, id, t, s); err != nil {
			log.Printf("warning: failed to send news notification for news %d: %v", id, err)
		}
	}(newsID, title, summary)
}
