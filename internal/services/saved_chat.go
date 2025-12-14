package services

import (
	"bytes"
	"context"
	"fmt"

	"github.com/anton1ks96/college-core-api/internal/domain"
	"github.com/anton1ks96/college-core-api/pkg/logger"
)

type SavedChatServiceImpl struct {
	repos *Repositories
}

func NewSavedChatService(repos *Repositories) *SavedChatServiceImpl {
	return &SavedChatServiceImpl{
		repos: repos,
	}
}

func (s *SavedChatServiceImpl) checkChatAccess(ctx context.Context, datasetID, userID, role string) error {
	if role != "teacher" && role != "admin" {
		return fmt.Errorf("access denied: insufficient permissions")
	}

	_, err := s.repos.Dataset.GetByID(ctx, datasetID)
	if err != nil {
		return fmt.Errorf("dataset not found")
	}

	if role == "admin" {
		return nil
	}

	hasPermission, err := s.repos.DatasetPermission.HasPermission(ctx, datasetID, userID)
	if err != nil {
		return fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return fmt.Errorf("access denied")
	}

	return nil
}

func (s *SavedChatServiceImpl) CreateChat(ctx context.Context, datasetID, userID, username, role, title string, messages []domain.ChatMessageInput) (*domain.SavedChatResponse, error) {
	if err := s.checkChatAccess(ctx, datasetID, userID, role); err != nil {
		return nil, err
	}

	chat := &domain.SavedChat{
		DatasetID: datasetID,
		Title:     title,
		CreatedBy: username,
		UserID:    userID,
	}

	if err := s.repos.SavedChat.Create(ctx, chat); err != nil {
		return nil, fmt.Errorf("failed to create chat: %w", err)
	}

	chatMessages := make([]domain.ChatMessage, len(messages))
	for i, msg := range messages {
		chatMessages[i] = domain.ChatMessage{
			Question:  msg.Question,
			Answer:    msg.Answer,
			Citations: msg.Citations,
		}
	}

	if err := s.repos.SavedChat.SaveMessages(ctx, chat.ID, chatMessages); err != nil {
		_ = s.repos.SavedChat.Delete(ctx, chat.ID)
		return nil, fmt.Errorf("failed to save messages: %w", err)
	}

	savedMessages, err := s.repos.SavedChat.GetMessagesByChatID(ctx, chat.ID)
	if err != nil {
		logger.Error(fmt.Errorf("failed to get saved messages: %w", err))
	}

	return &domain.SavedChatResponse{
		ID:        chat.ID,
		DatasetID: chat.DatasetID,
		Title:     chat.Title,
		CreatedBy: chat.CreatedBy,
		UserID:    chat.UserID,
		Messages:  savedMessages,
		CreatedAt: chat.CreatedAt,
		UpdatedAt: chat.UpdatedAt,
	}, nil
}

