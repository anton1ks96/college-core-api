package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/anton1ks96/college-core-api/internal/config"
	"github.com/anton1ks96/college-core-api/internal/domain"
	"github.com/anton1ks96/college-core-api/pkg/logger"
	"net/http"
)

type AuthServiceImpl struct {
	cfg        *config.Config
	httpClient *http.Client
}

func NewAuthService(cfg *config.Config) *AuthServiceImpl {
	return &AuthServiceImpl{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.AuthService.Timeout,
		},
	}
}

func (s *AuthServiceImpl) ValidateToken(ctx context.Context, token string) (*domain.User, error) {
	url := fmt.Sprintf("%s/api/v1/auth/validate", s.cfg.AuthService.URL)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		logger.Error(fmt.Errorf("failed to create request: %w", err))
		return nil, fmt.Errorf("failed to create auth request")
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error(fmt.Errorf("failed to validate token: %w", err))
		return nil, fmt.Errorf("failed to validate token")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("invalid or expired token")
		}
		return nil, fmt.Errorf("auth service returned status %d", resp.StatusCode)
	}

	var response struct {
		User struct {
			ID       string `json:"id"`
			Username string `json:"username"`
			Role     string `json:"role"`
		} `json:"user"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		logger.Error(fmt.Errorf("failed to decode auth response: %w", err))
		return nil, fmt.Errorf("failed to decode auth response")
	}

	role := "student"
	if response.User.Role != "" {
		role = response.User.Role
	}

	user := &domain.User{
		ID:       response.User.ID,
		Username: response.User.Username,
		Role:     role,
	}

	logger.Debug(fmt.Sprintf("token validated for user %s with role %s", user.Username, user.Role))

	return user, nil
}
