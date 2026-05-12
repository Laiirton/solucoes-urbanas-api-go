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

// GetUserRole returns the user's type string, or nil if not found
func GetUserRole(ctx context.Context, userRepo *repository.UserRepository, userID int64) *string {
	user, err := userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil
	}
	return user.Type
}

// GetTeamFilterForUser returns the team_id for attendant/secretary users who have a team.
// Returns nil if the user is not part of a team.
func GetTeamFilterForUser(ctx context.Context, userRepo *repository.UserRepository, userID int64) *int64 {
	user, err := userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil
	}

	// Only apply team filter for attendant and secretary
	if user.Type == nil || (*user.Type != "attendant" && *user.Type != "secretary") {
		return nil
	}

	return user.TeamID
}

// CanManageTeam checks if the user is admin OR the secretary/admin of the specified team.
// Used for team member management.
func CanManageTeam(ctx context.Context, userRepo *repository.UserRepository, userID, teamID int64) bool {
	user, err := userRepo.GetUserByID(ctx, userID)
	if err != nil || user.Type == nil {
		return false
	}

	// Admin can manage any team
	if *user.Type == "admin" {
		return true
	}

	// Secretary can manage their own team
	if *user.Type == "secretary" && user.TeamID != nil && *user.TeamID == teamID {
		return true
	}

	return false
}
