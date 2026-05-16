package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/laiirton/solucoes-urbanas-api/internal/middleware"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
	"github.com/laiirton/solucoes-urbanas-api/internal/repository"
	"github.com/laiirton/solucoes-urbanas-api/internal/services"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	userRepo *repository.UserRepository
	srRepo   *repository.ServiceRequestRepository
	teamRepo *repository.TeamRepository
	storage  services.StorageService
}

func NewUserHandler(userRepo *repository.UserRepository, srRepo *repository.ServiceRequestRepository, teamRepo *repository.TeamRepository, storage services.StorageService) *UserHandler {
	return &UserHandler{userRepo: userRepo, srRepo: srRepo, teamRepo: teamRepo, storage: storage}
}

// GET /users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	userType := r.URL.Query().Get("type")
	teamNull := r.URL.Query().Get("team_null") == "true"
	page, limit := parsePagination(r)

	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	currentUser, err := h.userRepo.GetUserByID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	var teamFilter *int64
	if currentUser.Type != nil && *currentUser.Type == "secretary" {
		teamFilter = currentUser.TeamID
		if teamFilter == nil {
			respondJSON(w, http.StatusOK, []*models.User{})
			return
		}
	}

	users, err := h.userRepo.ListUsers(r.Context(), search, userType, teamFilter, page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	// Filter to show only users without a team (team_null=true)
	if teamNull && users != nil {
		filtered := []*models.User{}
		for _, u := range users {
			if u.TeamID == nil {
				filtered = append(filtered, u)
			}
		}
		users = filtered
	}
	if users == nil {
		users = []*models.User{}
	}

	// Filter out current user if it's a secretary
	if currentUser.Type != nil && *currentUser.Type == "secretary" {
		filtered := []*models.User{}
		for _, u := range users {
			if u.ID != userID {
				filtered = append(filtered, u)
			}
		}
		users = filtered
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

	// Validate authentication and authorization
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	currentUser, err := h.userRepo.GetUserByID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get current user")
		return
	}

	if currentUser.Type == nil {
		respondError(w, http.StatusForbidden, "forbidden")
		return
	}

	switch *currentUser.Type {
	case "admin":
		// Admin can create any type of user
	case "secretary":
		// Secretary can only create attendants
		if req.Type == nil || *req.Type != "attendant" {
			respondError(w, http.StatusForbidden, "secretaries can only create attendants")
			return
		}

		// Auto-assign team and work area from secretary
		if currentUser.TeamID == nil {
			respondError(w, http.StatusBadRequest, "secretary does not have a team assigned")
			return
		}
		req.TeamID = currentUser.TeamID
		req.WorkArea = currentUser.WorkArea
	default:
		respondError(w, http.StatusForbidden, "forbidden")
		return
	}

	// Generate automatic password if not provided: CPF + birth date (digits only)
	if req.Password == "" {
		nonDigits := regexp.MustCompile("[^0-9]")
		cleanCPF := nonDigits.ReplaceAllString(*req.CPF, "")
		cleanBirthDate := nonDigits.ReplaceAllString(*req.BirthDate, "")
		req.Password = cleanCPF + cleanBirthDate
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

	// Sync team categories if creating a secretary with a team
	if user.Type != nil && *user.Type == "secretary" && user.TeamID != nil && h.teamRepo != nil {
		h.teamRepo.SyncTeamCategories(r.Context(), *user.TeamID)
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

	// Secretário só pode ver usuários do próprio time
	currentUserID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if ok {
		currentUser, err := h.userRepo.GetUserByID(r.Context(), currentUserID)
		if err == nil && currentUser.Type != nil && *currentUser.Type == "secretary" {
			targetUser, err := h.userRepo.GetUserByID(r.Context(), id)
			if err != nil || targetUser.TeamID == nil || currentUser.TeamID == nil || *targetUser.TeamID != *currentUser.TeamID {
				respondError(w, http.StatusNotFound, "user not found")
				return
			}
		}
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

	requests, err := h.srRepo.ListServiceRequestsByUser(r.Context(), id, "", "", nil, 1, 10)
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

	requests, err := h.srRepo.ListServiceRequestsByUser(r.Context(), userID, "", "", nil, 1, 10)
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

	// Get old user data before update (for sync)
	oldUser, _ := h.userRepo.GetUserByID(r.Context(), id)

	user, err := h.userRepo.UpdateUser(r.Context(), id, &req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Sync team categories if user is a secretary and team_id may have changed
	if h.teamRepo != nil && user.Type != nil && *user.Type == "secretary" {
		// Sync old team
		if oldUser != nil && oldUser.TeamID != nil {
			h.teamRepo.SyncTeamCategories(r.Context(), *oldUser.TeamID)
		}
		// Sync new team
		if user.TeamID != nil {
			h.teamRepo.SyncTeamCategories(r.Context(), *user.TeamID)
		}
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

	// Get user before deleting to know team_id
	user, _ := h.userRepo.GetUserByID(r.Context(), id)

	if err := h.userRepo.DeleteUser(r.Context(), id); err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	// Sync team categories if deleted user was a secretary with a team
	if user != nil && user.Type != nil && *user.Type == "secretary" && user.TeamID != nil && h.teamRepo != nil {
		h.teamRepo.SyncTeamCategories(r.Context(), *user.TeamID)
	}

	respondJSON(w, http.StatusOK, models.MessageResponse{Message: "user deleted successfully"})
}

// POST /users/{id}/profile-image
func (h *UserHandler) UploadProfileImage(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	// Validate authentication
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Users can only upload their own profile image (unless admin)
	if userID != id {
		// Check if user is admin
		user, err := h.userRepo.GetUserByID(r.Context(), userID)
		if err != nil || (user.Type == nil || *user.Type != "admin") {
			respondError(w, http.StatusForbidden, "forbidden")
			return
		}
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "Unable to parse form")
		return
	}

	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Image is required")
		return
	}
	defer file.Close()

	// Validate file type (only images)
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
	if !allowedExts[ext] {
		respondError(w, http.StatusBadRequest, "Invalid file type. Allowed: jpg, jpeg, png")
		return
	}

	// Validate file size (max 10MB for profile images)
	if fileHeader.Size > 10<<20 {
		respondError(w, http.StatusBadRequest, "File size exceeds 10MB limit")
		return
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" || !services.AllowedMIMETypes[contentType] {
		contentType = "image/jpeg" // default
	}

	// Generate path: profile_images/{userID}/{uuid}.{ext}
	filename := fmt.Sprintf("profile_images/%d/%s%s", id, uuid.New().String(), ext)

	if h.storage == nil {
		respondError(w, http.StatusInternalServerError, "Storage service not configured")
		return
	}

	imageURL, uploadErr := h.storage.UploadFile(file, filename, contentType)
	if uploadErr != nil {
		respondError(w, http.StatusInternalServerError, "Failed to upload image")
		return
	}

	// Update user's profile_image_url in database
	updateReq := &models.UpdateUserRequest{
		ProfileImageURL: &imageURL,
	}
	_, updateErr := h.userRepo.UpdateUser(r.Context(), id, updateReq)
	if updateErr != nil {
		// Rollback: delete uploaded file
		h.storage.DeleteFile(imageURL)
		respondError(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"url": imageURL})
}

// DELETE /users/{id}/profile-image
func (h *UserHandler) DeleteProfileImage(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	// Validate authentication
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Users can only delete their own profile image (unless admin)
	if userID != id {
		user, err := h.userRepo.GetUserByID(r.Context(), userID)
		if err != nil || (user.Type == nil || *user.Type != "admin") {
			respondError(w, http.StatusForbidden, "forbidden")
			return
		}
	}

	// Get current user to fetch profile image URL
	user, err := h.userRepo.GetUserByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	if user.ProfileImageURL != nil && *user.ProfileImageURL != "" {
		// Delete from storage
		h.storage.DeleteFile(*user.ProfileImageURL)

		// Update database
		emptyURL := ""
		updateReq := &models.UpdateUserRequest{
			ProfileImageURL: &emptyURL,
		}
		_, updateErr := h.userRepo.UpdateUser(r.Context(), id, updateReq)
		if updateErr != nil {
			respondError(w, http.StatusInternalServerError, "Failed to remove profile image")
			return
		}
	}

	respondJSON(w, http.StatusOK, models.MessageResponse{Message: "Profile image removed successfully"})
}

// helpers
