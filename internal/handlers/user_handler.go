package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/laiirton/solucoes-urbanas-api/internal/middleware"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
	"github.com/laiirton/solucoes-urbanas-api/internal/repository"
	"github.com/laiirton/solucoes-urbanas-api/internal/services"
	"golang.org/x/crypto/bcrypt"
)

var nonDigitsRegex = regexp.MustCompile("[^0-9]")

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
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	var startDatePtr, endDatePtr *string
	if startDate != "" { startDatePtr = &startDate }
	if endDate != "" { endDatePtr = &endDate }

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
		if !teamNull {
			teamFilter = currentUser.TeamID
			if teamFilter == nil {
				respondJSON(w, http.StatusOK, []*models.User{})
				return
			}
		}
	}

	users, err := h.userRepo.ListUsers(r.Context(), search, userType, teamFilter, page, limit, startDatePtr, endDatePtr)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list users")
		return
	}

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
	case "secretary":
		if req.Type == nil || *req.Type != "attendant" {
			respondError(w, http.StatusForbidden, "secretaries can only create attendants")
			return
		}

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

	if req.Password == "" {
		cleanCPF := nonDigitsRegex.ReplaceAllString(*req.CPF, "")
		cleanBirthDate := nonDigitsRegex.ReplaceAllString(*req.BirthDate, "")
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

	resp := h.buildUserDetail(r.Context(), id, user)
	respondJSON(w, http.StatusOK, resp)
}

func (h *UserHandler) buildUserDetail(ctx context.Context, userID int64, user *models.User) models.UserDetailResponse {
	var (
		total    int
		requests []*models.ServiceRequest
		summary  map[string]int
		wg       sync.WaitGroup
	)

	wg.Add(3)
	go func() {
		defer wg.Done()
		t, e := h.srRepo.CountServiceRequestsByUser(ctx, userID)
		if e == nil {
			total = t
		}
	}()
	go func() {
		defer wg.Done()
		r, e := h.srRepo.ListServiceRequestsByUser(ctx, userID, "", "", nil, 1, 10)
		if e == nil {
			requests = r
		} else {
			requests = []*models.ServiceRequest{}
		}
	}()
	go func() {
		defer wg.Done()
		s, e := h.srRepo.CountServiceRequestsByStatusByUser(ctx, userID)
		if e == nil {
			summary = s
		} else {
			summary = map[string]int{}
		}
	}()
	wg.Wait()

	return models.UserDetailResponse{
		User:           *user,
		TotalRequests:  total,
		Requests:       requests,
		RequestSummary: summary,
	}
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

	resp := h.buildUserDetail(r.Context(), userID, user)
	respondJSON(w, http.StatusOK, resp)
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

	oldUser, _ := h.userRepo.GetUserByID(r.Context(), id)

	user, err := h.userRepo.UpdateUser(r.Context(), id, &req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if h.teamRepo != nil && user.Type != nil && *user.Type == "secretary" {
		if oldUser != nil && oldUser.TeamID != nil {
			h.teamRepo.SyncTeamCategories(r.Context(), *oldUser.TeamID)
		}
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

	user, _ := h.userRepo.GetUserByID(r.Context(), id)

	if err := h.userRepo.DeleteUser(r.Context(), id); err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

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

	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if userID != id {
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

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
	if !allowedExts[ext] {
		respondError(w, http.StatusBadRequest, "Invalid file type. Allowed: jpg, jpeg, png")
		return
	}

	if fileHeader.Size > 10<<20 {
		respondError(w, http.StatusBadRequest, "File size exceeds 10MB limit")
		return
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" || !services.AllowedMIMETypes[contentType] {
		contentType = "image/jpeg"
	}

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

	updateReq := &models.UpdateUserRequest{
		ProfileImageURL: &imageURL,
	}
	_, updateErr := h.userRepo.UpdateUser(r.Context(), id, updateReq)
	if updateErr != nil {
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

	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if userID != id {
		user, err := h.userRepo.GetUserByID(r.Context(), userID)
		if err != nil || (user.Type == nil || *user.Type != "admin") {
			respondError(w, http.StatusForbidden, "forbidden")
			return
		}
	}

	user, err := h.userRepo.GetUserByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	if user.ProfileImageURL != nil && *user.ProfileImageURL != "" {
		h.storage.DeleteFile(*user.ProfileImageURL)

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
