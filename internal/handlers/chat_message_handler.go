package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/laiirton/solucoes-urbanas-api/internal/middleware"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
	"github.com/laiirton/solucoes-urbanas-api/internal/repository"
	"github.com/laiirton/solucoes-urbanas-api/internal/services"
)

type ChatMessageHandler struct {
	repo          *repository.ChatMessageRepository
	srRepo        *repository.ServiceRequestRepository
	userRepo      *repository.UserRepository
	uploadService *services.UploadService
	notifHandler  *ServiceRequestHandler
}

func NewChatMessageHandler(repo *repository.ChatMessageRepository, srRepo *repository.ServiceRequestRepository, userRepo *repository.UserRepository, uploadService *services.UploadService, notifHandler *ServiceRequestHandler) *ChatMessageHandler {
	return &ChatMessageHandler{
		repo:          repo,
		srRepo:        srRepo,
		userRepo:      userRepo,
		uploadService: uploadService,
		notifHandler:  notifHandler,
	}
}

// POST /service-requests/{id}/chat/messages
func (h *ChatMessageHandler) CreateChatMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	requestID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid service request id")
		return
	}

	var req models.CreateChatMessageRequest
	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(services.MaxTotalFilesSizeBytes); err != nil {
			respondError(w, http.StatusBadRequest, "failed to parse multipart form")
			return
		}

		req.ServiceRequestID = requestID
		req.Content = r.FormValue("content")

		files := r.MultipartForm.File["files"]
		attachmentURLs, err := h.uploadService.UploadChatFiles(userID, files)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}

		if len(attachmentURLs) > 0 {
			urlsJSON, _ := json.Marshal(attachmentURLs)
			req.Attachments = urlsJSON
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		req.ServiceRequestID = requestID
	}

	if req.Content == "" {
		respondError(w, http.StatusBadRequest, "content is required")
		return
	}

	// Permission check: only admin/secretary/attendant can send messages
	if !CanManageRequest(r.Context(), h.userRepo, h.srRepo, userID, requestID) {
		respondError(w, http.StatusNotFound, "service request not found")
		return
	}

	// Get sender name (denormalized)
	senderName := "Operador"
	user, err := h.userRepo.GetUserByID(r.Context(), userID)
	if err == nil && user != nil && user.FullName != nil {
		senderName = *user.FullName
	}

	// Verify request exists and get it for notification
	sr, err := h.srRepo.GetServiceRequestByID(r.Context(), requestID)
	if err != nil {
		respondError(w, http.StatusNotFound, "service request not found")
		return
	}

	msg, err := h.repo.Create(r.Context(), userID, senderName, &req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create chat message: "+err.Error())
		return
	}

	// Fire-and-forget notifications to the citizen
	h.notifHandler.SaveChatMessageNotification(sr.UserID, sr, senderName, req.Content)
	h.notifHandler.DispatchChatMessageNotification(sr.UserID, sr, senderName, req.Content)

	respondJSON(w, http.StatusCreated, msg)
}

// GET /service-requests/{id}/chat/messages
func (h *ChatMessageHandler) ListChatMessages(w http.ResponseWriter, r *http.Request) {
	requestID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid service request id")
		return
	}

	// Permission: user must own the request OR be able to manage it
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if ok {
		if !CanViewRequest(r.Context(), h.userRepo, h.srRepo, userID, requestID) {
			respondError(w, http.StatusNotFound, "service request not found")
			return
		}
	}

	list, err := h.repo.ListByRequestID(r.Context(), requestID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list chat messages")
		return
	}

	respondJSON(w, http.StatusOK, list)
}
