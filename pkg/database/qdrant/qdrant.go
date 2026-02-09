package qdrant

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/anton1ks96/college-core-api/internal/config"
	"github.com/anton1ks96/college-core-api/pkg/logger"
	"github.com/qdrant/go-client/qdrant"
)

type Qdrant struct {
	Client *qdrant.Client
}

func NewClient(cfg *config.Config) (*Qdrant, error) {
	port, err := strconv.Atoi(cfg.Qdrant.Port)
	if err != nil {
		return nil, fmt.Errorf("invalid qdrant port %q: %w", cfg.Qdrant.Port, err)
	}

	qCfg := &qdrant.Config{
		Host:   cfg.Qdrant.Host,
		Port:   port,
		APIKey: cfg.Qdrant.ApiKey,
		UseTLS: false,
	}

	client, err := qdrant.NewClient(qCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create qdrant client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.HealthCheck(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("qdrant health check failed: %w", err)
	}

	logger.Info(fmt.Sprintf("Connected to Qdrant at %s:%d", cfg.Qdrant.Host, port))

	return &Qdrant{Client: client}, nil
}

func (q *Qdrant) Close() error {
	return q.Client.Close()
}
