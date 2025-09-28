package services

import (
	"context"
	"io"

	"github.com/anton1ks96/college-core-api/internal/config"
	"github.com/anton1ks96/college-core-api/internal/domain"
	"github.com/anton1ks96/college-core-api/internal/repository"
)

type DatasetService interface {
	Create(ctx context.Context, userID, title, assignmentID string, content io.Reader) (*domain.Dataset, error)
	GetByID(ctx context.Context, datasetID, userID string, role string) (*domain.DatasetResponse, error)
	GetList(ctx context.Context, userID string, role string, page, limit int) (*domain.DatasetListResponse, error)
	Update(ctx context.Context, datasetID, userID, title string, content *string) (*domain.Dataset, error)
	Delete(ctx context.Context, datasetID, userID string) error
	AskQuestion(ctx context.Context, datasetID, userID, role, question string) (*domain.AskResponse, error)
	Reindex(ctx context.Context, datasetID, userID string) (*domain.IndexResponse, error)
}

type AuthService interface {
	ValidateToken(ctx context.Context, token string) (*domain.User, error)
}

type RAGService interface {
	IndexDataset(ctx context.Context, datasetID string, title, content string) (int, error)
	AskQuestion(ctx context.Context, datasetID string, question string) (*domain.AskResponse, error)
}

type TopicService interface {
	SearchStudents(ctx context.Context, query string) ([]domain.StudentInfo, int, error)
	CreateTopic(ctx context.Context, userID, title, description string, studentIDs []string) (*domain.Topic, error)
	GetMyTopics(ctx context.Context, userID string, page, limit int) ([]domain.Topic, int, error)
	GetAssignedTopics(ctx context.Context, studentID string) ([]domain.AssignedTopicResponse, error)
	AddStudents(ctx context.Context, topicID, userID string, studentIDs []string) error
	GetTopicStudents(ctx context.Context, topicID, userID string) ([]domain.TopicStudentResponse, error)
}

type Services struct {
	Dataset DatasetService
	Auth    AuthService
	RAG     RAGService
	Topic   TopicService
}

type Repositories struct {
	Dataset repository.DatasetRepository
	File    repository.FileRepository
	Topic   repository.TopicRepository
}

type Deps struct {
	Repos  *Repositories
	Config *config.Config
}

func NewServices(deps Deps) *Services {
	authService := NewAuthService(deps.Config)
	ragService := NewRAGService(deps.Config)
	datasetService := NewDatasetService(deps.Repos, ragService, deps.Config)
	topicService := NewTopicService(deps.Repos, deps.Config)

	return &Services{
		Dataset: datasetService,
		Auth:    authService,
		RAG:     ragService,
		Topic:   topicService,
	}
}

func checkPermission(userID, ownerID, role string) bool {
	if role == "teacher" {
		return true
	}
	return userID == ownerID
}

func checkEditPermission(userID, ownerID string) bool {
	return userID == ownerID
}