func (s *SavedChatServiceImpl) GetChatsByDataset(ctx context.Context, datasetID, userID, role string, page, limit int) (*domain.SavedChatListResponse, error) {
	if err := s.checkChatAccess(ctx, datasetID, userID, role); err != nil {
		return nil, err
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	chats, total, err := s.repos.SavedChat.GetByDatasetID(ctx, datasetID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get chats: %w", err)
	}

	return &domain.SavedChatListResponse{
		Chats: chats,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

func (s *SavedChatServiceImpl) GetChat(ctx context.Context, chatID, userID, role string) (*domain.SavedChatResponse, error) {
	chat, err := s.repos.SavedChat.GetByID(ctx, chatID)
	if err != nil {
		return nil, err
	}

	if err := s.checkChatAccess(ctx, chat.DatasetID, userID, role); err != nil {
		return nil, err
	}

	messages, err := s.repos.SavedChat.GetMessagesByChatID(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return &domain.SavedChatResponse{
		ID:        chat.ID,
		DatasetID: chat.DatasetID,
		Title:     chat.Title,
		CreatedBy: chat.CreatedBy,
		UserID:    chat.UserID,
		Messages:  messages,
		CreatedAt: chat.CreatedAt,
		UpdatedAt: chat.UpdatedAt,
	}, nil
}

func (s *SavedChatServiceImpl) UpdateChat(ctx context.Context, chatID, userID, role, title string, messages []domain.ChatMessageInput) (*domain.SavedChatResponse, error) {
	chat, err := s.repos.SavedChat.GetByID(ctx, chatID)
	if err != nil {
		return nil, err
	}

	if role != "admin" && chat.UserID != userID {
		return nil, fmt.Errorf("access denied: only creator or admin can update chat")
	}

	if err := s.checkChatAccess(ctx, chat.DatasetID, userID, role); err != nil {
		return nil, err
	}

	chat.Title = title
	if err := s.repos.SavedChat.Update(ctx, chat); err != nil {
		return nil, fmt.Errorf("failed to update chat: %w", err)
	}

	chatMessages := make([]domain.ChatMessage, len(messages))
	for i, msg := range messages {
		chatMessages[i] = domain.ChatMessage{
			Question:  msg.Question,
			Answer:    msg.Answer,
			Citations: msg.Citations,
		}
	}

	if err := s.repos.SavedChat.SaveMessages(ctx, chatID, chatMessages); err != nil {
		return nil, fmt.Errorf("failed to save messages: %w", err)
	}

	savedMessages, err := s.repos.SavedChat.GetMessagesByChatID(ctx, chatID)
	if err != nil {
		logger.Error(fmt.Errorf("failed to get saved messages: %w", err))
	}

	return &domain.SavedChatResponse{
		ID:        chat.ID,
		DatasetID: chat.DatasetID,
		Title:     chat.Title,
		CreatedBy: chat.CreatedBy,
		UserID:    chat.UserID,
		Messages:  savedMessages,
		CreatedAt: chat.CreatedAt,
		UpdatedAt: chat.UpdatedAt,
	}, nil
}

func (s *SavedChatServiceImpl) DeleteChat(ctx context.Context, chatID, userID, role string) error {
	chat, err := s.repos.SavedChat.GetByID(ctx, chatID)
	if err != nil {
		return err
	}

	if role != "admin" && chat.UserID != userID {
		return fmt.Errorf("access denied: only creator or admin can delete chat")
	}

	if err := s.repos.SavedChat.Delete(ctx, chatID); err != nil {
		return fmt.Errorf("failed to delete chat: %w", err)
	}

	logger.Info(fmt.Sprintf("chat %s deleted by user %s", chatID, userID))
	return nil
}

func (s *SavedChatServiceImpl) DownloadChatMarkdown(ctx context.Context, chatID, userID, role string) ([]byte, string, error) {
	chat, err := s.repos.SavedChat.GetByID(ctx, chatID)
	if err != nil {
		return nil, "", err
	}

	if err := s.checkChatAccess(ctx, chat.DatasetID, userID, role); err != nil {
		return nil, "", err
	}

	dataset, err := s.repos.Dataset.GetByID(ctx, chat.DatasetID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get dataset: %w", err)
	}

	messages, err := s.repos.SavedChat.GetMessagesByChatID(ctx, chatID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get messages: %w", err)
	}

	markdown := generateChatMarkdown(chat, dataset.Title, messages)

	filename := fmt.Sprintf("%s.md", sanitizeFilename(chat.Title))

	return markdown, filename, nil
}

func generateChatMarkdown(chat *domain.SavedChat, datasetTitle string, messages []domain.ChatMessage) []byte {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("# %s\n\n", chat.Title))
	buf.WriteString(fmt.Sprintf("**Dataset:** %s\n", datasetTitle))
	buf.WriteString(fmt.Sprintf("**Author:** %s\n", chat.CreatedBy))
	buf.WriteString(fmt.Sprintf("**Created:** %s\n\n", chat.CreatedAt.Format("2006-01-02 15:04")))
	buf.WriteString("---\n\n")

	for i, msg := range messages {
		buf.WriteString(fmt.Sprintf("## Question %d\n\n", i+1))
		buf.WriteString(fmt.Sprintf("**Q:** %s\n\n", msg.Question))
		buf.WriteString(fmt.Sprintf("**A:** %s\n\n", msg.Answer))
		buf.WriteString("---\n\n")
	}
	return buf.Bytes()
}

func sanitizeFilename(name string) string {
	result := make([]byte, 0, len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == ' ' {
			result = append(result, c)
		}
	}
	if len(result) == 0 {
		return "chat"
	}
	if len(result) > 50 {
		result = result[:50]
	}
	return string(result)
}
