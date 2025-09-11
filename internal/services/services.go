package services

import (
	"context"
	"io"

	"github.com/anton1ks96/college-core-api/internal/config"
	"github.com/anton1ks96/college-core-api/internal/domain"
	"github.com/anton1ks96/college-core-api/internal/repository"
)

type DatasetService interface {
	Create(ctx context.Context, userID, title string, content io.Reader) (*domain.Dataset, error)
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

type Services struct {
	Dataset DatasetService
	Auth    AuthService
	RAG     RAGService
}

type Repositories struct {
	Dataset repository.DatasetRepository
	File    repository.FileRepository
}

type Deps struct {
	Repos  *Repositories
	Config *config.Config
}

func NewServices(deps Deps) *Services {
	authService := NewAuthService(deps.Config)

	return &Services{
		Auth: authService,
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
