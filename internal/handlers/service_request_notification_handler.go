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

func (h *ServiceRequestHandler) SaveChatMessageNotification(userID *int64, sr *models.ServiceRequest, senderName string, content string) {
	if h.sysNotifRepo == nil || userID == nil {
		return
	}

	go func(uid int64, req *models.ServiceRequest, name string, bodyPreview string) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		protocol := ""
		if req.ProtocolNumber != nil {
			protocol = *req.ProtocolNumber
		}

		// Truncate content for notification body (max 120 chars)
		shortBody := bodyPreview
		if len(shortBody) > 120 {
			shortBody = shortBody[:120] + "..."
		}

		data, _ := json.Marshal(map[string]interface{}{
			"service_request_id": req.ID,
			"protocol_number":    protocol,
			"screen":             fmt.Sprintf("/(service-requests)/%d", req.ID),
		})

		_, err := h.sysNotifRepo.Create(ctx, &models.SystemNotification{
			UserID: &uid,
			Title:  fmt.Sprintf("Nova mensagem de %s", name),
			Body:   shortBody,
			Type:   "chat_message",
			Data:   data,
		})
		if err != nil {
			log.Printf("warning: failed to save chat message notification for service request %d: %v", req.ID, err)
		}
	}(*userID, sr, senderName, content)
}

func (h *ServiceRequestHandler) DispatchChatMessageNotification(userID *int64, sr *models.ServiceRequest, senderName string, content string) {
	if h.pushTokenRepo == nil || h.pushService == nil || userID == nil {
		return
	}

	go func(uid int64, req *models.ServiceRequest, name string, bodyPreview string) {
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
			"screen":             fmt.Sprintf("/(service-requests)/%d", req.ID),
			"type":               "chat_message",
		}

		shortBody := bodyPreview
		if len(shortBody) > 120 {
			shortBody = shortBody[:120] + "..."
		}

		title := fmt.Sprintf("Nova mensagem de %s", name)

		if err := h.pushService.SendToUser(ctx, tokens, title, shortBody, "default", data); err != nil {
			log.Printf("warning: failed to send chat message push notification for service request %d: %v", req.ID, err)
		}
	}(*userID, sr, senderName, content)
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
