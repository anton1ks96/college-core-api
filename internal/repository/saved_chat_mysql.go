package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/anton1ks96/college-core-api/internal/config"
	"github.com/anton1ks96/college-core-api/internal/domain"
	"github.com/anton1ks96/college-core-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type SavedChatMySQLRepository struct {
	db  *sqlx.DB
	cfg *config.Config
}

func NewSavedChatRepository(cfg *config.Config, db *sqlx.DB) *SavedChatMySQLRepository {
	return &SavedChatMySQLRepository{
		db:  db,
		cfg: cfg,
	}
}

func (r *SavedChatMySQLRepository) Create(ctx context.Context, chat *domain.SavedChat) error {
	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("failed to generate UUID v7: %w", err)
	}

	chat.ID = id.String()
	chat.CreatedAt = time.Now()
	chat.UpdatedAt = time.Now()

	query := `
		INSERT INTO saved_chats (id, dataset_id, title, created_by, user_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecContext(ctx, query,
		chat.ID,
		chat.DatasetID,
		chat.Title,
		chat.CreatedBy,
		chat.UserID,
		chat.CreatedAt,
		chat.UpdatedAt,
	)

	if err != nil {
		logger.Error(fmt.Errorf("failed to create saved chat: %w", err))
		return err
	}

	logger.Debug(fmt.Sprintf("saved chat created with ID: %s for dataset: %s", chat.ID, chat.DatasetID))
	return nil
}

func (r *SavedChatMySQLRepository) GetByID(ctx context.Context, id string) (*domain.SavedChat, error) {
	var chat domain.SavedChat
	query := `
		SELECT id, dataset_id, title, created_by, user_id, created_at, updated_at
		FROM saved_chats
		WHERE id = ?
	`

	err := r.db.GetContext(ctx, &chat, query, id)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return nil, fmt.Errorf("chat not found")
		}
		logger.Error(fmt.Errorf("failed to get saved chat by ID %s: %w", id, err))
		return nil, err
	}

	return &chat, nil
}

func (r *SavedChatMySQLRepository) GetByDatasetID(ctx context.Context, datasetID string, offset, limit int) ([]domain.SavedChat, int, error) {
	var chats []domain.SavedChat
	var total int

	countQuery := `SELECT COUNT(*) FROM saved_chats WHERE dataset_id = ?`
	err := r.db.GetContext(ctx, &total, countQuery, datasetID)
	if err != nil {
		logger.Error(fmt.Errorf("failed to count saved chats for dataset %s: %w", datasetID, err))
		return nil, 0, err
	}

	query := `
		SELECT id, dataset_id, title, created_by, user_id, created_at, updated_at
		FROM saved_chats
		WHERE dataset_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	err = r.db.SelectContext(ctx, &chats, query, datasetID, limit, offset)
	if err != nil {
		logger.Error(fmt.Errorf("failed to get saved chats for dataset %s: %w", datasetID, err))
		return nil, 0, err
	}

	return chats, total, nil
}

func (r *SavedChatMySQLRepository) Update(ctx context.Context, chat *domain.SavedChat) error {
	chat.UpdatedAt = time.Now()

	query := `
		UPDATE saved_chats
		SET title = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		chat.Title,
		chat.UpdatedAt,
		chat.ID,
	)

	if err != nil {
		logger.Error(fmt.Errorf("failed to update saved chat %s: %w", chat.ID, err))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("chat not found")
	}

	logger.Debug(fmt.Sprintf("saved chat %s updated successfully", chat.ID))
	return nil
}

func (r *SavedChatMySQLRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM saved_chats WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		logger.Error(fmt.Errorf("failed to delete saved chat %s: %w", id, err))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("chat not found")
	}

	logger.Debug(fmt.Sprintf("saved chat %s deleted", id))
	return nil
}

func (r *SavedChatMySQLRepository) GetMessagesByChatID(ctx context.Context, chatID string) ([]domain.ChatMessage, error) {
	var messages []domain.ChatMessage

	query := `
		SELECT id, chat_id, question, answer, citations, order_num, created_at
		FROM chat_messages
		WHERE chat_id = ?
		ORDER BY order_num ASC
	`

	err := r.db.SelectContext(ctx, &messages, query, chatID)
	if err != nil {
		logger.Error(fmt.Errorf("failed to get messages for chat %s: %w", chatID, err))
		return nil, err
	}

	for i := range messages {
		if messages[i].CitationsJSON != "" {
			var citations []domain.Citation
			if err := json.Unmarshal([]byte(messages[i].CitationsJSON), &citations); err != nil {
				logger.Error(fmt.Errorf("failed to unmarshal citations for message %s: %w", messages[i].ID, err))
			} else {
				messages[i].Citations = citations
			}
		}
	}

	return messages, nil
}

func (r *SavedChatMySQLRepository) SaveMessages(ctx context.Context, chatID string, messages []domain.ChatMessage) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	deleteQuery := `DELETE FROM chat_messages WHERE chat_id = ?`
	_, err = tx.ExecContext(ctx, deleteQuery, chatID)
	if err != nil {
		logger.Error(fmt.Errorf("failed to delete old messages for chat %s: %w", chatID, err))
		return err
	}

	insertQuery := `
		INSERT INTO chat_messages (id, chat_id, question, answer, citations, order_num, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	for i := range messages {
		id, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("failed to generate UUID v7 for message: %w", err)
		}

		messages[i].ID = id.String()
		messages[i].ChatID = chatID
		messages[i].OrderNum = i + 1
		messages[i].CreatedAt = time.Now()

		citationsJSON := "[]"
		if len(messages[i].Citations) > 0 {
			jsonBytes, err := json.Marshal(messages[i].Citations)
			if err != nil {
				logger.Error(fmt.Errorf("failed to marshal citations: %w", err))
			} else {
				citationsJSON = string(jsonBytes)
			}
		}

		_, err = tx.ExecContext(ctx, insertQuery,
			messages[i].ID,
			messages[i].ChatID,
			messages[i].Question,
			messages[i].Answer,
			citationsJSON,
			messages[i].OrderNum,
			messages[i].CreatedAt,
		)
		if err != nil {
			logger.Error(fmt.Errorf("failed to insert message for chat %s: %w", chatID, err))
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.Debug(fmt.Sprintf("saved %d messages for chat %s", len(messages), chatID))
	return nil
}

func (r *SavedChatMySQLRepository) DeleteMessages(ctx context.Context, chatID string) error {
	query := `DELETE FROM chat_messages WHERE chat_id = ?`

	_, err := r.db.ExecContext(ctx, query, chatID)
	if err != nil {
		logger.Error(fmt.Errorf("failed to delete messages for chat %s: %w", chatID, err))
		return err
	}

	logger.Debug(fmt.Sprintf("deleted messages for chat %s", chatID))
	return nil
}
