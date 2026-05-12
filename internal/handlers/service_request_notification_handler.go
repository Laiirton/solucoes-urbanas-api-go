package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/laiirton/solucoes-urbanas-api/internal/models"
)

func (h *ServiceRequestHandler) SaveServiceRequestStatusUpdatedNotification(userID *int64, sr *models.ServiceRequest, newStatus string) {
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
			Body:   fmt.Sprintf("Seu chamado #%s agora está: %s", protocol, statusLabels[status]),
			Type:   "service_request",
			Data:   data,
		})
		if err != nil {
			log.Printf("warning: failed to save status update notification for service request %d: %v", req.ID, err)
		}
	}(*userID, sr, newStatus)
}

func (h *ServiceRequestHandler) DispatchServiceRequestStatusUpdated(userID *int64, sr *models.ServiceRequest, newStatus string) {
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
		body := fmt.Sprintf("Seu chamado #%s agora está: %s", protocol, statusLabels[status])

		if err := h.pushService.SendToUser(ctx, tokens, title, body, "default", data); err != nil {
			log.Printf("warning: failed to send status update push notification for service request %d: %v", req.ID, err)
		}
	}(*userID, sr, newStatus)
}
