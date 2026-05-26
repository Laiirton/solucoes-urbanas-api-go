package repository

import (
	"context"
	"fmt"
)

type CompletedRequest struct {
	ID             int64
	UserID         int64
	ProtocolNumber string
	ServiceTitle   string
}

func (r *ServiceRequestRepository) FindCompletedUnrated(ctx context.Context) ([]CompletedRequest, error) {
	query := `
		SELECT sr.id, sr.user_id, COALESCE(sr.protocol_number, ''), sr.service_title
		FROM service_requests sr
		LEFT JOIN service_ratings rat ON rat.service_request_id = sr.id
		WHERE sr.status = 'completed'
		  AND sr.user_id IS NOT NULL
		  AND rat.id IS NULL
		  AND NOT EXISTS (
			SELECT 1 FROM system_notifications sn
			WHERE sn.type = 'rating_reminder'
			  AND sn.user_id = sr.user_id
			  AND sn.data->>'service_request_id' = sr.id::text
			  AND sn.created_at > NOW() - INTERVAL '48 hours'
		  )
		ORDER BY sr.updated_at DESC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find completed unrated requests: %w", err)
	}
	defer rows.Close()

	var results []CompletedRequest
	for rows.Next() {
		var cr CompletedRequest
		if err := rows.Scan(&cr.ID, &cr.UserID, &cr.ProtocolNumber, &cr.ServiceTitle); err != nil {
			return nil, fmt.Errorf("failed to scan completed request: %w", err)
		}
		results = append(results, cr)
	}
	if results == nil {
		results = []CompletedRequest{}
	}
	return results, nil
}
