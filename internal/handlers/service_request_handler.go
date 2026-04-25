package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	srRepo        *repository.ServiceRequestRepository
	userRepo      *repository.UserRepository
	sysNotifRepo  *repository.SystemNotificationRepository
	pushTokenRepo *repository.PushTokenRepository
	pushService   *services.ExpoPushService
	uploadService *services.UploadService
	geoService    *services.GeocodingService
}

func NewServiceRequestHandler(srRepo *repository.ServiceRequestRepository, userRepo *repository.UserRepository, sysNotifRepo *repository.SystemNotificationRepository, pushTokenRepo *repository.PushTokenRepository, pushService *services.ExpoPushService, uploadService *services.UploadService, geoService *services.GeocodingService) *ServiceRequestHandler {
	return &ServiceRequestHandler{
		srRepo:        srRepo,
		userRepo:      userRepo,
		sysNotifRepo:  sysNotifRepo,
		pushTokenRepo: pushTokenRepo,
		pushService:   pushService,
		uploadService: uploadService,
		geoService:    geoService,
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

		// Handle file uploads using the UploadService
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
		// Rollback uploaded files if DB insert fails
		if urls := services.ParseAttachmentURLs(req.Attachments); len(urls) > 0 {
			h.uploadService.RollbackFiles(urls)
		}
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

	search := r.URL.Query().Get("search")
	page, limit := parsePagination(r)

	var categoryFilter string
	user, err := h.userRepo.GetUserByID(r.Context(), userID)
	if err == nil && user.Type != nil && *user.Type == "admin" && user.Team != nil {
		categoryFilter = user.Team.ServiceCategory
	}

	var list []*models.ServiceRequest
	if r.URL.Query().Get("all") == "true" {
		list, err = h.srRepo.ListServiceRequests(r.Context(), search, categoryFilter, page, limit)
	} else {
		list, err = h.srRepo.ListServiceRequestsByUser(r.Context(), userID, search, categoryFilter, page, limit)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list service requests")
		return
	}
	respondJSON(w, http.StatusOK, list)
}

// GET /service-requests/{id}/geocode - Geocodifica o endereço do service request
func (h *ServiceRequestHandler) GeocodeServiceRequest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid service request id")
		return
	}

	// Buscar o service request
	sr, err := h.srRepo.GetServiceRequestByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "service request not found")
		return
	}

	// Extrair endereço do request_data
	address := extractAddressFromRequestData(sr.RequestData)
	if address == "" {
		respondError(w, http.StatusBadRequest, "no address found in service request data")
		return
	}

	// Geocodificar o endereço
	geoResult, err := h.geoService.GeocodeAddress(address)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to geocode address")
		return
	}

	response := map[string]interface{}{
		"service_request_id": id,
		"address":            address,
		"latitude":           geoResult.Latitude,
		"longitude":          geoResult.Longitude,
		"display_name":       geoResult.DisplayName,
		"found":              geoResult.Found,
	}

	respondJSON(w, http.StatusOK, response)
}

// GET /service-requests/geocode-all - Retorna todos os service requests com coordenadas para o mapa
func (h *ServiceRequestHandler) GeocodeAllServiceRequests(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	search := r.URL.Query().Get("search")
	page, limit := parsePagination(r)

	var categoryFilter string
	user, err := h.userRepo.GetUserByID(r.Context(), userID)
	if err == nil && user.Type != nil && *user.Type == "admin" && user.Team != nil {
		categoryFilter = user.Team.ServiceCategory
	}

	var list []*models.ServiceRequest
	if r.URL.Query().Get("all") == "true" {
		list, err = h.srRepo.ListServiceRequests(r.Context(), search, categoryFilter, page, limit)
	} else {
		list, err = h.srRepo.ListServiceRequestsByUser(r.Context(), userID, search, categoryFilter, page, limit)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list service requests")
		return
	}

	// Preparar resposta com coordenadas
	type MapLocation struct {
		ID           int64   `json:"id"`
		Address      string  `json:"address"`
		Latitude     float64 `json:"latitude"`
		Longitude    float64 `json:"longitude"`
		ServiceTitle string  `json:"service_title"`
		Status       string  `json:"status"`
		Icon         string  `json:"icon"`
		Found        bool    `json:"found"`
	}

	var locations []MapLocation
	for _, sr := range list {
		address := extractAddressFromRequestData(sr.RequestData)
		geoResult, _ := h.geoService.GeocodeAddress(address)

		// Incluir apenas se o endereço foi encontrado
		if geoResult.Found {
			icon := ""
			if sr.ServiceID != nil {
				icon = models.GetServiceIcon(*sr.ServiceID)
			}
			locations = append(locations, MapLocation{
				ID:           sr.ID,
				Address:      address,
				Latitude:     geoResult.Latitude,
				Longitude:    geoResult.Longitude,
				ServiceTitle: sr.ServiceTitle,
				Status:       sr.Status,
				Icon:         icon,
				Found:        geoResult.Found,
			})
		}
	}

	respondJSON(w, http.StatusOK, locations)
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

	// Tentar campos comuns de endereço
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

	// Fetch existing to get user_id before updating
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

	h.saveServiceRequestStatusUpdatedNotification(existing.UserID, sr, req.Status)
	h.dispatchServiceRequestStatusUpdated(existing.UserID, sr, req.Status)

	respondJSON(w, http.StatusOK, sr)
}

