package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/laiirton/solucoes-urbanas-api/internal/middleware"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
	"github.com/laiirton/solucoes-urbanas-api/internal/repository"
)

type NotificationHandler struct {
	tokenRepo *repository.PushTokenRepository
}

func NewNotificationHandler(tokenRepo *repository.PushTokenRepository) *NotificationHandler {
	return &NotificationHandler{tokenRepo: tokenRepo}
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
