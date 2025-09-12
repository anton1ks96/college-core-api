package repository

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/anton1ks96/college-core-api/internal/config"
	"github.com/anton1ks96/college-core-api/pkg/logger"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type FileMinIORepository struct {
	client *minio.Client
	bucket string
}

func NewFileRepository(cfg *config.Config) (*FileMinIORepository, error) {
	client, err := minio.New(cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, ""),
		Secure: cfg.MinIO.UseSSL,
	})

	if err != nil {
		logger.Error(fmt.Errorf("failed to create MinIO client: %w", err))
		return nil, err
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.MinIO.Bucket)
	if err != nil {
		logger.Error(fmt.Errorf("failed to check bucket existence: %w", err))
		return nil, err
	}

	if !exists {
		err = client.MakeBucket(ctx, cfg.MinIO.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			logger.Error(fmt.Errorf("failed to create bucket: %w", err))
			return nil, err
		}
		logger.Info(fmt.Sprintf("bucket %s created successfully", cfg.MinIO.Bucket))
	}

	logger.Info(fmt.Sprintf("successfully connected to MinIO at %s (bucket: %s)", cfg.MinIO.Endpoint, cfg.MinIO.Bucket))

	return &FileMinIORepository{
		client: client,
		bucket: cfg.MinIO.Bucket,
	}, nil
}

func (r *FileMinIORepository) Upload(ctx context.Context, path string, content io.Reader, contentType string) error {
	buf := new(bytes.Buffer)
	size, err := io.Copy(buf, content)
	if err != nil {
		logger.Error(fmt.Errorf("failed to read content: %w", err))
		return err
	}

	_, err = r.client.PutObject(ctx, r.bucket, path, buf, size, minio.PutObjectOptions{
		ContentType: contentType,
	})

	if err != nil {
		logger.Error(fmt.Errorf("failed to upload file to MinIO: %w", err))
		return err
	}

	logger.Debug(fmt.Sprintf("file uploaded to MinIO: %s", path))
	return nil
}

func (r *FileMinIORepository) Download(ctx context.Context, path string) ([]byte, error) {
	object, err := r.client.GetObject(ctx, r.bucket, path, minio.GetObjectOptions{})
	if err != nil {
		logger.Error(fmt.Errorf("failed to get object from MinIO: %w", err))
		return nil, err
	}
	defer object.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, object)
	if err != nil {
		logger.Error(fmt.Errorf("failed to read object content: %w", err))
		return nil, err
	}

	logger.Debug(fmt.Sprintf("file downloaded from MinIO: %s", path))
	return buf.Bytes(), nil
}

func (r *FileMinIORepository) Delete(ctx context.Context, path string) error {
	err := r.client.RemoveObject(ctx, r.bucket, path, minio.RemoveObjectOptions{})
	if err != nil {
		logger.Error(fmt.Errorf("failed to delete object from MinIO: %w", err))
		return err
	}

	logger.Debug(fmt.Sprintf("file deleted from MinIO: %s", path))
	return nil
}

func (r *FileMinIORepository) Exists(ctx context.Context, path string) (bool, error) {
	_, err := r.client.StatObject(ctx, r.bucket, path, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}
		logger.Error(fmt.Errorf("failed to check object existence: %w", err))
		return false, err
	}

	return true, nil
}
