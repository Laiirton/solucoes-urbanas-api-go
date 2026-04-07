package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/laiirton/solucoes-urbanas-api/internal/middleware"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
	"github.com/laiirton/solucoes-urbanas-api/internal/repository"
	"github.com/laiirton/solucoes-urbanas-api/internal/services"
)

type ServiceRequestHandler struct {
	srRepo         *repository.ServiceRequestRepository
	storageService services.StorageService
}

func NewServiceRequestHandler(srRepo *repository.ServiceRequestRepository, storageService services.StorageService) *ServiceRequestHandler {
	return &ServiceRequestHandler{srRepo: srRepo, storageService: storageService}
}

// POST /service-requests
func (h *ServiceRequestHandler) CreateServiceRequest(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.CreateServiceRequestRequest
	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB limit
			respondError(w, http.StatusBadRequest, "failed to parse multipart form")
			return
		}

		serviceIDStr := r.FormValue("service_id")
		serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid service_id")
			return
		}
		req.ServiceID = &serviceID
		req.ServiceTitle = r.FormValue("service_title")

		requestData := r.FormValue("request_data")
		if requestData != "" {
			req.RequestData = []byte(requestData)
		} else {
			req.RequestData = []byte("{}")
		}

		// Handle file uploads
		files := r.MultipartForm.File["files"]
		var attachmentURLs []string

		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				respondError(w, http.StatusInternalServerError, "failed to open uploaded file")
				return
			}
			defer file.Close()

			// Create a folder path using the userID: "userID/timestamp_filename"
			filename := fmt.Sprintf("%d/%d_%s", userID, time.Now().UnixNano(), fileHeader.Filename)

			fileContentType := fileHeader.Header.Get("Content-Type")
			if fileContentType == "" {
				fileContentType = "application/octet-stream"
			}

			publicURL, err := h.storageService.UploadFile(file, filename, fileContentType)
			if err != nil {
				respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to upload file %s: %v", fileHeader.Filename, err))
				return
			}
			attachmentURLs = append(attachmentURLs, publicURL)
		}

		if len(attachmentURLs) > 0 {
			urlsJSON, _ := json.Marshal(attachmentURLs)
			req.Attachments = urlsJSON
		}
	} else {
		// Fallback to standard JSON
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
	}

	if req.ServiceID == nil || *req.ServiceID == 0 || req.ServiceTitle == "" {
		respondError(w, http.StatusBadRequest, "service_id and service_title are required")
		return
	}
	if len(req.RequestData) == 0 {
		req.RequestData = []byte("{}")
	}

	sr, err := h.srRepo.CreateServiceRequest(r.Context(), &userID, &req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, sr)
}

// GET /service-requests
func (h *ServiceRequestHandler) ListServiceRequests(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// By default, list only the authenticated user's requests
	// Pass ?all=true for admins to see all (simple check — no role system yet)
	search := r.URL.Query().Get("search")
	page, limit := parsePagination(r)

	var list []*models.ServiceRequest
	var err error
	if r.URL.Query().Get("all") == "true" {
		list, err = h.srRepo.ListServiceRequests(r.Context(), search, page, limit)
	} else {
		list, err = h.srRepo.ListServiceRequestsByUser(r.Context(), userID, search, page, limit)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list service requests")
		return
	}
	respondJSON(w, http.StatusOK, list)
}

// GET /service-requests/{id}
func (h *ServiceRequestHandler) GetServiceRequest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid service request id")
		return
	}
	sr, err := h.srRepo.GetServiceRequestByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "service request not found")
		return
	}
	respondJSON(w, http.StatusOK, sr)
}

// PATCH /service-requests/{id}/status
func (h *ServiceRequestHandler) UpdateServiceRequestStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid service request id")
		return
	}

	var req models.UpdateServiceRequestStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Status == "" {
		respondError(w, http.StatusBadRequest, "status is required")
		return
	}

	sr, err := h.srRepo.UpdateServiceRequestStatus(r.Context(), id, req.Status)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, sr)
}

// DELETE /service-requests/{id}
func (h *ServiceRequestHandler) DeleteServiceRequest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid service request id")
		return
	}
	if err := h.srRepo.DeleteServiceRequest(r.Context(), id); err != nil {
		respondError(w, http.StatusNotFound, "service request not found")
		return
	}
	respondJSON(w, http.StatusOK, models.MessageResponse{Message: "service request deleted successfully"})
}
