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

type DatasetPermissionMySQLRepository struct {
	db  *sqlx.DB
	cfg *config.Config
}

func NewDatasetPermissionRepository(cfg *config.Config, db *sqlx.DB) *DatasetPermissionMySQLRepository {
	return &DatasetPermissionMySQLRepository{
		db:  db,
		cfg: cfg,
	}
}

func (r *DatasetPermissionMySQLRepository) GrantPermission(ctx context.Context, permission *domain.DatasetPermission) error {
	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("failed to generate UUID v7: %w", err)
	}

	permission.ID = id.String()
	permission.GrantedAt = time.Now()

	query := `
		INSERT INTO datasets_permission (id, dataset_id, teacher_id, teacher_name, granted_by, granted_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecContext(ctx, query,
		permission.ID,
		permission.DatasetID,
		permission.TeacherID,
		permission.TeacherName,
		permission.GrantedBy,
		permission.GrantedAt,
	)

	if err != nil {
		logger.Error(fmt.Errorf("failed to grant permission: %w", err))
		return err
	}

	logger.Debug(fmt.Sprintf("permission granted: dataset %s to teacher %s by %s", permission.DatasetID, permission.TeacherID, permission.GrantedBy))
	return nil
}

func (r *DatasetPermissionMySQLRepository) RevokePermission(ctx context.Context, datasetID, teacherID string) error {
	query := `
		DELETE FROM datasets_permission
		WHERE dataset_id = ? AND teacher_id = ?
	`

	result, err := r.db.ExecContext(ctx, query, datasetID, teacherID)
	if err != nil {
		logger.Error(fmt.Errorf("failed to revoke permission: %w", err))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("permission not found")
	}

	logger.Debug(fmt.Sprintf("permission revoked: dataset %s from teacher %s", datasetID, teacherID))
	return nil
}

func (r *DatasetPermissionMySQLRepository) HasPermission(ctx context.Context, datasetID, teacherID string) (bool, error) {
	var count int

	query := `
		SELECT COUNT(*)
		FROM datasets_permission
		WHERE dataset_id = ? AND teacher_id = ?
	`

	err := r.db.GetContext(ctx, &count, query, datasetID, teacherID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Error(fmt.Errorf("failed to check permission: %w", err))
		return false, err
	}

	return count > 0, nil
}

func (r *DatasetPermissionMySQLRepository) GetAllPermissions(ctx context.Context, offset, limit int) ([]domain.DatasetPermission, int, error) {
	var datasetsPerms []domain.DatasetPermission
	var total int

	countQuery := `SELECT COUNT(*) FROM datasets_permission`
	err := r.db.GetContext(ctx, &total, countQuery)
	if err != nil {
		logger.Error(fmt.Errorf("failed to count all datasets permissions: %w", err))
		return nil, 0, err
	}

	query := `
		SELECT
			dp.id,
			dp.dataset_id,
			d.title as dataset_title,
			dp.teacher_id,
			dp.teacher_name,
			dp.granted_by,
			dp.granted_at
		FROM datasets_permission dp
		LEFT JOIN datasets d ON dp.dataset_id = d.id
		ORDER BY dp.granted_at DESC
		LIMIT ? OFFSET ?
	`

	err = r.db.SelectContext(ctx, &datasetsPerms, query, limit, offset)
	if err != nil {
		logger.Error(fmt.Errorf("failed to get all datasets permissions: %w", err))
		return nil, 0, err
	}

	return datasetsPerms, total, nil
}
