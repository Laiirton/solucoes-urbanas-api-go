package services

import (
	"context"
	"encoding/json"
	"mime/multipart"

	"github.com/laiirton/solucoes-urbanas-api/internal/models"
	"github.com/laiirton/solucoes-urbanas-api/internal/repository"
)

type ServiceRequestService struct {
	repo          *repository.ServiceRequestRepository
	userRepo      *repository.UserRepository
	uploadService *UploadService
}

func NewServiceRequestService(repo *repository.ServiceRequestRepository, userRepo *repository.UserRepository, uploadService *UploadService) *ServiceRequestService {
	return &ServiceRequestService{
		repo:          repo,
		userRepo:      userRepo,
		uploadService: uploadService,
	}
}

func (s *ServiceRequestService) Create(ctx context.Context, userID int64, req *models.CreateServiceRequestRequest, files []*multipart.FileHeader) (*models.ServiceRequest, error) {
	// 1. Handle file uploads
	attachmentURLs, err := s.uploadService.UploadServiceRequestFiles(userID, files)
	if err != nil {
		return nil, err
	}

	if len(attachmentURLs) > 0 {
		urlsJSON, _ := json.Marshal(attachmentURLs)
		req.Attachments = urlsJSON
	}

	// 2. Persist in DB
	sr, err := s.repo.CreateServiceRequest(ctx, &userID, req)
	if err != nil {
		// Rollback uploaded files if DB insert fails
		if len(attachmentURLs) > 0 {
			s.uploadService.RollbackFiles(attachmentURLs)
		}
		return nil, err
	}

	return sr, nil
}

func (s *ServiceRequestService) List(ctx context.Context, userID int64, search string, isAdmin bool, all bool, page, limit int) ([]*models.ServiceRequest, error) {
	var categoryFilter string
	if isAdmin {
		user, err := s.userRepo.GetUserByID(ctx, userID)
		if err == nil && user.Team != nil {
			categoryFilter = user.Team.ServiceCategory
		}
	}

	if all && isAdmin {
		return s.repo.ListServiceRequests(ctx, search, categoryFilter, page, limit)
	}
	return s.repo.ListServiceRequestsByUser(ctx, userID, search, categoryFilter, page, limit)
}

func (s *ServiceRequestService) GetDetails(ctx context.Context, id int64) (*models.ServiceRequestDetailResponse, error) {
	sr, err := s.repo.GetServiceRequestByID(ctx, id)
	if err != nil {
		return nil, err
	}

	detail := &models.ServiceRequestDetailResponse{ServiceRequest: sr}
	if sr.UserID != nil {
		user, err := s.userRepo.GetUserByID(ctx, *sr.UserID)
		if err == nil {
			detail.CreatedBy = user
		}
		count, _ := s.repo.CountServiceRequestsByUser(ctx, *sr.UserID)
		detail.UserRequests = count
	}

	return detail, nil
}

func (s *ServiceRequestService) UpdateStatus(ctx context.Context, id int64, status string) (*models.ServiceRequest, error) {
	return s.repo.UpdateServiceRequestStatus(ctx, id, status)
}

func (s *ServiceRequestService) Delete(ctx context.Context, id int64) error {
	// 1. Fetch to get attachments
	sr, err := s.repo.GetServiceRequestByID(ctx, id)
	if err != nil {
		return err
	}

	// 2. Delete from DB
	if err := s.repo.DeleteServiceRequest(ctx, id); err != nil {
		return err
	}

	// 3. Cleanup files
	if urls := ParseAttachmentURLs(sr.Attachments); len(urls) > 0 {
		s.uploadService.RollbackFiles(urls)
	}

	return nil
}

func (s *ServiceRequestService) GetHomeStats(ctx context.Context, userID int64) (*models.HomeResponse, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	isAdmin := user.Type != nil && *user.Type == "admin"
	var categoryFilter string
	if isAdmin && user.Team != nil {
		categoryFilter = user.Team.ServiceCategory
	}

	return s.repo.GetHomeStats(ctx, isAdmin, userID, categoryFilter)
}
