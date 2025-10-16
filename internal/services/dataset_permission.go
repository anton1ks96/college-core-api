package services

import (
	"context"
	"fmt"

	"github.com/anton1ks96/college-core-api/internal/domain"
)

type DatasetPermissionServiceImpl struct {
	repos *Repositories
}

func NewDatasetPermissionService(repos *Repositories) *DatasetPermissionServiceImpl {
	return &DatasetPermissionServiceImpl{
		repos: repos,
	}
}

func (s *DatasetPermissionServiceImpl) GrantDatasetPermission(ctx context.Context, datasetID, teacherID, teacherName, grantedBy string) (string, error) {
	dataset, err := s.repos.Dataset.GetByID(ctx, datasetID)
	if err != nil {
		return "", err
	}

	if dataset == nil {
		return "", fmt.Errorf("dataset not found")
	}

	hasPermission, err := s.repos.DatasetPermission.HasPermission(ctx, datasetID, teacherID)
	if err != nil {
		return "", fmt.Errorf("failed to check existing permission: %w", err)
	}

	if hasPermission {
		return "", fmt.Errorf("permission already exists")
	}

	permission := &domain.DatasetPermission{
		DatasetID:   datasetID,
		TeacherID:   teacherID,
		TeacherName: teacherName,
		GrantedBy:   grantedBy,
	}

	if err := s.repos.DatasetPermission.GrantPermission(ctx, permission); err != nil {
		return "", fmt.Errorf("failed to grant permission: %w", err)
	}

	return permission.ID, nil
}

func (s *DatasetPermissionServiceImpl) RevokeDatasetPermission(ctx context.Context, datasetID, teacherID string) error {
	dataset, err := s.repos.Dataset.GetByID(ctx, datasetID)
	if err != nil {
		return err
	}

	if dataset == nil {
		return fmt.Errorf("dataset not found")
	}

	if err := s.repos.DatasetPermission.RevokePermission(ctx, datasetID, teacherID); err != nil {
		return fmt.Errorf("failed to revoke permission: %w", err)
	}

	return nil
}

func (s *DatasetPermissionServiceImpl) GetAllPermissions(ctx context.Context, page, limit int) ([]domain.DatasetPermission, int, error) {
	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	permissions, total, err := s.repos.DatasetPermission.GetAllPermissions(ctx, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get all permissions: %w", err)
	}

	return permissions, total, nil
}

func (s *DatasetPermissionServiceImpl) GetDatasetPermissions(ctx context.Context, datasetID string) ([]domain.DatasetPermission, error) {
	dataset, err := s.repos.Dataset.GetByID(ctx, datasetID)
	if err != nil {
		return nil, err
	}

	if dataset == nil {
		return nil, fmt.Errorf("dataset not found")
	}

	permissions, err := s.repos.DatasetPermission.GetPermissionsByDatasetID(ctx, datasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset permissions: %w", err)
	}

	return permissions, nil
}
