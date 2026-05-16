package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/laiirton/solucoes-urbanas-api/internal/middleware"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
	"github.com/laiirton/solucoes-urbanas-api/internal/repository"
)

type TeamHandler struct {
	teamRepo *repository.TeamRepository
	userRepo *repository.UserRepository
	srRepo   *repository.ServiceRequestRepository
}

func NewTeamHandler(teamRepo *repository.TeamRepository, userRepo *repository.UserRepository, srRepo *repository.ServiceRequestRepository) *TeamHandler {
	return &TeamHandler{
		teamRepo: teamRepo,
		userRepo: userRepo,
		srRepo:   srRepo,
	}
}

// GET /teams
func (h *TeamHandler) ListTeams(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	page, limit := parsePagination(r)

	teams, err := h.teamRepo.ListTeams(r.Context(), search, page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list teams")
		return
	}

	respondJSON(w, http.StatusOK, teams)
}

// POST /teams
func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.RegionID == 0 {
		respondError(w, http.StatusBadRequest, "name and region_id are required")
		return
	}

	team, err := h.teamRepo.CreateTeam(r.Context(), &req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create team: "+err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, team)
}

// GET /teams/{id}
func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid team id")
		return
	}

	team, err := h.teamRepo.GetTeamByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "team not found")
		return
	}

	respondJSON(w, http.StatusOK, team)
}

// PUT /teams/{id}
func (h *TeamHandler) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid team id")
		return
	}

	var req models.UpdateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	team, err := h.teamRepo.UpdateTeam(r.Context(), id, &req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, team)
}

// DELETE /teams/{id}
func (h *TeamHandler) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid team id")
		return
	}

	if err := h.teamRepo.DeleteTeam(r.Context(), id); err != nil {
		respondError(w, http.StatusNotFound, "team not found or cannot be deleted")
		return
	}

	respondJSON(w, http.StatusOK, models.MessageResponse{Message: "team deleted successfully"})
}

// GET /my-team — retorna a equipe do usuário autenticado (secretary ou attendant)
func (h *TeamHandler) GetMyTeam(w http.ResponseWriter, r *http.Request) {
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

	if user.Type == nil || (*user.Type != "secretary" && *user.Type != "attendant") {
		respondError(w, http.StatusForbidden, "only secretary and attendant users have a team")
		return
	}

	if user.TeamID == nil {
		respondError(w, http.StatusNotFound, "user has no team assigned")
		return
	}

	members, err := h.teamRepo.ListMembers(r.Context(), *user.TeamID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list team members")
		return
	}

	var secretary *models.User
	attendants := []*models.User{}

	for _, m := range members {
		if m.ID == userID {
			continue
		}
		if m.Type != nil && *m.Type == "secretary" {
			secretary = m
		} else {
			attendants = append(attendants, m)
		}
	}

	resp := models.MyTeamResponse{
		Team:       *user.Team,
		Secretary:  secretary,
		Attendants: attendants,
	}

	respondJSON(w, http.StatusOK, resp)
}

// GET /teams/{id}/members
func (h *TeamHandler) ListTeamMembers(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid team id")
		return
	}

	// Check authorization: admin, secretary, or team attendant
	currentUserID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if ok {
		user, userErr := h.userRepo.GetUserByID(r.Context(), currentUserID)
		if userErr == nil && user.Type != nil && *user.Type == "attendant" {
			if user.TeamID == nil || *user.TeamID != id {
				respondError(w, http.StatusForbidden, "only admins or the team secretary can view members")
				return
			}
		} else if !CanManageTeam(r.Context(), h.userRepo, currentUserID, id) {
			respondError(w, http.StatusForbidden, "only admins or the team secretary can view members")
			return
		}
	} else {
		respondError(w, http.StatusForbidden, "only admins or the team secretary can view members")
		return
	}

	members, err := h.teamRepo.ListMembers(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list team members")
		return
	}

	respondJSON(w, http.StatusOK, members)
}

// POST /teams/{id}/members
func (h *TeamHandler) AddTeamMember(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid team id")
		return
	}

	currentUserID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok || !CanManageTeam(r.Context(), h.userRepo, currentUserID, id) {
		respondError(w, http.StatusForbidden, "only admins or the team secretary can manage members")
		return
	}

	var req struct {
		UserID int64 `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.UserID == 0 {
		respondError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	if err := h.teamRepo.AddMember(r.Context(), id, req.UserID); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, models.MessageResponse{Message: "member added successfully"})
}

// DELETE /teams/{id}/members/{userId}
func (h *TeamHandler) RemoveTeamMember(w http.ResponseWriter, r *http.Request) {
	teamID, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid team id")
		return
	}

	currentUserID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok || !CanManageTeam(r.Context(), h.userRepo, currentUserID, teamID) {
		respondError(w, http.StatusForbidden, "only admins or the team secretary can manage members")
		return
	}

	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := h.teamRepo.RemoveMember(r.Context(), teamID, userID); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, models.MessageResponse{Message: "member removed successfully"})
}

// GET /teams/{id}/stats
func (h *TeamHandler) GetTeamDashboard(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid team id")
		return
	}

	// Check authorization: admin, secretary, or team attendant
	currentUserID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if ok {
		user, userErr := h.userRepo.GetUserByID(r.Context(), currentUserID)
		if userErr == nil && user.Type != nil && *user.Type == "attendant" {
			if user.TeamID == nil || *user.TeamID != id {
				respondError(w, http.StatusForbidden, "Você não tem permissão para acessar este recurso")
				return
			}
		} else if !CanManageTeam(r.Context(), h.userRepo, currentUserID, id) {
			respondError(w, http.StatusForbidden, "Você não tem permissão para acessar este recurso")
			return
		}
	} else {
		respondError(w, http.StatusForbidden, "Você não tem permissão para acessar este recurso")
		return
	}

	stats, err := h.teamRepo.GetTeamStats(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "team not found")
		return
	}

	respondJSON(w, http.StatusOK, stats)
}