// DELETE /service-requests/{id}
func (h *ServiceRequestHandler) DeleteServiceRequest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid service request id")
		return
	}

	// Fetch the service request first to get attachments for cleanup
	sr, err := h.srRepo.GetServiceRequestByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "service request not found")
		return
	}

	if err := h.srRepo.DeleteServiceRequest(r.Context(), id); err != nil {
		respondError(w, http.StatusNotFound, "service request not found")
		return
	}

	// Delete attachment files from storage after successful DB deletion
	if urls := services.ParseAttachmentURLs(sr.Attachments); len(urls) > 0 {
		h.uploadService.RollbackFiles(urls)
	}

	respondJSON(w, http.StatusOK, models.MessageResponse{Message: "service request deleted successfully"})
}

func (h *ServiceRequestHandler) saveServiceRequestStatusUpdatedNotification(userID *int64, sr *models.ServiceRequest, newStatus string) {
	if h.sysNotifRepo == nil || userID == nil {
		return
	}

	go func(uid int64, req *models.ServiceRequest, status string) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		protocol := ""
		if req.ProtocolNumber != nil {
			protocol = *req.ProtocolNumber
		}

		data, _ := json.Marshal(map[string]interface{}{
			"service_request_id": req.ID,
			"protocol_number":    protocol,
			"status":             status,
			"screen":             fmt.Sprintf("/(service-requests)/%d", req.ID),
		})

		_, err := h.sysNotifRepo.Create(ctx, &models.SystemNotification{
			UserID: &uid,
			Title:  "Status do chamado atualizado",
			Body:   fmt.Sprintf("Seu chamado #%s agora está: %s", protocol, status),
			Type:   "service_request",
			Data:   data,
		})
		if err != nil {
			log.Printf("warning: failed to save status update notification for service request %d: %v", req.ID, err)
		}
	}(*userID, sr, newStatus)
}

func (h *ServiceRequestHandler) dispatchServiceRequestStatusUpdated(userID *int64, sr *models.ServiceRequest, newStatus string) {
	if h.pushTokenRepo == nil || h.pushService == nil || userID == nil {
		return
	}

	go func(uid int64, req *models.ServiceRequest, status string) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		tokens, err := h.pushTokenRepo.ListTokensByUser(ctx, uid)
		if err != nil {
			log.Printf("warning: failed to list push tokens for user %d: %v", uid, err)
			return
		}

		if len(tokens) == 0 {
			return
		}

		protocol := ""
		if req.ProtocolNumber != nil {
			protocol = *req.ProtocolNumber
		}

		data := map[string]any{
			"service_request_id": req.ID,
			"protocol_number":    protocol,
			"status":             status,
			"screen":             fmt.Sprintf("/(service-requests)/%d", req.ID),
		}

		title := "Status do chamado atualizado"
		body := fmt.Sprintf("Seu chamado #%s agora está: %s", protocol, status)

		if err := h.pushService.SendToUser(ctx, tokens, title, body, data); err != nil {
			log.Printf("warning: failed to send status update push notification for service request %d: %v", req.ID, err)
		}
	}(*userID, sr, newStatus)
}
