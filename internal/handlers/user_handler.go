package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/laiirton/solucoes-urbanas-api/internal/middleware"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
	"github.com/laiirton/solucoes-urbanas-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	userRepo *repository.UserRepository
	srRepo   *repository.ServiceRequestRepository
}

func NewUserHandler(userRepo *repository.UserRepository, srRepo *repository.ServiceRequestRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo, srRepo: srRepo}
}

// GET /users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	userType := r.URL.Query().Get("type")
	page, limit := parsePagination(r)

	users, err := h.userRepo.ListUsers(r.Context(), search, userType, page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list users")
		return
	}
	if users == nil {
		users = []*models.User{}
	}
	respondJSON(w, http.StatusOK, users)
}

// POST /users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user, err := h.userRepo.CreateUser(r.Context(), &req, string(hashedPassword))
	if err != nil {
		respondError(w, http.StatusConflict, "could not create user: "+err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, user)
}

// GET /users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	user, err := h.userRepo.GetUserByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	total, err := h.srRepo.CountServiceRequestsByUser(r.Context(), id)
	if err != nil {
		total = 0
	}

	requests, err := h.srRepo.ListServiceRequestsByUser(r.Context(), id, "", "", 1, 10)
	if err != nil {
		requests = []*models.ServiceRequest{}
	}

	summary, err := h.srRepo.CountServiceRequestsByStatusByUser(r.Context(), id)
	if err != nil {
		summary = map[string]int{}
	}

	respondJSON(w, http.StatusOK, models.UserDetailResponse{
		User:           *user,
		TotalRequests:  total,
		Requests:       requests,
		RequestSummary: summary,
	})
}

// GET /users/me
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.userRepo.GetUserByID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	total, err := h.srRepo.CountServiceRequestsByUser(r.Context(), userID)
	if err != nil {
		total = 0
	}

	requests, err := h.srRepo.ListServiceRequestsByUser(r.Context(), userID, "", "", 1, 10)
	if err != nil {
		requests = []*models.ServiceRequest{}
	}

	summary, err := h.srRepo.CountServiceRequestsByStatusByUser(r.Context(), userID)
	if err != nil {
		summary = map[string]int{}
	}

	respondJSON(w, http.StatusOK, models.UserDetailResponse{
		User:           *user,
		TotalRequests:  total,
		Requests:       requests,
		RequestSummary: summary,
	})
}

// PUT /users/{id}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.userRepo.UpdateUser(r.Context(), id, &req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, user)
}

// DELETE /users/{id}
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := h.userRepo.DeleteUser(r.Context(), id); err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	respondJSON(w, http.StatusOK, models.MessageResponse{Message: "user deleted successfully"})
}

// helpers
