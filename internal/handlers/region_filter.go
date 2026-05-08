package handlers

import (
	"context"

	"github.com/laiirton/solucoes-urbanas-api/internal/repository"
)

// GetRegionFilterForAdmin returns the region_id the admin should be scoped to.
// If the user is an admin with a team that has a region assigned, it returns the region_id.
// Regular users or admins without a team region get nil (no filter).
func GetRegionFilterForAdmin(ctx context.Context, userRepo *repository.UserRepository, userID int64) *int64 {
	user, err := userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil
	}

	isAdmin := user.Type != nil && *user.Type == "admin"
	if !isAdmin {
		return nil
	}

	// Admin with a team → filter by the team's region
	if user.Team != nil && user.Team.RegionID != nil {
		return user.Team.RegionID
	}

	// Admin without team → no filter
	return nil
}
