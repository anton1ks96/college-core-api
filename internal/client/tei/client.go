package tei

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/anton1ks96/college-core-api/internal/config"
)

type Client struct {
	httpClient    *http.Client
	embeddingsURL string
	rerankerURL   string
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: cfg.TEI.Timeout,
		},
		embeddingsURL: cfg.TEI.EmbeddingsURL,
		rerankerURL:   cfg.TEI.RerankerURL,
	}
}

type RerankResult struct {
	Index int     `json:"index"`
	Score float64 `json:"score"`
}

func (c *Client) Embed(ctx context.Context, input string) ([]float32, error) {
	vectors, err := c.EmbedBatch(ctx, []string{input})
	if err != nil {
		return nil, err
	}
	if len(vectors) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}
	return vectors[0], nil
}

const embedBatchSize = 32

func (c *Client) EmbedBatch(ctx context.Context, inputs []string) ([][]float32, error) {
	result := make([][]float32, 0, len(inputs))

	for start := 0; start < len(inputs); start += embedBatchSize {
		end := min(start+embedBatchSize, len(inputs))

		vectors, err := c.embedBatchChunk(ctx, inputs[start:end])
		if err != nil {
			return nil, err
		}

		result = append(result, vectors...)
	}

	return result, nil
}

func (c *Client) embedBatchChunk(ctx context.Context, inputs []string) ([][]float32, error) {
	body := struct {
		Inputs []string `json:"inputs"`
	}{Inputs: inputs}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embed request: %w", err)
	}

	url := fmt.Sprintf("%s/embed", c.embeddingsURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create embed request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send embed request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tei embeddings returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var vectors [][]float32
	if err := json.NewDecoder(resp.Body).Decode(&vectors); err != nil {
		return nil, fmt.Errorf("failed to decode embed response: %w", err)
	}

	return vectors, nil
}

func (c *Client) Rerank(ctx context.Context, query string, texts []string) ([]RerankResult, error) {
	body := struct {
		Query string   `json:"query"`
		Texts []string `json:"texts"`
	}{Query: query, Texts: texts}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal rerank request: %w", err)
	}

	url := fmt.Sprintf("%s/rerank", c.rerankerURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create rerank request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send rerank request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tei reranker returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var results []RerankResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode rerank response: %w", err)
	}

	return results, nil
}
