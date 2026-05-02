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

type ServiceAttendanceHandler struct {
	repo          *repository.ServiceAttendanceRepository
	srRepo        *repository.ServiceRequestRepository
	uploadService *services.UploadService
	notifHandler  *ServiceRequestHandler // Reuse notification logic
}

func NewServiceAttendanceHandler(repo *repository.ServiceAttendanceRepository, srRepo *repository.ServiceRequestRepository, uploadService *services.UploadService, notifHandler *ServiceRequestHandler) *ServiceAttendanceHandler {
	return &ServiceAttendanceHandler{
		repo:          repo,
		srRepo:        srRepo,
		uploadService: uploadService,
		notifHandler:  notifHandler,
	}
}

// POST /service-requests/{id}/attendances
func (h *ServiceAttendanceHandler) CreateAttendance(w http.ResponseWriter, r *http.Request) {
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

	var req models.CreateServiceAttendanceRequest
	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(services.MaxTotalFilesSizeBytes); err != nil {
			respondError(w, http.StatusBadRequest, "failed to parse multipart form")
			return
		}

		req.ServiceRequestID = requestID
		req.Notes = r.FormValue("notes")
		req.NewStatus = r.FormValue("new_status")

		files := r.MultipartForm.File["files"]
		attachmentURLs, err := h.uploadService.UploadServiceRequestFiles(userID, files)
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

	if req.Notes == "" {
		respondError(w, http.StatusBadRequest, "notes are required")
		return
	}

	// Verify if request exists
	sr, err := h.srRepo.GetServiceRequestByID(r.Context(), requestID)
	if err != nil {
		respondError(w, http.StatusNotFound, "service request not found")
		return
	}

	attendance, err := h.repo.Create(r.Context(), userID, &req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create attendance: "+err.Error())
		return
	}

	// If status was updated, send notifications
	if req.NewStatus != "" {
		updatedSR, _ := h.srRepo.GetServiceRequestByID(r.Context(), requestID)
		if updatedSR != nil {
			h.notifHandler.SaveServiceRequestStatusUpdatedNotification(sr.UserID, updatedSR, req.NewStatus)
			h.notifHandler.DispatchServiceRequestStatusUpdated(sr.UserID, updatedSR, req.NewStatus)
		}
	}

	respondJSON(w, http.StatusCreated, attendance)
}

// GET /service-requests/{id}/attendances
func (h *ServiceAttendanceHandler) ListAttendances(w http.ResponseWriter, r *http.Request) {
	requestID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid service request id")
		return
	}

	list, err := h.repo.ListByRequestID(r.Context(), requestID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list attendances")
		return
	}

	respondJSON(w, http.StatusOK, list)
}
