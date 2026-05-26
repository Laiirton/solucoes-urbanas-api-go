package handlers

import (
	"context"
	"strings"

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

// CanViewRequest checks if the user can view the service request.
// Returns true if the user owns the request OR can manage it (admin/secretary/attendant of the team).
func CanViewRequest(ctx context.Context, userRepo *repository.UserRepository, srRepo *repository.ServiceRequestRepository, userID, requestID int64) bool {
	sr, err := srRepo.GetServiceRequestByID(ctx, requestID)
	if err != nil {
		return false
	}
	if sr.UserID != nil && *sr.UserID == userID {
		return true
	}
	return CanManageRequest(ctx, userRepo, srRepo, userID, requestID)
}

// canViewRequestFromSR checks if user can view the already-fetched SR.
func canViewRequestFromSR(ctx context.Context, userRepo *repository.UserRepository, userID int64, sr *models.ServiceRequest) bool {
	if sr.UserID != nil && *sr.UserID == userID {
		return true
	}
	return canManageRequestFromSR(ctx, userRepo, userID, sr)
}

// CanManageRequest checks if the user can manage the given service request.
// This is the public variant that fetches the SR from DB.
func CanManageRequest(ctx context.Context, userRepo *repository.UserRepository, srRepo *repository.ServiceRequestRepository, userID, requestID int64) bool {
	sr, err := srRepo.GetServiceRequestByID(ctx, requestID)
	if err != nil {
		return false
	}
	return canManageRequestFromSR(ctx, userRepo, userID, sr)
}

// canManageRequestFromSR checks if user can manage the already-fetched SR.
func canManageRequestFromSR(ctx context.Context, userRepo *repository.UserRepository, userID int64, sr *models.ServiceRequest) bool {
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

	if user.TeamID == nil || user.Team == nil {
		return false
	}

	// Case 1: The request is already assigned to a team
	if sr.TeamID != nil {
		return *sr.TeamID == *user.TeamID
	}

	// Case 2: The request is NOT assigned (TeamID == nil)
	var categories []string
	categories = append(categories, user.Team.Categories...)
	categories = append(categories, user.WorkArea...)

	if !categoryMatches(sr.Category, categories) {
		return false
	}

	if user.Team.CityWide {
		return true
	}

	if sr.RegionID != nil && user.Team.RegionID != nil && *sr.RegionID == *user.Team.RegionID {
		return true
	}

	return false
}

var accentReplacer = strings.NewReplacer(
	"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"í", "i", "ì", "i", "î", "i", "ï", "i",
	"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
	"ú", "u", "ù", "u", "û", "u", "ü", "u",
	"ç", "c", "ñ", "n",
)

func normalizeString(s string) string {
	return accentReplacer.Replace(strings.ToLower(s))
}

func categoryMatches(targetCat string, categories []string) bool {
	normTarget := normalizeString(targetCat)
	for _, cat := range categories {
		if normalizeString(cat) == normTarget {
			return true
		}
	}
	return false
}
