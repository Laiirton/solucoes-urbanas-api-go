package handlers

import (
	"context"
	"encoding/json"
	"log"
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
	srRepo          *repository.ServiceRequestRepository
	userRepo        *repository.UserRepository
	regionRepo      *repository.RegionRepository
	teamRepo        *repository.TeamRepository
	sysNotifRepo    *repository.SystemNotificationRepository
	pushTokenRepo   *repository.PushTokenRepository
	pushService     *services.ExpoPushService
	uploadService   *services.UploadService
	geoService      *services.GeocodingService
	ratingRepo      *repository.ServiceRatingRepository
	attendanceRepo  *repository.ServiceAttendanceRepository
	chatMessageRepo *repository.ChatMessageRepository
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
	chatMessageRepo *repository.ChatMessageRepository,
) *ServiceRequestHandler {
	return &ServiceRequestHandler{
		srRepo:          srRepo,
		userRepo:        userRepo,
		regionRepo:      regionRepo,
		teamRepo:        teamRepo,
		sysNotifRepo:    sysNotifRepo,
		pushTokenRepo:   pushTokenRepo,
		pushService:     pushService,
		uploadService:   uploadService,
		geoService:      geoService,
		ratingRepo:      ratingRepo,
		attendanceRepo:  attendanceRepo,
		chatMessageRepo: chatMessageRepo,
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
	var serviceCategory string

	// 0. Get service category for team routing (match by work_area)
	if req.ServiceID != nil {
		cat, err := h.srRepo.GetServiceCategory(r.Context(), *req.ServiceID)
		if err == nil {
			serviceCategory = cat
		}
	}

	// 0.5 Team override: attendant/secretary always create for their own team
	var appBairro string
	var appLat, appLon *float64
	var appAddress string

	currentUserForTeam, _ := h.userRepo.GetUserByID(r.Context(), userID)
	if currentUserForTeam != nil && currentUserForTeam.TeamID != nil &&
		(currentUserForTeam.Type != nil && (*currentUserForTeam.Type == "attendant" || *currentUserForTeam.Type == "secretary")) {
		teamID = currentUserForTeam.TeamID
		teamForRegion, regionErr := h.teamRepo.GetTeamByID(r.Context(), *teamID)
		if regionErr == nil && teamForRegion.RegionID != nil {
			regionID = teamForRegion.RegionID
		}
	} else {
		appBairro = extractBairroFromRequestData(req.RequestData)
		appLat, appLon = extractCoordinatesFromRequestData(req.RequestData)
		appAddress = extractAddressFromRequestData(req.RequestData)

		if appBairro != "" {
			regionID, teamID, _ = h.lookupRegionAndTeam(r.Context(), appBairro, serviceCategory, regionID, teamID)
		}

		if regionID == nil && appLat != nil && appLon != nil {
			geoResult, err := h.geoService.ReverseGeocode(*appLat, *appLon)
			if err == nil && geoResult.Found {
				geoAddress = &geoResult.DisplayName
				if geoResult.Bairro != "" {
					regionID, teamID, _ = h.lookupRegionAndTeam(r.Context(), geoResult.Bairro, serviceCategory, regionID, teamID)
				}
			}
		}

		if regionID == nil && appAddress != "" {
			geoResult, err := h.geoService.GeocodeAddress(appAddress)
			if err == nil && geoResult.Found {
				if geoAddress == nil {
					geoAddress = &geoResult.DisplayName
				}
				if geoResult.Bairro != "" {
					regionID, teamID, _ = h.lookupRegionAndTeam(r.Context(), geoResult.Bairro, serviceCategory, regionID, teamID)
				}
			}
		}
	}

	// 5. Se encontrou região mas nenhuma equipe atende a categoria → bloquear
	if regionID != nil && teamID == nil {
		if urls := services.ParseAttachmentURLs(req.Attachments); len(urls) > 0 {
			h.uploadService.RollbackFiles(urls)
		}
		respondError(w, http.StatusBadRequest, "Nenhuma equipe disponível para atender esta solicitação na região informada.")
		return
	}

	// 6. Always use user's original coordinates when available
	if appLat != nil && appLon != nil {
		geoLat = appLat
		geoLon = appLon
	} else if geoLat == nil && geoLon == nil && appAddress != "" {
		// Preserve geocoded coords as fallback
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

	// Permission check: user must own the request OR be admin/secretary/attendant of the team
	currentUserID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if ok && !CanViewRequest(r.Context(), h.userRepo, h.srRepo, currentUserID, id) {
		respondError(w, http.StatusNotFound, "service request not found")
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

	chatMessages, _ := h.chatMessageRepo.ListByRequestID(r.Context(), id)
	detail.ChatMessages = chatMessages

	respondJSON(w, http.StatusOK, detail)
}

// PUT /service-requests/{id}
func (h *ServiceRequestHandler) UpdateServiceRequest(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid service request id")
		return
	}

	existing, err := h.srRepo.GetServiceRequestByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "service request not found")
		return
	}
	if existing.UserID == nil || *existing.UserID != userID {
		respondError(w, http.StatusForbidden, "you do not have permission to update this request")
		return
	}
	if existing.Status != "pending" {
		respondError(w, http.StatusBadRequest, "only pending requests can be edited")
		return
	}

	var req models.CreateServiceRequestRequest
	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(services.MaxTotalFilesSizeBytes); err != nil {
			respondError(w, http.StatusBadRequest, "failed to parse multipart form")
			return
		}

		requestData := r.FormValue("request_data")
		if requestData != "" {
			req.RequestData = []byte(requestData)
		} else {
			req.RequestData = []byte("{}")
		}

		// Preserve existing attachments that haven't been removed
		oldURLs := services.ParseAttachmentURLs(existing.Attachments)
		keepURLs := make(map[string]bool)
		for _, u := range oldURLs {
			keepURLs[u] = true
		}

		// Upload new files
		files := r.MultipartForm.File["files"]
		if len(files) > 0 {
			newURLs, err := h.uploadService.UploadServiceRequestFiles(userID, files)
			if err != nil {
				respondError(w, http.StatusBadRequest, err.Error())
				return
			}
			for _, u := range newURLs {
				keepURLs[u] = true
			}
		}

		var finalURLs []string
		for u := range keepURLs {
			finalURLs = append(finalURLs, u)
		}
		if len(finalURLs) > 0 {
			urlsJSON, _ := json.Marshal(finalURLs)
			req.Attachments = urlsJSON
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
	}

	if len(req.RequestData) == 0 {
		req.RequestData = []byte("{}")
	}

	sr, err := h.srRepo.UpdateServiceRequest(r.Context(), id, &req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, sr)
}

