package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/anton1ks96/college-core-api/pkg/logger"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type (
	Config struct {
		Server      Server
		Database    Database
		MinIO       MinIOConfig
		AuthService AuthServiceConfig
		RAGService  RAGServiceConfig
		Limits      LimitsConfig
	}

	Server struct {
		Host           string
		Port           string
		ReadTimeout    time.Duration
		WriteTimeout   time.Duration
		MaxHeaderBytes int
	}

	Database struct {
		DSN             string
		MaxConnections  int
		MaxIdle         int
		ConnMaxLifetime time.Duration
	}

	MinIOConfig struct {
		Endpoint  string
		Bucket    string
		UseSSL    bool
		AccessKey string
		SecretKey string
	}

	AuthServiceConfig struct {
		URL     string
		Timeout time.Duration
	}

	RAGServiceConfig struct {
		URL     string
		Token   string
		Timeout time.Duration
	}

	LimitsConfig struct {
		MaxFileSize        int64
		MaxDatasetsPerUser int
		UploadRateLimit    int
		AskRateLimit       int
	}
)

func Init() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		logger.Warn("No .env file found, using system environment variables")
	}

	if err := parseConfigFile("./configs"); err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to parse configuration file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	if err := setFromEnv(&cfg); err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to set environment variables: %w", err)
	}

	return &cfg, nil
}

func parseConfigFile(folder string) error {
	viper.AddConfigPath(folder)
	viper.SetConfigName("main")
	viper.SetConfigType("yml")

	return viper.ReadInConfig()
}

func setFromEnv(cfg *Config) error {
	cfg.Database.DSN = os.Getenv("DB_DSN")
	if cfg.Database.DSN == "" {
		return errors.New("DB_DSN environment variable is required")
	}

	cfg.MinIO.Endpoint = os.Getenv("MINIO_ENDPOINT")
	cfg.MinIO.AccessKey = os.Getenv("MINIO_ACCESS_KEY")
	cfg.MinIO.SecretKey = os.Getenv("MINIO_SECRET_KEY")

	if cfg.MinIO.Endpoint == "" {
		return errors.New("MINIO_ENDPOINT environment variable is required")
	}
	if cfg.MinIO.AccessKey == "" {
		return errors.New("MINIO_ACCESS_KEY environment variable is required")
	}
	if cfg.MinIO.SecretKey == "" {
		return errors.New("MINIO_SECRET_KEY environment variable is required")
	}

	cfg.AuthService.URL = os.Getenv("AUTH_SERVICE_URL")
	if cfg.AuthService.URL == "" {
		return errors.New("AUTH_SERVICE_URL environment variable is required")
	}

	cfg.RAGService.URL = os.Getenv("RAG_SERVICE_URL")
	cfg.RAGService.Token = os.Getenv("RAG_SERVICE_TOKEN")

	if cfg.RAGService.URL == "" {
		return errors.New("RAG_SERVICE_URL environment variable is required")
	}
	if cfg.RAGService.Token == "" {
		return errors.New("RAG_SERVICE_TOKEN environment variable is required")
	}

	return nil
}
