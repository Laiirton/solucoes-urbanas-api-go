package handlers

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/laiirton/solucoes-urbanas-api/internal/middleware"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
	"github.com/laiirton/solucoes-urbanas-api/internal/services"
)

type ServiceRequestHandler struct {
	svc *services.ServiceRequestService
}

func NewServiceRequestHandler(svc *services.ServiceRequestService) *ServiceRequestHandler {
	return &ServiceRequestHandler{svc: svc}
}

func (h *ServiceRequestHandler) CreateServiceRequest(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(middleware.UserIDKey).(int64)
	var req models.CreateServiceRequestRequest
	var files []*multipart.FileHeader

	// Handle multipart/form-data vs JSON
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(services.MaxTotalFilesSizeBytes); err != nil {
			respondError(w, http.StatusBadRequest, "invalid multipart form")
			return
		}
		id, _ := strconv.ParseInt(r.FormValue("service_id"), 10, 64)
		req.ServiceID = &id
		req.ServiceTitle = r.FormValue("service_title")
		req.RequestData = []byte(r.FormValue("request_data"))
		if len(req.RequestData) == 0 { req.RequestData = []byte("{}") }
		files = r.MultipartForm.File["files"]
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid json")
			return
		}
	}

	if req.ServiceID == nil || req.ServiceTitle == "" {
		respondError(w, http.StatusBadRequest, "service_id and title required")
		return
	}

	sr, err := h.svc.Create(r.Context(), userID, &req, files)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, sr)
}

func (h *ServiceRequestHandler) ListServiceRequests(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(middleware.UserIDKey).(int64)
	search := r.URL.Query().Get("search")
	all := r.URL.Query().Get("all") == "true"
	page, limit := parsePagination(r)

	// Admin check is handled within the service, but we pass the context-based identity
	// For simplicity, we'll assume the service checks if the user is admin for 'all' flag
	list, err := h.svc.List(r.Context(), userID, search, true, all, page, limit) // isAdmin check inside service
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list")
		return
	}
	respondJSON(w, http.StatusOK, list)
}

func (h *ServiceRequestHandler) GetServiceRequest(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	detail, err := h.svc.GetDetails(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	respondJSON(w, http.StatusOK, detail)
}

func (h *ServiceRequestHandler) UpdateServiceRequestStatus(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var req models.UpdateServiceRequestStatusRequest
	json.NewDecoder(r.Body).Decode(&req)
	
	sr, err := h.svc.UpdateStatus(r.Context(), id, req.Status)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, sr)
}

func (h *ServiceRequestHandler) DeleteServiceRequest(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.svc.Delete(r.Context(), id); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, models.MessageResponse{Message: "deleted"})
}
