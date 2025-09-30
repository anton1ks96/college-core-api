package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/anton1ks96/college-core-api/internal/config"
	"github.com/anton1ks96/college-core-api/internal/domain"
	"github.com/anton1ks96/college-core-api/pkg/logger"
	"github.com/google/uuid"
)

type TopicServiceImpl struct {
	repos      *Repositories
	cfg        *config.Config
	httpClient *http.Client
}

func NewTopicService(repos *Repositories, cfg *config.Config) *TopicServiceImpl {
	return &TopicServiceImpl{
		repos: repos,
		cfg:   cfg,
		httpClient: &http.Client{
			Timeout: cfg.AuthService.Timeout,
		},
	}
}

func (s *TopicServiceImpl) SearchStudents(ctx context.Context, query string) ([]domain.StudentInfo, int, error) {
	url := fmt.Sprintf("%s/api/v1/students/search", s.cfg.AuthService.URL)

	requestBody := domain.SearchStudentsRequest{
		Query: query,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error(fmt.Errorf("failed to create request: %w", err))
		return nil, 0, fmt.Errorf("failed to create search request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Token", s.cfg.AuthService.InternalToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error(fmt.Errorf("failed to search students: %w", err))
		return nil, 0, fmt.Errorf("failed to search students")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("auth service returned status %d", resp.StatusCode)
	}

	var response domain.SearchStudentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		logger.Error(fmt.Errorf("failed to decode search response: %w", err))
		return nil, 0, fmt.Errorf("failed to decode search response")
	}

	return response.Students, response.Total, nil
}

func (s *TopicServiceImpl) CreateTopic(ctx context.Context, userID, userName, title, description string, students []domain.StudentInfo) (*domain.Topic, error) {
	topic := &domain.Topic{
		Title:       title,
		Description: description,
		CreatedBy:   userName,
		CreatedByID: userID,
	}

	if err := s.repos.Topic.Create(ctx, topic); err != nil {
		return nil, fmt.Errorf("failed to create topic: %w", err)
	}

	if len(students) > 0 {
		assignments := make([]domain.TopicAssignment, 0, len(students))
		now := time.Now()

		for _, student := range students {
			assignments = append(assignments, domain.TopicAssignment{
				ID:          uuid.New().String(),
				TopicID:     topic.ID,
				StudentID:   student.ID,
				StudentName: student.Username,
				AssignedBy:  userName,
				AssignedAt:  now,
			})
		}

		if err := s.repos.Topic.AddAssignments(ctx, assignments); err != nil {
			return nil, fmt.Errorf("failed to assign students: %w", err)
		}
	}

	return topic, nil
}

func (s *TopicServiceImpl) GetMyTopics(ctx context.Context, userID string, page, limit int) ([]domain.Topic, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	topics, total, err := s.repos.Topic.GetByCreatorID(ctx, userID, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get topics: %w", err)
	}

	return topics, total, nil
}

func (s *TopicServiceImpl) GetAssignedTopics(ctx context.Context, studentID string) ([]domain.AssignedTopicResponse, error) {
	details, err := s.repos.Topic.GetAssignmentsWithDetailsByStudentID(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assignments: %w", err)
	}

	result := make([]domain.AssignedTopicResponse, 0, len(details))

	for _, detail := range details {
		result = append(result, domain.AssignedTopicResponse{
			ID: detail.AssignmentID,
			Topic: domain.TopicResponse{
				ID:          detail.TopicID,
				Title:       detail.TopicTitle,
				Description: detail.Description,
				CreatedBy:   detail.CreatedBy,
				CreatedAt:   detail.TopicCreated,
				UpdatedAt:   detail.TopicUpdated,
			},
			AssignmentID: detail.AssignmentID,
			AssignedAt:   detail.AssignedAt,
			HasDataset:   detail.HasDataset,
		})
	}

	return result, nil
}

func (s *TopicServiceImpl) AddStudents(ctx context.Context, topicID, userID, userName string, students []domain.StudentInfo) error {
	topic, err := s.repos.Topic.GetByID(ctx, topicID)
	if err != nil {
		return err
	}

	if topic.CreatedByID != userID {
		return fmt.Errorf("access denied: only topic creator can add students")
	}

	if len(students) == 0 {
		return fmt.Errorf("student list cannot be empty")
	}

	existingAssignments, err := s.repos.Topic.GetAssignmentsByTopicID(ctx, topicID)
	if err != nil {
		return fmt.Errorf("failed to get existing assignments: %w", err)
	}

	existingStudents := make(map[string]bool)
	for _, assignment := range existingAssignments {
		existingStudents[assignment.StudentID] = true
	}

	assignments := make([]domain.TopicAssignment, 0)
	now := time.Now()

	for _, student := range students {
		if existingStudents[student.ID] {
			continue
		}

		id, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("failed to generate UUID v7: %w", err)
		}

		assignments = append(assignments, domain.TopicAssignment{
			ID:          id.String(),
			TopicID:     topicID,
			StudentID:   student.ID,
			StudentName: student.Username,
			AssignedBy:  userName,
			AssignedAt:  now,
		})
	}

	if len(assignments) == 0 {
		return fmt.Errorf("all specified students are already assigned")
	}

	if err := s.repos.Topic.AddAssignments(ctx, assignments); err != nil {
		return fmt.Errorf("failed to add students: %w", err)
	}

	return nil
}

func (s *TopicServiceImpl) GetTopicStudents(ctx context.Context, topicID, userID string) ([]domain.TopicStudentResponse, error) {
	topic, err := s.repos.Topic.GetByID(ctx, topicID)
	if err != nil {
		return nil, err
	}

	if topic.CreatedByID != userID {
		return nil, fmt.Errorf("access denied: only topic creator can view students")
	}

	assignments, err := s.repos.Topic.GetAssignmentsByTopicID(ctx, topicID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assignments: %w", err)
	}

	result := make([]domain.TopicStudentResponse, 0, len(assignments))

	for _, assignment := range assignments {
		result = append(result, domain.TopicStudentResponse{
			ID: assignment.ID,
			Student: domain.StudentInfo{
				ID:       assignment.StudentID,
				Username: assignment.StudentName,
			},
			AssignedAt: assignment.AssignedAt,
		})
	}

	return result, nil
}
