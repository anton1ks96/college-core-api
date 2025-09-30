package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/anton1ks96/college-core-api/internal/config"
	"github.com/anton1ks96/college-core-api/internal/domain"
	"github.com/anton1ks96/college-core-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type DatasetMySQLRepository struct {
	db  *sqlx.DB
	cfg *config.Config
}

func NewDatasetRepository(cfg *config.Config, db *sqlx.DB) *DatasetMySQLRepository {
	return &DatasetMySQLRepository{
		db:  db,
		cfg: cfg,
	}
}

func (r *DatasetMySQLRepository) Create(ctx context.Context, dataset *domain.Dataset) error {
	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("failed to generate UUID v7: %w", err)
	}

	dataset.ID = id.String()
	dataset.CreatedAt = time.Now()
	dataset.UpdatedAt = time.Now()

	query := `
		INSERT INTO datasets (id, user_id, author, title, file_path, created_at, updated_at, topic_id, assignment_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecContext(ctx, query,
		dataset.ID,
		dataset.UserID,
		dataset.Author,
		dataset.Title,
		dataset.FilePath,
		dataset.CreatedAt,
		dataset.UpdatedAt,
		dataset.TopicID,
		dataset.AssignmentID,
	)

	if err != nil {
		logger.Error(fmt.Errorf("failed to create dataset: %w", err))
		return err
	}

	logger.Debug(fmt.Sprintf("dataset created with ID: %s for user: %s", dataset.ID, dataset.UserID))
	return nil
}

func (r *DatasetMySQLRepository) GetByID(ctx context.Context, id string) (*domain.Dataset, error) {
	var dataset domain.Dataset
	query := `
		SELECT id, user_id, author, title, file_path, created_at, updated_at, indexed_at, topic_id, assignment_id
		FROM datasets
		WHERE id = ?
	`

	err := r.db.GetContext(ctx, &dataset, query, id)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return nil, fmt.Errorf("dataset not found")
		}
		logger.Error(fmt.Errorf("failed to get dataset by ID %s: %w", id, err))
		return nil, err
	}

	return &dataset, nil
}

func (r *DatasetMySQLRepository) GetByUserID(ctx context.Context, userID string, offset, limit int) ([]domain.Dataset, int, error) {
	var datasets []domain.Dataset
	var total int

	countQuery := `SELECT COUNT(*) FROM datasets WHERE user_id = ?`
	err := r.db.GetContext(ctx, &total, countQuery, userID)
	if err != nil {
		logger.Error(fmt.Errorf("failed to count datasets for user %s: %w", userID, err))
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, author, title, file_path, created_at, updated_at, indexed_at, topic_id, assignment_id
		FROM datasets
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	err = r.db.SelectContext(ctx, &datasets, query, userID, limit, offset)
	if err != nil {
		logger.Error(fmt.Errorf("failed to get datasets for user %s: %w", userID, err))
		return nil, 0, err
	}

	return datasets, total, nil
}

func (r *DatasetMySQLRepository) GetAll(ctx context.Context, offset, limit int) ([]domain.Dataset, int, error) {
	var datasets []domain.Dataset
	var total int

	countQuery := `SELECT COUNT(*) FROM datasets`
	err := r.db.GetContext(ctx, &total, countQuery)
	if err != nil {
		logger.Error(fmt.Errorf("failed to count all datasets: %w", err))
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, author, title, file_path, created_at, updated_at, indexed_at, topic_id, assignment_id
		FROM datasets
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	err = r.db.SelectContext(ctx, &datasets, query, limit, offset)
	if err != nil {
		logger.Error(fmt.Errorf("failed to get all datasets: %w", err))
		return nil, 0, err
	}

	return datasets, total, nil
}

func (r *DatasetMySQLRepository) Update(ctx context.Context, dataset *domain.Dataset) error {
	dataset.UpdatedAt = time.Now()

	query := `
		UPDATE datasets
		SET title = ?, file_path = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		dataset.Title,
		dataset.FilePath,
		dataset.UpdatedAt,
		dataset.ID,
	)

	if err != nil {
		logger.Error(fmt.Errorf("failed to update dataset %s: %w", dataset.ID, err))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("dataset not found or already deleted")
	}

	logger.Debug(fmt.Sprintf("dataset %s updated successfully", dataset.ID))
	return nil
}

func (r *DatasetMySQLRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE datasets
		SET updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		logger.Error(fmt.Errorf("failed to delete dataset %s: %w", id, err))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("dataset not found or already deleted")
	}

	logger.Debug(fmt.Sprintf("dataset %s soft deleted", id))
	return nil
}

func (r *DatasetMySQLRepository) UpdateIndexedAt(ctx context.Context, id string) error {
	now := time.Now()
	query := `
		UPDATE datasets
		SET indexed_at = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		logger.Error(fmt.Errorf("failed to update indexed_at for dataset %s: %w", id, err))
		return err
	}

	logger.Debug(fmt.Sprintf("dataset %s indexed_at and updated_at updated", id))
	return nil
}