// PATCH /service-requests/{id}/notes
func (h *ServiceRequestHandler) UpdateServiceRequestNotes(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid service request id")
		return
	}

	// Permission check: must be able to manage the request
	currentUserID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if ok && !CanManageRequest(r.Context(), h.userRepo, h.srRepo, currentUserID, id) {
		respondError(w, http.StatusNotFound, "service request not found")
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Store notes in request_data as observacoes_internas
	existing, err := h.srRepo.GetServiceRequestByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "service request not found")
		return
	}

	var requestData map[string]interface{}
	if existing.RequestData != nil {
		json.Unmarshal(existing.RequestData, &requestData)
	}
	if requestData == nil {
		requestData = make(map[string]interface{})
	}
	requestData["observacoes_internas"] = req.Notes

	updatedData, _ := json.Marshal(requestData)

	updateReq := &models.CreateServiceRequestRequest{
		RequestData: updatedData,
	}

	sr, err := h.srRepo.UpdateServiceRequest(r.Context(), id, updateReq)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
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

	// Team ownership check for secretary/attendant
	currentUserID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if ok && !CanManageRequest(r.Context(), h.userRepo, h.srRepo, currentUserID, id) {
		respondError(w, http.StatusNotFound, "service request not found")
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

	// Only admin can delete service requests
	currentUserID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if ok {
		user, userErr := h.userRepo.GetUserByID(r.Context(), currentUserID)
		if userErr != nil || user.Type == nil || *user.Type != "admin" {
			respondError(w, http.StatusForbidden, "only admins can delete service requests")
			return
		}
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
	// Cascade delete related system notifications
	if h.sysNotifRepo != nil {
		if err := h.sysNotifRepo.DeleteByTypeAndRefID(r.Context(), "service_request", id); err != nil {
			log.Printf("warning: failed to delete system notifications for SR %d: %v", id, err)
		}
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

	// Try prioritized keys first
	if val := getFirstNonEmpty(data, "endereco", "address", "logradouro", "street"); val != "" {
		return val
	}

	// Dynamic fallback: find a string key where there's a corresponding key + "_coords" in the map
	for key, val := range data {
		if strVal, ok := val.(string); ok && strVal != "" {
			coordsKey := key + "_coords"
			if _, hasCoords := data[coordsKey]; hasCoords {
				return strVal
			}
		}
	}

	return ""
}

// extractBairroFromRequestData extrai o bairro enviado pelo app (via MapModal)
func extractBairroFromRequestData(requestData json.RawMessage) string {
	if len(requestData) == 0 {
		return ""
	}

	var data map[string]interface{}
	if err := json.Unmarshal(requestData, &data); err != nil {
		return ""
	}

	// Tenta "endereco_bairro" primeiro (enviado explicitamente pelo app)
	if bairro := getField[string](data, "endereco_bairro"); bairro != "" {
		return bairro
	}
	// Fallback: tenta "localizacao_bairro" (se o campo se chamar "localizacao")
	if bairro := getField[string](data, "localizacao_bairro"); bairro != "" {
		return bairro
	}

	// Dynamic fallback: find a string key ending with "_bairro"
	for key, val := range data {
		if strings.HasSuffix(key, "_bairro") {
			if strVal, ok := val.(string); ok && strVal != "" {
				return strVal
			}
		}
	}

	return ""
}

// extractCoordinatesFromRequestData extrai as coordenadas do JSON de request_data
// enviadas pelo app quando o usuário seleciona no mapa
func extractCoordinatesFromRequestData(requestData json.RawMessage) (*float64, *float64) {
	if len(requestData) == 0 {
		return nil, nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal(requestData, &data); err != nil {
		return nil, nil
	}

	// First try prioritized keys
	keys := []string{"endereco_coords", "localizacao_coords"}
	for _, key := range keys {
		if coords, ok := data[key].(map[string]interface{}); ok {
			lat, hasLat := getNestedFloat(data, key, "latitude")
			lon, hasLon := getNestedFloat(data, key, "longitude")
			if hasLat && hasLon {
				return &lat, &lon
			}
			// Try direct keys inside coords object
			if latVal, ok := coords["latitude"].(float64); ok {
				if lonVal, ok := coords["longitude"].(float64); ok {
					return &latVal, &lonVal
				}
			}
		}
	}

	// Dynamic fallback: look for any key ending with "_coords"
	for key, val := range data {
		if strings.HasSuffix(key, "_coords") {
			if coords, ok := val.(map[string]interface{}); ok {
				lat, hasLat := getNestedFloat(data, key, "latitude")
				lon, hasLon := getNestedFloat(data, key, "longitude")
				if hasLat && hasLon {
					return &lat, &lon
				}
				if latVal, ok := coords["latitude"].(float64); ok {
					if lonVal, ok := coords["longitude"].(float64); ok {
						return &latVal, &lonVal
					}
				}
			}
		}
	}

	return nil, nil
}

// getFirstNonEmpty returns the first non-empty string value for the given keys
func getFirstNonEmpty(data map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if val, ok := data[key].(string); ok && val != "" {
			return val
		}
	}
	return ""
}

// getField extracts a typed field from a map
func getField[T any](data map[string]interface{}, key string) T {
	var zero T
	if val, ok := data[key]; ok {
		if typed, ok := val.(T); ok {
			return typed
		}
	}
	return zero
}

// getNestedFloat extracts a float from a nested object path like "endereco_coords.latitude"
func getNestedFloat(data map[string]interface{}, objKey, fieldKey string) (float64, bool) {
	if obj, ok := data[objKey].(map[string]interface{}); ok {
		if val, ok := obj[fieldKey]; ok {
			switch v := val.(type) {
			case float64:
				return v, true
			case int:
				return float64(v), true
			case int64:
				return float64(v), true
			}
		}
	}
	return 0, false
}

// lookupRegionAndTeam tenta encontrar região pelo bairro e equipe pela região + categoria.
func (h *ServiceRequestHandler) lookupRegionAndTeam(ctx context.Context, bairro, serviceCategory string, currentRegionID, currentTeamID *int64) (*int64, *int64, bool) {
	// Step 1: Try to find region by neighborhood name
	if bairro != "" {
		region, err := h.regionRepo.FindByNeighborhood(ctx, bairro)
		if err != nil {
			region, err = h.regionRepo.FindByNeighborhoodCaseInsensitive(ctx, bairro)
		}
		if region != nil {
			regionID := &region.ID
			// Try to find a team for this region + category
			if serviceCategory != "" {
				team, err := h.teamRepo.FindTeamByRegionAndCategory(ctx, region.ID, serviceCategory)
				if err == nil {
					return regionID, &team.ID, true
				}
				return regionID, currentTeamID, false
			}
			return regionID, currentTeamID, true
		}
	}

	// Step 2: City-wide fallback — find any city-wide team that handles this category
	if serviceCategory != "" {
		team, err := h.teamRepo.FindCityWideTeamByCategory(ctx, serviceCategory)
		if err == nil {
			return nil, &team.ID, true
		}
	}

	if bairro == "" {
		return currentRegionID, currentTeamID, true
	}
	return currentRegionID, currentTeamID, false
}
