package services

import (
	"context"
	"io"

	"github.com/anton1ks96/college-core-api/internal/client/llm"
	"github.com/anton1ks96/college-core-api/internal/client/tei"
	"github.com/anton1ks96/college-core-api/internal/config"
	"github.com/anton1ks96/college-core-api/internal/domain"
	"github.com/anton1ks96/college-core-api/internal/repository"
)

type DatasetService interface {
	Create(ctx context.Context, userID, username, title, assignmentID string, content io.Reader) (*domain.Dataset, error)
	GetByID(ctx context.Context, datasetID, userID string, role string) (*domain.DatasetResponse, error)
	GetList(ctx context.Context, userID string, role string, page, limit int) (*domain.DatasetListResponse, error)
	Update(ctx context.Context, datasetID, userID, title string, content *string) (*domain.Dataset, error)
	Delete(ctx context.Context, datasetID, userID, role string) error
	AskQuestion(ctx context.Context, datasetID, userID, role, question string) (<-chan domain.AskEvent, error)
	Reindex(ctx context.Context, datasetID, userID string) (*domain.IndexResponse, error)
	SetTag(ctx context.Context, datasetID, userID, role string, tag *string) error
	SearchByTag(ctx context.Context, userID, role, tag string, page, limit int) (*domain.DatasetListResponse, error)
}

type AuthService interface {
	ValidateToken(ctx context.Context, token string) (*domain.User, error)
}

type TopicService interface {
	SearchStudents(ctx context.Context, query string) ([]domain.StudentInfo, int, error)
	SearchTeachers(ctx context.Context, query string) ([]domain.StudentInfo, int, error)
	CreateTopic(ctx context.Context, userID, userName, title, description string, students []domain.StudentInfo) (*domain.Topic, error)
	GetMyTopics(ctx context.Context, userID string, page, limit int) ([]domain.Topic, int, error)
	GetAllTopics(ctx context.Context, page, limit int) ([]domain.Topic, int, error)
	GetAssignedTopics(ctx context.Context, studentID string) ([]domain.AssignedTopicResponse, error)
	AddStudents(ctx context.Context, topicID, userID, userName, role string, students []domain.StudentInfo) error
	GetTopicStudents(ctx context.Context, topicID, userID, role string) ([]domain.TopicStudentResponse, error)
	RemoveStudent(ctx context.Context, topicID, studentID, userID, role string) error
}

type DatasetPermissionService interface {
	GrantDatasetPermission(ctx context.Context, datasetID, teacherID, teacherName, grantedBy string) (string, error)
	RevokeDatasetPermission(ctx context.Context, datasetID, teacherID string) error
	GetAllPermissions(ctx context.Context, page, limit int) ([]domain.DatasetPermission, int, error)
	GetDatasetPermissions(ctx context.Context, datasetID string) ([]domain.DatasetPermission, error)
}

type SavedChatService interface {
	CreateChat(ctx context.Context, datasetID, userID, username, role, title string, messages []domain.ChatMessageInput) (*domain.SavedChatResponse, error)
	GetChatsByDataset(ctx context.Context, datasetID, userID, role string, page, limit int) (*domain.SavedChatListResponse, error)
	GetChat(ctx context.Context, chatID, userID, role string) (*domain.SavedChatResponse, error)
	UpdateChat(ctx context.Context, chatID, userID, role, title string, messages []domain.ChatMessageInput) (*domain.SavedChatResponse, error)
	DeleteChat(ctx context.Context, chatID, userID, role string) error
	DownloadChatMarkdown(ctx context.Context, chatID, userID, role string) ([]byte, string, error)
}

type Services struct {
	Dataset           DatasetService
	Auth              AuthService
	Topic             TopicService
	DatasetPermission DatasetPermissionService
	SavedChat         SavedChatService
}

type Repositories struct {
	Dataset           repository.DatasetRepository
	File              repository.FileRepository
	Topic             repository.TopicRepository
	DatasetPermission repository.DatasetPermissionRepository
	SavedChat         repository.SavedChatRepository
	Vector            repository.VectorRepository
}

type Clients struct {
	LLM *llm.Client
	TEI *tei.Client
}

type Deps struct {
	Repos   *Repositories
	Clients *Clients
	Config  *config.Config
}

func NewServices(deps Deps) *Services {
	authService := NewAuthService(deps.Config)
	datasetService := NewDatasetService(deps.Repos, deps.Clients, deps.Config)
	topicService := NewTopicService(deps.Repos, deps.Config)
	datasetPermissionService := NewDatasetPermissionService(deps.Repos)
	savedChatService := NewSavedChatService(deps.Repos)

	return &Services{
		Dataset:           datasetService,
		Auth:              authService,
		Topic:             topicService,
		DatasetPermission: datasetPermissionService,
		SavedChat:         savedChatService,
	}
}

func checkEditPermission(userID, ownerID string) bool {
	return userID == ownerID
}
