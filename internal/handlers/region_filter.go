package handlers

import (
	"context"

	"github.com/laiirton/solucoes-urbanas-api/internal/models"
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
	return GetRegionFilterForUser(user)
}

// GetRegionFilterForUser is the User-variant that avoids a redundant DB fetch.
func GetRegionFilterForUser(user *models.User) *int64 {
	isAdmin := user.Type != nil && *user.Type == "admin"
	if !isAdmin {
		return nil
	}

	if user.Team != nil && user.Team.RegionID != nil {
		return user.Team.RegionID
	}

	return nil
}

// GetUserRole returns the user's type string, or nil if not found
func GetUserRole(ctx context.Context, userRepo *repository.UserRepository, userID int64) *string {
	user, err := userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil
	}
	return GetUserRoleForUser(user)
}

// GetUserRoleForUser is the User-variant that avoids a redundant DB fetch.
func GetUserRoleForUser(user *models.User) *string {
	return user.Type
}

// GetTeamFilterForUser returns the team_id for attendant/secretary users who have a team.
// Returns nil if the user is not part of a team.
func GetTeamFilterForUser(ctx context.Context, userRepo *repository.UserRepository, userID int64) *int64 {
	user, err := userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil
	}
	return GetTeamFilterForUserForUser(user)
}

// GetTeamFilterForUserForUser is the User-variant that avoids a redundant DB fetch.
func GetTeamFilterForUserForUser(user *models.User) *int64 {
	if user.Type == nil || (*user.Type != "attendant" && *user.Type != "secretary") {
		return nil
	}
	return user.TeamID
}

// CanManageTeam checks if the user is admin OR the secretary/admin of the specified team.
func CanManageTeam(ctx context.Context, userRepo *repository.UserRepository, userID, teamID int64) bool {
	user, err := userRepo.GetUserByID(ctx, userID)
	if err != nil || user.Type == nil {
		return false
	}

	if *user.Type == "admin" {
		return true
	}

	if *user.Type == "secretary" && user.TeamID != nil && *user.TeamID == teamID {
		return true
	}

	return false
}

// CanManageRequest checks if the user can manage the given service request.
func CanManageRequest(ctx context.Context, userRepo *repository.UserRepository, srRepo *repository.ServiceRequestRepository, userID, requestID int64) bool {
	user, err := userRepo.GetUserByID(ctx, userID)
	if err != nil || user.Type == nil {
		return false
	}

	if *user.Type == "admin" {
		return true
	}

	if *user.Type != "secretary" && *user.Type != "attendant" {
		return false
	}

	sr, err := srRepo.GetServiceRequestByID(ctx, requestID)
	if err != nil || sr.TeamID == nil || user.TeamID == nil {
		return false
	}

	return *sr.TeamID == *user.TeamID
}
