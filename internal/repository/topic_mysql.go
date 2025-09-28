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

type TopicMySQLRepository struct {
	db  *sqlx.DB
	cfg *config.Config
}

func NewTopicRepository(cfg *config.Config, db *sqlx.DB) *TopicMySQLRepository {
	return &TopicMySQLRepository{
		db:  db,
		cfg: cfg,
	}
}

func (r *TopicMySQLRepository) Create(ctx context.Context, topic *domain.Topic) error {
	topic.ID = uuid.New().String()
	topic.CreatedAt = time.Now()
	topic.UpdatedAt = time.Now()

	query := `
		INSERT INTO topics (id, title, description, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		topic.ID,
		topic.Title,
		topic.Description,
		topic.CreatedBy,
		topic.CreatedAt,
		topic.UpdatedAt,
	)

	if err != nil {
		logger.Error(fmt.Errorf("failed to create topic: %w", err))
		return err
	}

	logger.Debug(fmt.Sprintf("topic created with ID: %s by user: %s", topic.ID, topic.CreatedBy))
	return nil
}

func (r *TopicMySQLRepository) GetByID(ctx context.Context, id string) (*domain.Topic, error) {
	var topic domain.Topic
	query := `
		SELECT id, title, description, created_by, created_at, updated_at
		FROM topics
		WHERE id = ?
	`

	err := r.db.GetContext(ctx, &topic, query, id)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return nil, fmt.Errorf("topic not found")
		}
		logger.Error(fmt.Errorf("failed to get topic by ID %s: %w", id, err))
		return nil, err
	}

	return &topic, nil
}

func (r *TopicMySQLRepository) GetByCreatorID(ctx context.Context, creatorID string, offset, limit int) ([]domain.Topic, int, error) {
	var topics []domain.Topic
	var total int

	countQuery := `SELECT COUNT(*) FROM topics WHERE created_by = ?`
	err := r.db.GetContext(ctx, &total, countQuery, creatorID)
	if err != nil {
		logger.Error(fmt.Errorf("failed to count topics for creator %s: %w", creatorID, err))
		return nil, 0, err
	}

	query := `
		SELECT id, title, description, created_by, created_at, updated_at
		FROM topics
		WHERE created_by = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	err = r.db.SelectContext(ctx, &topics, query, creatorID, limit, offset)
	if err != nil {
		logger.Error(fmt.Errorf("failed to get topics for creator %s: %w", creatorID, err))
		return nil, 0, err
	}

	return topics, total, nil
}

func (r *TopicMySQLRepository) AddAssignments(ctx context.Context, assignments []domain.TopicAssignment) error {
	if len(assignments) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO topic_assignments (id, topic_id, student_id, assigned_by, assigned_at)
		VALUES (?, ?, ?, ?, ?)
	`

	for _, assignment := range assignments {
		_, err := tx.ExecContext(ctx, query,
			assignment.ID,
			assignment.TopicID,
			assignment.StudentID,
			assignment.AssignedBy,
			assignment.AssignedAt,
		)
		if err != nil {
			logger.Error(fmt.Errorf("failed to create assignment: %w", err))
			return fmt.Errorf("failed to create assignment: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.Debug(fmt.Sprintf("created %d assignments for topic %s", len(assignments), assignments[0].TopicID))
	return nil
}

func (r *TopicMySQLRepository) GetAssignmentsByStudentID(ctx context.Context, studentID string) ([]domain.TopicAssignment, error) {
	var assignments []domain.TopicAssignment

	query := `
		SELECT id, topic_id, student_id, assigned_by, assigned_at
		FROM topic_assignments
		WHERE student_id = ?
		ORDER BY assigned_at DESC
	`

	err := r.db.SelectContext(ctx, &assignments, query, studentID)
	if err != nil {
		logger.Error(fmt.Errorf("failed to get assignments for student %s: %w", studentID, err))
		return nil, err
	}

	return assignments, nil
}

func (r *TopicMySQLRepository) GetAssignmentsByTopicID(ctx context.Context, topicID string) ([]domain.TopicAssignment, error) {
	var assignments []domain.TopicAssignment

	query := `
		SELECT id, topic_id, student_id, assigned_by, assigned_at
		FROM topic_assignments
		WHERE topic_id = ?
		ORDER BY assigned_at DESC
	`

	err := r.db.SelectContext(ctx, &assignments, query, topicID)
	if err != nil {
		logger.Error(fmt.Errorf("failed to get assignments for topic %s: %w", topicID, err))
		return nil, err
	}

	return assignments, nil
}

func (r *TopicMySQLRepository) GetAssignmentByID(ctx context.Context, id string) (*domain.TopicAssignment, error) {
	var assignment domain.TopicAssignment

	query := `
		SELECT id, topic_id, student_id, assigned_by, assigned_at
		FROM topic_assignments
		WHERE id = ?
	`

	err := r.db.GetContext(ctx, &assignment, query, id)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return nil, fmt.Errorf("assignment not found")
		}
		logger.Error(fmt.Errorf("failed to get assignment by ID %s: %w", id, err))
		return nil, err
	}

	return &assignment, nil
}