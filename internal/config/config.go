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
		Limits      LimitsConfig
		Qdrant      QdrantConfig
		LLM         LLMConfig
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
		URL           string
		Timeout       time.Duration
		InternalToken string
	}

	LimitsConfig struct {
		MaxFileSize        int64
		MaxDatasetsPerUser int
		UploadRateLimit    int
		AskRateLimit       int
	}

	QdrantConfig struct {
		Host   string
		Port   string
		ApiKey string
	}

	LLMConfig struct {
		URL     string
		Model   string
		Timeout time.Duration
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

	cfg.AuthService.InternalToken = os.Getenv("INTERNAL_SERVICE_TOKEN")
	if cfg.AuthService.InternalToken == "" {
		return errors.New("INTERNAL_SERVICE_TOKEN environment variable is required")
	}

	cfg.Qdrant.Host = os.Getenv("QDRANT_HOST")
	if cfg.Qdrant.Host == "" {
		cfg.Qdrant.Host = "localhost"
	}

	cfg.Qdrant.Port = os.Getenv("QDRANT_PORT")
	if cfg.Qdrant.Port == "" {
		cfg.Qdrant.Port = "6334"
	}

	cfg.Qdrant.ApiKey = os.Getenv("QDRANT_API_KEY")

	cfg.LLM.URL = os.Getenv("LLM_URL")
	if cfg.LLM.URL == "" {
		cfg.LLM.URL = "http://localhost:11434"
	}

	cfg.LLM.Model = os.Getenv("LLM_MODEL")
	if cfg.LLM.Model == "" {
		cfg.LLM.Model = "default"
	}

	return nil
}
