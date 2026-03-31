package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/anton1ks96/college-core-api/internal/client/llm"
	"github.com/anton1ks96/college-core-api/internal/config"
	"github.com/anton1ks96/college-core-api/internal/domain"
	"github.com/anton1ks96/college-core-api/internal/rag"
	"github.com/anton1ks96/college-core-api/pkg/logger"
	"github.com/google/uuid"
)

const datasetVersion = 1

type DatasetServiceImpl struct {
	repos   *Repositories
	clients *Clients
	cfg     *config.Config
}

func NewDatasetService(repos *Repositories, clients *Clients, cfg *config.Config) *DatasetServiceImpl {
	return &DatasetServiceImpl{
		repos:   repos,
		clients: clients,
		cfg:     cfg,
	}
}

func (s *DatasetServiceImpl) Create(ctx context.Context, userID, username, title, assignmentID string, content io.Reader) (*domain.Dataset, error) {
	assignment, err := s.repos.Topic.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		return nil, fmt.Errorf("assignment not found")
	}

	if assignment.StudentID != userID {
		return nil, fmt.Errorf("access denied: assignment belongs to another student")
	}

	exists, err := s.repos.Dataset.ExistsByUserIDAndTopicID(ctx, userID, assignment.TopicID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing dataset: %w", err)
	}

	if exists {
		return nil, fmt.Errorf("dataset already exists for this topic")
	}

	buf := new(bytes.Buffer)
	size, err := io.Copy(buf, content)
	if err != nil {
		return nil, fmt.Errorf("failed to read content: %w", err)
	}

	if size > s.cfg.Limits.MaxFileSize {
		return nil, fmt.Errorf("file size exceeds limit: %d > %d bytes", size, s.cfg.Limits.MaxFileSize)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUID v7: %w", err)
	}

	dataset := &domain.Dataset{
		ID:           id.String(),
		UserID:       userID,
		Author:       username,
		Title:        title,
		FilePath:     fmt.Sprintf("students/%s/%s/dataset.md", userID, uuid.New().String()),
		TopicID:      &assignment.TopicID,
		AssignmentID: &assignmentID,
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

	return dataset, nil
}

func (s *DatasetServiceImpl) hasReadAccess(ctx context.Context, datasetID, userID, ownerID, role string) (bool, error) {
	if role == "admin" {
		return true, nil
	}
	if userID == ownerID {
		return true, nil
	}
	if role == "teacher" {
		hasPermission, err := s.repos.DatasetPermission.HasPermission(ctx, datasetID, userID)
		if err != nil {
			return false, fmt.Errorf("failed to check permission: %w", err)
		}
		return hasPermission, nil
	}
	return false, nil
}

