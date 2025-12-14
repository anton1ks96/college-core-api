package repository

import (
	"context"
	"io"

	"github.com/anton1ks96/college-core-api/internal/domain"
)

type DatasetRepository interface {
	Create(ctx context.Context, dataset *domain.Dataset) error
	GetByID(ctx context.Context, id string) (*domain.Dataset, error)
	GetByUserID(ctx context.Context, userID string, offset, limit int) ([]domain.Dataset, int, error)
	GetByTeacherID(ctx context.Context, teacherID string, offset, limit int) ([]domain.Dataset, int, error)
	GetAll(ctx context.Context, offset, limit int) ([]domain.Dataset, int, error)
	Update(ctx context.Context, dataset *domain.Dataset) error
	Delete(ctx context.Context, id string) error
	UpdateIndexedAt(ctx context.Context, id string) error
	ExistsByUserIDAndTopicID(ctx context.Context, userID, topicID string) (bool, error)
}

type FileRepository interface {
	Upload(ctx context.Context, path string, content io.Reader, contentType string) error
	Download(ctx context.Context, path string) ([]byte, error)
	Delete(ctx context.Context, path string) error
	Exists(ctx context.Context, path string) (bool, error)
}

type TopicRepository interface {
	Create(ctx context.Context, topic *domain.Topic) error
	GetByID(ctx context.Context, id string) (*domain.Topic, error)
	GetByCreatorID(ctx context.Context, creatorID string, offset, limit int) ([]domain.Topic, int, error)
	GetAll(ctx context.Context, offset, limit int) ([]domain.Topic, int, error)
	AddAssignments(ctx context.Context, assignments []domain.TopicAssignment) error
	RemoveAssignment(ctx context.Context, topicID, studentID string) error
	GetAssignmentsByStudentID(ctx context.Context, studentID string) ([]domain.TopicAssignment, error)
	GetAssignmentsWithDetailsByStudentID(ctx context.Context, studentID string) ([]domain.AssignmentWithDetails, error)
	GetAssignmentsByTopicID(ctx context.Context, topicID string) ([]domain.TopicAssignment, error)
	GetAssignmentByID(ctx context.Context, id string) (*domain.TopicAssignment, error)
}

type DatasetPermissionRepository interface {
	GrantPermission(ctx context.Context, permission *domain.DatasetPermission) error
	RevokePermission(ctx context.Context, datasetID, teacherID string) error
	HasPermission(ctx context.Context, datasetID, teacherID string) (bool, error)
	GetAllPermissions(ctx context.Context, offset, limit int) ([]domain.DatasetPermission, int, error)
	GetPermissionsByDatasetID(ctx context.Context, datasetID string) ([]domain.DatasetPermission, error)
}

type SavedChatRepository interface {
	Create(ctx context.Context, chat *domain.SavedChat) error
	GetByID(ctx context.Context, id string) (*domain.SavedChat, error)
	GetByDatasetID(ctx context.Context, datasetID string, offset, limit int) ([]domain.SavedChat, int, error)
	Update(ctx context.Context, chat *domain.SavedChat) error
	Delete(ctx context.Context, id string) error
	GetMessagesByChatID(ctx context.Context, chatID string) ([]domain.ChatMessage, error)
	SaveMessages(ctx context.Context, chatID string, messages []domain.ChatMessage) error
	DeleteMessages(ctx context.Context, chatID string) error
}
