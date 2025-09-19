package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/anton1ks96/college-core-api/internal/config"
	"github.com/anton1ks96/college-core-api/internal/domain"
	"github.com/anton1ks96/college-core-api/pkg/logger"
)

type RAGServiceImpl struct {
	cfg        *config.Config
	httpClient *http.Client
}

func NewRAGService(cfg *config.Config) *RAGServiceImpl {
	return &RAGServiceImpl{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.RAGService.Timeout,
		},
	}
}

func (s *RAGServiceImpl) IndexDataset(ctx context.Context, datasetID string, title, content string) (int, error) {
	url := fmt.Sprintf("%s/index/file", s.cfg.RAGService.URL)

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	if err := writer.WriteField("dataset_id", datasetID); err != nil {
		logger.Error(fmt.Errorf("failed to write dataset_id field: %w", err))
		return 0, err
	}

	if title != "" {
		if err := writer.WriteField("title", title); err != nil {
			logger.Error(fmt.Errorf("failed to write title field: %w", err))
			return 0, err
		}
	}

	if err := writer.WriteField("version", "1"); err != nil {
		logger.Error(fmt.Errorf("failed to write version field: %w", err))
		return 0, err
	}

	if err := writer.WriteField("overwrite", "true"); err != nil {
		logger.Error(fmt.Errorf("failed to write overwrite field: %w", err))
		return 0, err
	}

	fileName := fmt.Sprintf("%s.md", datasetID)
	fileWriter, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		logger.Error(fmt.Errorf("failed to create form file: %w", err))
		return 0, err
	}

	if _, err := io.Copy(fileWriter, bytes.NewReader([]byte(content))); err != nil {
		logger.Error(fmt.Errorf("failed to write file content: %w", err))
		return 0, err
	}

	contentType := writer.FormDataContentType()
	if err := writer.Close(); err != nil {
		logger.Error(fmt.Errorf("failed to close multipart writer: %w", err))
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, &requestBody)
	if err != nil {
		logger.Error(fmt.Errorf("failed to create index request: %w", err))
		return 0, err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+s.cfg.RAGService.Token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error(fmt.Errorf("failed to index dataset in RAG: %w", err))
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		logger.Error(fmt.Errorf("RAG service returned status %d: %s", resp.StatusCode, string(bodyBytes)))
		return 0, fmt.Errorf("RAG service returned status %d", resp.StatusCode)
	}

	var response struct {
		OK     bool `json:"ok"`
		Chunks int  `json:"chunks"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		logger.Error(fmt.Errorf("failed to decode index response: %w", err))
		return 0, err
	}

	if !response.OK {
		return 0, fmt.Errorf("indexing failed")
	}

	logger.Info(fmt.Sprintf("dataset %s indexed with %d chunks", datasetID, response.Chunks))
	return response.Chunks, nil
}

func (s *RAGServiceImpl) AskQuestion(ctx context.Context, datasetID string, question string) (*domain.AskResponse, error) {
	url := fmt.Sprintf("%s/ask", s.cfg.RAGService.URL)

	payload := map[string]interface{}{
		"dataset_id":      datasetID,
		"question":        question,
		"k":               6,
		"min_score":       0.0,
		"max_ctx_chars":   8000,
		"use_reranking":   true,
		"debug_reranking": false,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		logger.Error(fmt.Errorf("failed to marshal ask request: %w", err))
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error(fmt.Errorf("failed to create ask request: %w", err))
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.cfg.RAGService.Token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error(fmt.Errorf("failed to ask question to RAG: %w", err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RAG service returned status %d", resp.StatusCode)
	}

	var ragResponse struct {
		Answer    string `json:"answer"`
		Citations []struct {
			ChunkID          int     `json:"chunk_id"`
			Score            float64 `json:"score"`
			OriginalScore    float64 `json:"original_score,omitempty"`
			ScoreImprovement float64 `json:"score_improvement,omitempty"`
		} `json:"citations"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ragResponse); err != nil {
		logger.Error(fmt.Errorf("failed to decode ask response: %w", err))
		return nil, err
	}

	response := &domain.AskResponse{
		Answer: ragResponse.Answer,
	}

	for _, citation := range ragResponse.Citations {
		response.Citations = append(response.Citations, domain.Citation{
			ChunkID:          citation.ChunkID,
			Score:            citation.Score,
			OriginalScore:    citation.OriginalScore,
			ScoreImprovement: citation.ScoreImprovement,
		})
	}

	logger.Debug(fmt.Sprintf("question answered for dataset %s with %d citations", datasetID, len(response.Citations)))
	return response, nil
}
