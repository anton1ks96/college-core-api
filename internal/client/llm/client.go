package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/anton1ks96/college-core-api/internal/config"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	model      string
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: cfg.LLM.Timeout,
		},
		baseURL: cfg.LLM.URL,
		model:   cfg.LLM.Model,
	}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

type StreamChunk struct {
	ID      string         `json:"id"`
	Choices []StreamChoice `json:"choices"`
}

type StreamChoice struct {
	Index        int          `json:"index"`
	Delta        MessageDelta `json:"delta"`
	FinishReason *string      `json:"finish_reason"`
}

type MessageDelta struct {
	Role             string `json:"role,omitempty"`
	Content          string `json:"content,omitempty"`
	ReasoningContent string `json:"reasoning_content,omitempty"`
}

func (c *Client) ChatCompletionStream(ctx context.Context, messages []Message, temperature float64, maxTokens int) (<-chan StreamChunk, <-chan error) {
	chunks := make(chan StreamChunk)
	errc := make(chan error, 1)

	go func() {
		defer close(chunks)
		defer close(errc)

		reqBody := chatRequest{
			Model:       c.model,
			Messages:    messages,
			Stream:      true,
			Temperature: temperature,
			MaxTokens:   maxTokens,
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			errc <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		url := fmt.Sprintf("%s/v1/chat/completions", c.baseURL)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
		if err != nil {
			errc <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			errc <- fmt.Errorf("failed to send request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			errc <- fmt.Errorf("llm returned status %d: %s", resp.StatusCode, string(body))
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			if data == "[DONE]" {
				return
			}

			var chunk StreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				errc <- fmt.Errorf("failed to decode chunk: %w", err)
				return
			}

			select {
			case chunks <- chunk:
			case <-ctx.Done():
				errc <- ctx.Err()
				return
			}
		}

		if err := scanner.Err(); err != nil {
			errc <- fmt.Errorf("stream read error: %w", err)
		}
	}()

	return chunks, errc
}
