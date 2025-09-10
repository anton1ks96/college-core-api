package config

import "github.com/spf13/viper"

type Config struct {
	Server struct {
		Host string
		Port int
	}

	Database struct {
		DSN string
	}

	MinIo struct {
		Endpoint   string
		BucketName string
		UseSSL     bool
		AccessKey  string
		SecretKey  string
	}

	RAG struct {
		Endpoint string
		APIKey   string
	}
}

func InitConfig(folder string) (*Config, error) {
	v := viper.New()
	v.AddConfigPath(folder)
	v.SetConfigName("main")
	v.SetConfigType("yml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	v.AutomaticEnv()

	_ = v.BindEnv("database.dsn", "DB_DSN")
	_ = v.BindEnv("minio.accesskey", "MINIO_ACCESS_KEY")
	_ = v.BindEnv("minio.secretkey", "MINIO_SECRET_KEY")
	_ = v.BindEnv("rag.service_token", "RAG_SERVICE_TOKEN")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
