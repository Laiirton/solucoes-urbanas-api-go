package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/laiirton/solucoes-urbanas-api/internal/models"
	"github.com/laiirton/solucoes-urbanas-api/internal/repository"
)

type TeamHandler struct {
	teamRepo *repository.TeamRepository
}

func NewTeamHandler(teamRepo *repository.TeamRepository) *TeamHandler {
	return &TeamHandler{teamRepo: teamRepo}
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

	if req.Name == "" || req.ServiceCategory == "" {
		respondError(w, http.StatusBadRequest, "name and service_category are required")
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
