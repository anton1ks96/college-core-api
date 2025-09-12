package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anton1ks96/college-core-api/internal/config"
	"github.com/anton1ks96/college-core-api/internal/handlers"
	"github.com/anton1ks96/college-core-api/internal/repository"
	"github.com/anton1ks96/college-core-api/internal/server"
	"github.com/anton1ks96/college-core-api/internal/services"
	"github.com/anton1ks96/college-core-api/pkg/database/mysql"
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

	datasetRepo := repository.NewDatasetRepository(cfg, db)
	fileRepo, err := repository.NewFileRepository(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	repos := &services.Repositories{
		Dataset: datasetRepo,
		File:    fileRepo,
	}

	servicesInstance := services.NewServices(services.Deps{
		Repos:  repos,
		Config: cfg,
	})

	handler := handlers.NewHandler(servicesInstance, cfg)

	router := handler.Init()

	srv := server.NewServer(cfg, router)

	go func() {
		if err := srv.Run(); err != nil {
			logger.Fatal(err)
		}
	}()

	logger.Info(fmt.Sprintf("college-core-api started on port %s", cfg.Server.Port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		logger.Error(fmt.Errorf("server forced to shutdown: %w", err))
	}

	logger.Info("server exited")
}
