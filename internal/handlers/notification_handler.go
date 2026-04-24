package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/laiirton/solucoes-urbanas-api/internal/middleware"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
	"github.com/laiirton/solucoes-urbanas-api/internal/repository"
)

type NotificationHandler struct {
	tokenRepo    *repository.PushTokenRepository
	sysNotifRepo *repository.SystemNotificationRepository
}

func NewNotificationHandler(tokenRepo *repository.PushTokenRepository, sysNotifRepo *repository.SystemNotificationRepository) *NotificationHandler {
	return &NotificationHandler{
		tokenRepo:    tokenRepo,
		sysNotifRepo: sysNotifRepo,
	}
}

func (h *NotificationHandler) RegisterPushToken(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.RegisterPushTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if h.tokenRepo == nil {
		respondError(w, http.StatusInternalServerError, "notification service not configured")
		return
	}

	if err := h.tokenRepo.UpsertPushToken(r.Context(), userID, req.Token); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, models.MessageResponse{Message: "push token saved successfully"})
}

func (h *NotificationHandler) CreateSystemNotification(w http.ResponseWriter, r *http.Request) {
	if h.sysNotifRepo == nil {
		respondError(w, http.StatusInternalServerError, "system notification service not configured")
		return
	}

	var req models.CreateSystemNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	n := &models.SystemNotification{
		UserID: req.UserID,
		Title:  req.Title,
		Body:   req.Body,
		Type:   req.Type,
		Data:   req.Data,
	}

	created, err := h.sysNotifRepo.Create(r.Context(), n)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, created)
}

func (h *NotificationHandler) ListSystemNotifications(w http.ResponseWriter, r *http.Request) {
	if h.sysNotifRepo == nil {
		respondError(w, http.StatusInternalServerError, "system notification service not configured")
		return
	}

	userIDVal := r.Context().Value(middleware.UserIDKey)
	var currentUserID int64
	if id, ok := userIDVal.(int64); ok {
		currentUserID = id
	} else if id, ok := userIDVal.(float64); ok {
		currentUserID = int64(id)
	}

	queryUserID := r.URL.Query().Get("user_id")
	var filterUserID *int64
	if queryUserID != "" {
		if id, err := strconv.ParseInt(queryUserID, 10, 64); err == nil {
			filterUserID = &id
		}
	}

	// If no explicit user_id filter is provided, restrict to current user + broadcast
	if filterUserID == nil && currentUserID > 0 {
		filterUserID = &currentUserID
	}

	notificationType := r.URL.Query().Get("type")
	unreadOnly := r.URL.Query().Get("unread_only") == "true"
	page, limit := parsePagination(r)

	notifications, err := h.sysNotifRepo.List(r.Context(), filterUserID, notificationType, unreadOnly, page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if notifications == nil {
		notifications = []*models.SystemNotification{}
	}
	respondJSON(w, http.StatusOK, notifications)
}

func (h *NotificationHandler) GetSystemNotification(w http.ResponseWriter, r *http.Request) {
	if h.sysNotifRepo == nil {
		respondError(w, http.StatusInternalServerError, "system notification service not configured")
		return
	}

	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	n, err := h.sysNotifRepo.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, n)
}

func (h *NotificationHandler) UpdateSystemNotification(w http.ResponseWriter, r *http.Request) {
	if h.sysNotifRepo == nil {
		respondError(w, http.StatusInternalServerError, "system notification service not configured")
		return
	}

	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req models.UpdateSystemNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if !hasSystemNotificationUpdateFields(&req) {
		respondError(w, http.StatusBadRequest, "at least one field is required")
		return
	}

	n, err := h.sysNotifRepo.Update(r.Context(), id, &req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, n)
}

func (h *NotificationHandler) MarkSystemNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	if h.sysNotifRepo == nil {
		respondError(w, http.StatusInternalServerError, "system notification service not configured")
		return
	}

	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	n, err := h.sysNotifRepo.MarkAsRead(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, n)
}

func (h *NotificationHandler) DeleteSystemNotification(w http.ResponseWriter, r *http.Request) {
	if h.sysNotifRepo == nil {
		respondError(w, http.StatusInternalServerError, "system notification service not configured")
		return
	}

	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.sysNotifRepo.Delete(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func hasSystemNotificationUpdateFields(req *models.UpdateSystemNotificationRequest) bool {
	return req.Title != nil ||
		req.Body != nil ||
		req.Type != nil ||
		req.Data != nil ||
		req.ReadAt != nil
}
