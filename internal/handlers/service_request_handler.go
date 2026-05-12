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

type ServiceRequestHandler struct {
	srRepo         *repository.ServiceRequestRepository
	userRepo       *repository.UserRepository
	regionRepo     *repository.RegionRepository
	teamRepo       *repository.TeamRepository
	sysNotifRepo   *repository.SystemNotificationRepository
	pushTokenRepo  *repository.PushTokenRepository
	pushService    *services.ExpoPushService
	uploadService  *services.UploadService
	geoService     *services.GeocodingService
	ratingRepo     *repository.ServiceRatingRepository
	attendanceRepo *repository.ServiceAttendanceRepository
}

var statusLabels = map[string]string{
	"pending":     "Pendente",
	"in_progress": "Em andamento",
	"completed":   "Concluído",
	"cancelled":   "Cancelado",
}

func NewServiceRequestHandler(
	srRepo *repository.ServiceRequestRepository,
	userRepo *repository.UserRepository,
	regionRepo *repository.RegionRepository,
	teamRepo *repository.TeamRepository,
	sysNotifRepo *repository.SystemNotificationRepository,
	pushTokenRepo *repository.PushTokenRepository,
	pushService *services.ExpoPushService,
	uploadService *services.UploadService,
	geoService *services.GeocodingService,
	ratingRepo *repository.ServiceRatingRepository,
	attendanceRepo *repository.ServiceAttendanceRepository,
) *ServiceRequestHandler {
	return &ServiceRequestHandler{
		srRepo:         srRepo,
		userRepo:       userRepo,
		regionRepo:     regionRepo,
		teamRepo:       teamRepo,
		sysNotifRepo:   sysNotifRepo,
		pushTokenRepo:  pushTokenRepo,
		pushService:    pushService,
		uploadService:  uploadService,
		geoService:     geoService,
		ratingRepo:     ratingRepo,
		attendanceRepo: attendanceRepo,
	}
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
		if err := r.ParseMultipartForm(services.MaxTotalFilesSizeBytes); err != nil {
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
	}

	if req.ServiceID == nil || *req.ServiceID == 0 || req.ServiceTitle == "" {
		respondError(w, http.StatusBadRequest, "service_id and service_title are required")
		return
	}
	if len(req.RequestData) == 0 {
		req.RequestData = []byte("{}")
	}

	var regionID, teamID *int64
	var geoLat, geoLon *float64
	var geoAddress *string

	address := extractAddressFromRequestData(req.RequestData)
	if address != "" {
		geoResult, err := h.geoService.GeocodeAddress(address)
		if err == nil && geoResult.Found {
			geoLat = &geoResult.Latitude
			geoLon = &geoResult.Longitude
			geoAddress = &geoResult.DisplayName

			if geoResult.Bairro != "" {
				region, err := h.regionRepo.FindByNeighborhood(r.Context(), geoResult.Bairro)
				if err == nil {
					regionID = &region.ID
					team, err := h.teamRepo.GetTeamByRegion(r.Context(), region.ID)
					if err == nil {
						teamID = &team.ID
					}
				}
			}
		}
	}

	sr, err := h.srRepo.CreateServiceRequest(r.Context(), &userID, &req, regionID, teamID, geoLat, geoLon, geoAddress)
	if err != nil {
		if urls := services.ParseAttachmentURLs(req.Attachments); len(urls) > 0 {
			h.uploadService.RollbackFiles(urls)
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, sr)
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

	detail := models.ServiceRequestDetailResponse{
		ServiceRequest: sr,
	}

	if sr.UserID != nil {
		user, err := h.userRepo.GetUserByID(r.Context(), *sr.UserID)
		if err == nil {
			detail.CreatedBy = user
		}
		count, err := h.srRepo.CountServiceRequestsByUser(r.Context(), *sr.UserID)
		if err == nil {
			detail.UserRequests = count
		}
	}

	rating, _ := h.ratingRepo.GetByRequestID(r.Context(), id)
	detail.Rating = rating

	attendances, _ := h.attendanceRepo.ListByRequestID(r.Context(), id)
	detail.Attendances = attendances

	respondJSON(w, http.StatusOK, detail)
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

	existing, err := h.srRepo.GetServiceRequestByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "service request not found")
		return
	}

	sr, err := h.srRepo.UpdateServiceRequestStatus(r.Context(), id, req.Status)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.SaveServiceRequestStatusUpdatedNotification(existing.UserID, sr, req.Status)
	h.DispatchServiceRequestStatusUpdated(existing.UserID, sr, req.Status)

	respondJSON(w, http.StatusOK, sr)
}

// DELETE /service-requests/{id}
func (h *ServiceRequestHandler) DeleteServiceRequest(w http.ResponseWriter, r *http.Request) {
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

	if err := h.srRepo.DeleteServiceRequest(r.Context(), id); err != nil {
		respondError(w, http.StatusNotFound, "service request not found")
		return
	}

	if urls := services.ParseAttachmentURLs(sr.Attachments); len(urls) > 0 {
		h.uploadService.RollbackFiles(urls)
	}

	respondJSON(w, http.StatusOK, models.MessageResponse{Message: "service request deleted successfully"})
}

// extractAddressFromRequestData extrai o endereço do JSON de request_data
func extractAddressFromRequestData(requestData json.RawMessage) string {
	if len(requestData) == 0 {
		return ""
	}

	var data map[string]interface{}
	if err := json.Unmarshal(requestData, &data); err != nil {
		return ""
	}

	if addr, ok := data["endereco"].(string); ok && addr != "" {
		return addr
	}
	if addr, ok := data["address"].(string); ok && addr != "" {
		return addr
	}
	if addr, ok := data["logradouro"].(string); ok && addr != "" {
		return addr
	}
	if addr, ok := data["street"].(string); ok && addr != "" {
		return addr
	}

	return ""
}
