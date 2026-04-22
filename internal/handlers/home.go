package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/laiirton/solucoes-urbanas-api/internal/middleware"
	"github.com/laiirton/solucoes-urbanas-api/internal/repository"
)

type HomeHandler struct {
	srRepo   *repository.ServiceRequestRepository
	userRepo *repository.UserRepository
}

func NewHomeHandler(srRepo *repository.ServiceRequestRepository, userRepo *repository.UserRepository) *HomeHandler {
	return &HomeHandler{
		srRepo:   srRepo,
		userRepo: userRepo,
	}
}

func (h *HomeHandler) Index(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userRepo.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	isAdmin := user.Type != nil && *user.Type == "admin"
	var categoryFilter string
	if isAdmin && user.Team != nil {
		categoryFilter = user.Team.ServiceCategory
	}

	resp, err := h.srRepo.GetHomeStats(r.Context(), isAdmin, userID, categoryFilter)
	if err != nil {
		http.Error(w, "Error computing home stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
