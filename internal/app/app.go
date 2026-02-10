package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anton1ks96/college-core-api/internal/client/llm"
	"github.com/anton1ks96/college-core-api/internal/client/tei"
	"github.com/anton1ks96/college-core-api/internal/config"
	"github.com/anton1ks96/college-core-api/internal/handlers"
	"github.com/anton1ks96/college-core-api/internal/repository"
	"github.com/anton1ks96/college-core-api/internal/server"
	"github.com/anton1ks96/college-core-api/internal/services"
	"github.com/anton1ks96/college-core-api/pkg/database/mysql"
	"github.com/anton1ks96/college-core-api/pkg/database/qdrant"
	"github.com/anton1ks96/college-core-api/pkg/logger"
)

func Run() {
	cfg, err := config.Init()
	if err != nil {
		logger.Fatal(err)
	}

	db, err := mysql.NewClient(cfg)
	if err != nil {
		logger.Fatal(err)
	}
	defer mysql.Close(db)

	qdrantClient, err := qdrant.NewClient(cfg)
	if err != nil {
		logger.Fatal(err)
	}
	defer qdrantClient.Close()

	llmClient := llm.NewClient(cfg)
	teiClient := tei.NewClient(cfg)

	vectorRepo := repository.NewVectorRepository(cfg, qdrantClient)

	ctx := context.Background()
	if err := vectorRepo.EnsureCollection(ctx, 1024); err != nil {
		logger.Fatal(err)
	}

	datasetRepo := repository.NewDatasetRepository(cfg, db)
	fileRepo, err := repository.NewFileRepository(cfg)
	if err != nil {
		logger.Fatal(err)
	}
	topicRepo := repository.NewTopicRepository(cfg, db)
	datasetPermissionRepo := repository.NewDatasetPermissionRepository(cfg, db)
	savedChatRepo := repository.NewSavedChatRepository(cfg, db)

	repos := &services.Repositories{
		Dataset:           datasetRepo,
		File:              fileRepo,
		Topic:             topicRepo,
		DatasetPermission: datasetPermissionRepo,
		SavedChat:         savedChatRepo,
		Vector:            vectorRepo,
	}

	clients := &services.Clients{
		LLM: llmClient,
		TEI: teiClient,
	}

	servicesInstance := services.NewServices(services.Deps{
		Repos:   repos,
		Clients: clients,
		Config:  cfg,
	})

	handler := handlers.NewHandler(servicesInstance, cfg)

	router := handler.Init()

	srv := server.NewServer(cfg, router)

	go func() {
		if err := srv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal(err)
		}
	}()

	logger.Info(fmt.Sprintf("College Core API started - PORT: %s", cfg.Server.Port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		logger.Error(fmt.Errorf("server forced to shutdown: %w", err))
	}

	logger.Info("Server exited")
}