func (s *DatasetServiceImpl) GetByID(ctx context.Context, datasetID, userID string, role string) (*domain.DatasetResponse, error) {
	dataset, err := s.repos.Dataset.GetByID(ctx, datasetID)
	if err != nil {
		return nil, err
	}

	hasAccess, err := s.hasReadAccess(ctx, datasetID, userID, dataset.UserID, role)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
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
		Author:    dataset.Author,
		UserID:    dataset.UserID,
		CreatedAt: dataset.CreatedAt,
		UpdatedAt: dataset.UpdatedAt,
		IndexedAt: dataset.IndexedAt,
		Tag:       dataset.Tag,
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

	if role == "admin" {
		datasets, total, err = s.repos.Dataset.GetAll(ctx, offset, limit)
	} else if role == "teacher" {
		datasets, total, err = s.repos.Dataset.GetByTeacherID(ctx, userID, offset, limit)
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

func (s *DatasetServiceImpl) Delete(ctx context.Context, datasetID, userID, role string) error {
	dataset, err := s.repos.Dataset.GetByID(ctx, datasetID)
	if err != nil {
		return err
	}

	hasAccess, err := s.hasReadAccess(ctx, datasetID, userID, dataset.UserID, role)
	if err != nil {
		return err
	}
	if !hasAccess {
		return fmt.Errorf("access denied")
	}

	err = s.repos.Dataset.Delete(ctx, datasetID)
	if err != nil {
		return fmt.Errorf("failed to delete dataset: %w", err)
	}

	logger.Info(fmt.Sprintf("dataset %s deleted by user %s (role: %s)", datasetID, userID, role))
	return nil
}

func (s *DatasetServiceImpl) AskQuestion(ctx context.Context, datasetID, userID, role, question string) (<-chan domain.AskEvent, error) {
	dataset, err := s.repos.Dataset.GetByID(ctx, datasetID)
	if err != nil {
		return nil, err
	}

	hasAccess, err := s.hasReadAccess(ctx, datasetID, userID, dataset.UserID, role)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, fmt.Errorf("access denied")
	}

	if dataset.IndexedAt == nil {
		return nil, fmt.Errorf("dataset is not indexed yet, please wait")
	}

	events := make(chan domain.AskEvent)

	go func() {
		defer close(events)

		queryVector, err := s.clients.TEI.Embed(ctx, question)
		if err != nil {
			s.sendEvent(ctx, events, domain.AskEvent{Type: "error", Error: "failed to embed question"})
			return
		}

		hits, err := s.repos.Vector.Search(ctx, datasetID, datasetVersion, queryVector, uint64(s.cfg.RAG.SearchTopK))
		if err != nil {
			s.sendEvent(ctx, events, domain.AskEvent{Type: "error", Error: "failed to search vectors"})
			return
		}

		if len(hits) == 0 {
			s.sendEvent(ctx, events, domain.AskEvent{Type: "error", Error: "no relevant content found"})
			return
		}

		texts := make([]string, len(hits))
		for i, h := range hits {
			texts[i] = h.Text
		}

		reranked, err := s.clients.TEI.Rerank(ctx, question, texts)
		if err != nil {
			s.sendEvent(ctx, events, domain.AskEvent{Type: "error", Error: "failed to rerank"})
			return
		}

		sort.Slice(reranked, func(i, j int) bool {
			return reranked[i].Score > reranked[j].Score
		})

		topN := s.cfg.RAG.RerankTopN
		if len(reranked) < topN {
			topN = len(reranked)
		}

		contextChunks := make([]string, topN)
		citations := make([]domain.Citation, topN)
		for i := 0; i < topN; i++ {
			idx := reranked[i].Index
			contextChunks[i] = hits[idx].Text
			citations[i] = domain.Citation{
				ChunkID:          hits[idx].ChunkID,
				Score:            reranked[i].Score,
				OriginalScore:    float64(hits[idx].Score),
				ScoreImprovement: reranked[i].Score - float64(hits[idx].Score),
			}
		}

		messages := []llm.Message{
			{Role: "system", Content: rag.SystemPrompt},
			{Role: "user", Content: rag.BuildUserPrompt(question, contextChunks)},
		}

		chunks, errc := s.clients.LLM.ChatCompletionStream(ctx, messages, s.cfg.RAG.LLMTemperature, s.cfg.RAG.LLMMaxTokens)

		for chunk := range chunks {
			for _, choice := range chunk.Choices {
				if choice.Delta.ReasoningContent != "" {
					if !s.sendEvent(ctx, events, domain.AskEvent{Type: "thinking", Delta: choice.Delta.ReasoningContent}) {
						return
					}
				}
				if choice.Delta.Content != "" {
					if !s.sendEvent(ctx, events, domain.AskEvent{Type: "delta", Delta: choice.Delta.Content}) {
						return
					}
				}
			}
		}

		if err := <-errc; err != nil {
			s.sendEvent(ctx, events, domain.AskEvent{Type: "error", Error: "llm generation failed"})
			return
		}

		s.sendEvent(ctx, events, domain.AskEvent{Type: "citations", Citations: citations})
		s.sendEvent(ctx, events, domain.AskEvent{Type: "done"})
	}()

	return events, nil
}

// sendEvent отправляет событие в канал с проверкой отмены контекста
// Возвращает false если контекст отменён (клиент отключился)
func (s *DatasetServiceImpl) sendEvent(ctx context.Context, ch chan<- domain.AskEvent, event domain.AskEvent) bool {
	select {
	case ch <- event:
		return true
	case <-ctx.Done():
		return false
	}
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
		return nil, fmt.Errorf("failed to download dataset: %w", err)
	}

	normalized := rag.NormalizeMarkdown(string(content))

	assignmentID := ""
	if dataset.AssignmentID != nil {
		assignmentID = *dataset.AssignmentID
	}

	docs := rag.ChunkStudentMarkdown(normalized, dataset.UserID, assignmentID, datasetVersion, dataset.Title)
	if len(docs) == 0 {
		return nil, fmt.Errorf("dataset content is empty or has no sections")
	}

	texts := make([]string, len(docs))
	for i, doc := range docs {
		texts[i] = doc.PageContent
	}

	vectors, err := s.clients.TEI.EmbedBatch(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings: %w", err)
	}

	if err := s.repos.Vector.DeleteByDatasetID(ctx, datasetID); err != nil {
		return nil, fmt.Errorf("failed to delete old vectors: %w", err)
	}

	chunks := make([]domain.ChunkData, len(docs))
	for i, doc := range docs {
		chunks[i] = domain.ChunkData{
			Index: doc.Metadata.ChunkID,
			Text:  doc.PageContent,
		}
	}

	count, err := s.repos.Vector.UpsertChunks(ctx, datasetID, datasetVersion, dataset.Title, chunks, vectors)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert vectors: %w", err)
	}

	if err := s.repos.Dataset.UpdateIndexedAt(ctx, datasetID); err != nil {
		return nil, fmt.Errorf("failed to update indexed_at: %w", err)
	}

	logger.Info(fmt.Sprintf("dataset %s reindexed: %d chunks", datasetID, count))

	return &domain.IndexResponse{
		Success: true,
		Chunks:  count,
		Message: fmt.Sprintf("Successfully indexed %d chunks", count),
	}, nil
}

func (s *DatasetServiceImpl) SetTag(ctx context.Context, datasetID, userID, role string, tag *string) error {
	dataset, err := s.repos.Dataset.GetByID(ctx, datasetID)
	if err != nil {
		return err
	}

	hasAccess, err := s.hasReadAccess(ctx, datasetID, userID, dataset.UserID, role)
	if err != nil {
		return err
	}
	if !hasAccess {
		return fmt.Errorf("access denied")
	}

	var normalizedTag *string
	if tag != nil {
		t := strings.ToLower(strings.TrimSpace(*tag))
		normalizedTag = &t
	}

	return s.repos.Dataset.SetTag(ctx, datasetID, normalizedTag)
}

func (s *DatasetServiceImpl) SearchByTag(ctx context.Context, userID, role, tag string, page, limit int) (*domain.DatasetListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	if tag == "*" {
		return s.GetList(ctx, userID, role, page, limit)
	}

	normalizedTag := strings.ToLower(strings.TrimSpace(tag))

	offset := (page - 1) * limit
	var datasets []domain.Dataset
	var total int
	var err error

	if role == "admin" {
		datasets, total, err = s.repos.Dataset.GetByTagAll(ctx, normalizedTag, offset, limit)
	} else {
		datasets, total, err = s.repos.Dataset.GetByTagAndTeacherID(ctx, normalizedTag, userID, offset, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to search datasets by tag: %w", err)
	}

	return &domain.DatasetListResponse{
		Datasets: datasets,
		Total:    total,
		Page:     page,
		Limit:    limit,
	}, nil
}
