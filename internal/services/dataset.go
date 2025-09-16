package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/anton1ks96/college-core-api/internal/config"
	"github.com/anton1ks96/college-core-api/internal/domain"
	"github.com/anton1ks96/college-core-api/pkg/logger"
	"github.com/google/uuid"
)

type DatasetServiceImpl struct {
	repos      *Repositories
	ragService RAGService
	cfg        *config.Config
}

func NewDatasetService(repos *Repositories, ragService RAGService, cfg *config.Config) *DatasetServiceImpl {
	return &DatasetServiceImpl{
		repos:      repos,
		ragService: ragService,
		cfg:        cfg,
	}
}

func (s *DatasetServiceImpl) Create(ctx context.Context, userID, title string, content io.Reader) (*domain.Dataset, error) {
	count, err := s.repos.Dataset.CountByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count user datasets: %w", err)
	}

	if count >= s.cfg.Limits.MaxDatasetsPerUser {
		return nil, fmt.Errorf("dataset limit exceeded: maximum %d datasets allowed", s.cfg.Limits.MaxDatasetsPerUser)
	}

	buf := new(bytes.Buffer)
	size, err := io.Copy(buf, content)
	if err != nil {
		return nil, fmt.Errorf("failed to read content: %w", err)
	}

	if size > s.cfg.Limits.MaxFileSize {
		return nil, fmt.Errorf("file size exceeds limit: %d > %d bytes", size, s.cfg.Limits.MaxFileSize)
	}

	dataset := &domain.Dataset{
		ID:       uuid.New().String(),
		UserID:   userID,
		Title:    title,
		FilePath: fmt.Sprintf("students/%s/%s/current.md", userID, uuid.New().String()),
	}

	err = s.repos.File.Upload(ctx, dataset.FilePath, bytes.NewReader(buf.Bytes()), "text/markdown")
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	err = s.repos.Dataset.Create(ctx, dataset)
	if err != nil {
		_ = s.repos.File.Delete(ctx, dataset.FilePath)
		return nil, fmt.Errorf("failed to save dataset metadata: %w", err)
	}

	go func() {
		chunks, err := s.ragService.IndexDataset(context.Background(), dataset.ID, title, buf.String())
		if err != nil {
			logger.Error(fmt.Errorf("failed to index dataset %s: %w", dataset.ID, err))
			return
		}

		_ = s.repos.Dataset.UpdateIndexedAt(context.Background(), dataset.ID)
		logger.Info(fmt.Sprintf("dataset %s indexed successfully with %d chunks", dataset.ID, chunks))
	}()

	return dataset, nil
}

func (s *DatasetServiceImpl) GetByID(ctx context.Context, datasetID, userID string, role string) (*domain.DatasetResponse, error) {
	dataset, err := s.repos.Dataset.GetByID(ctx, datasetID)
	if err != nil {
		return nil, err
	}

	if !checkPermission(userID, dataset.UserID, role) {
		return nil, fmt.Errorf("access denied")
	}

	content, err := s.repos.File.Download(ctx, dataset.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	response := &domain.DatasetResponse{
		ID:        dataset.ID,
		Title:     dataset.Title,
		Content:   string(content),
		Author:    dataset.UserID, // TODO: получать имя пользователя из auth-svc
		UserID:    dataset.UserID,
		CreatedAt: dataset.CreatedAt,
		UpdatedAt: dataset.UpdatedAt,
		IndexedAt: dataset.IndexedAt,
	}

	return response, nil
}

func (s *DatasetServiceImpl) GetList(ctx context.Context, userID string, role string, page, limit int) (*domain.DatasetListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit
	var datasets []domain.Dataset
	var total int
	var err error

	if role == "teacher" || role == "admin" {
		datasets, total, err = s.repos.Dataset.GetAll(ctx, offset, limit)
	} else {
		datasets, total, err = s.repos.Dataset.GetByUserID(ctx, userID, offset, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get datasets: %w", err)
	}

	return &domain.DatasetListResponse{
		Datasets: datasets,
		Total:    total,
		Page:     page,
		Limit:    limit,
	}, nil
}

func (s *DatasetServiceImpl) Update(ctx context.Context, datasetID, userID, title string, content *string) (*domain.Dataset, error) {
	dataset, err := s.repos.Dataset.GetByID(ctx, datasetID)
	if err != nil {
		return nil, err
	}

	if !checkEditPermission(userID, dataset.UserID) {
		return nil, fmt.Errorf("access denied: only owner can edit dataset")
	}

	if title != "" {
		dataset.Title = title
	}

	if content != nil && *content != "" {
		if err := s.repos.File.Upload(ctx, dataset.FilePath, strings.NewReader(*content), "text/markdown"); err != nil {
			return nil, fmt.Errorf("failed to upload new content: %w", err)
		}
	}

	if err := s.repos.Dataset.Update(ctx, dataset); err != nil {
		return nil, fmt.Errorf("failed to update dataset: %w", err)
	}

	return dataset, nil
}

func (s *DatasetServiceImpl) Delete(ctx context.Context, datasetID, userID string) error {
	dataset, err := s.repos.Dataset.GetByID(ctx, datasetID)
	if err != nil {
		return err
	}

	if !checkEditPermission(userID, dataset.UserID) {
		return fmt.Errorf("access denied: only owner can delete dataset")
	}

	err = s.repos.Dataset.Delete(ctx, datasetID)
	if err != nil {
		return fmt.Errorf("failed to delete dataset: %w", err)
	}

	logger.Info(fmt.Sprintf("dataset %s soft deleted by user %s", datasetID, userID))
	return nil
}

func (s *DatasetServiceImpl) AskQuestion(ctx context.Context, datasetID, userID, role, question string) (*domain.AskResponse, error) {
	dataset, err := s.repos.Dataset.GetByID(ctx, datasetID)
	if err != nil {
		return nil, err
	}

	if !checkPermission(userID, dataset.UserID, role) {
		return nil, fmt.Errorf("access denied")
	}

	if dataset.IndexedAt == nil {
		return nil, fmt.Errorf("dataset is not indexed yet, please wait")
	}

	response, err := s.ragService.AskQuestion(ctx, datasetID, question)
	if err != nil {
		return nil, fmt.Errorf("failed to get answer: %w", err)
	}

	return response, nil
}

func (s *DatasetServiceImpl) Reindex(ctx context.Context, datasetID, userID string) (*domain.IndexResponse, error) {
	dataset, err := s.repos.Dataset.GetByID(ctx, datasetID)
	if err != nil {
		return nil, err
	}

	if !checkEditPermission(userID, dataset.UserID) {
		return nil, fmt.Errorf("access denied: only owner can reindex dataset")
	}

	content, err := s.repos.File.Download(ctx, dataset.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	chunks, err := s.ragService.IndexDataset(ctx, datasetID, dataset.Title, string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to index dataset: %w", err)
	}

	err = s.repos.Dataset.UpdateIndexedAt(ctx, datasetID)
	if err != nil {
		logger.Error(fmt.Errorf("failed to update indexed_at: %w", err))
	}

	return &domain.IndexResponse{
		Success: true,
		Chunks:  chunks,
		Message: fmt.Sprintf("Dataset reindexed successfully with %d chunks", chunks),
	}, nil
}
