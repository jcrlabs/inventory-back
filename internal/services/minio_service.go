package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/jonathanCaamano/inventory-back/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOService struct {
	client *minio.Client
	bucket string
}

func NewMinIOService(cfg *config.Config) (*MinIOService, error) {
	client, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: cfg.MinIOUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	svc := &MinIOService{client: client, bucket: cfg.MinIOBucket}

	if err := svc.ensureBucket(); err != nil {
		return nil, err
	}

	return svc, nil
}

func (s *MinIOService) ensureBucket() error {
	ctx := context.Background()
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket: %w", err)
	}
	if !exists {
		if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}

		policy := fmt.Sprintf(`{
			"Version":"2012-10-17",
			"Statement":[{
				"Effect":"Allow",
				"Principal":{"AWS":["*"]},
				"Action":["s3:GetObject"],
				"Resource":["arn:aws:s3:::%s/products/*"]
			}]
		}`, s.bucket)

		if err := s.client.SetBucketPolicy(ctx, s.bucket, policy); err != nil {
			log.Printf("Warning: failed to set bucket policy: %v", err)
		}

		log.Printf("Bucket %s created", s.bucket)
	}
	return nil
}

func (s *MinIOService) UploadProductImage(file multipart.File, header *multipart.FileHeader) (objectKey string, err error) {
	ext := ""
	contentType := header.Header.Get("Content-Type")

	switch contentType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	default:
		return "", fmt.Errorf("unsupported image type: %s", contentType)
	}

	objectKey = fmt.Sprintf("products/%s%s", uuid.New().String(), ext)

	_, err = s.client.PutObject(
		context.Background(),
		s.bucket,
		objectKey,
		file,
		header.Size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %w", err)
	}

	return objectKey, nil
}

func (s *MinIOService) GetPresignedURL(objectKey string, expiry time.Duration) (string, error) {
	if objectKey == "" {
		return "", nil
	}

	reqParams := make(url.Values)
	presignedURL, err := s.client.PresignedGetObject(
		context.Background(),
		s.bucket,
		objectKey,
		expiry,
		reqParams,
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL.String(), nil
}

func (s *MinIOService) DeleteObject(objectKey string) error {
	if objectKey == "" {
		return nil
	}
	return s.client.RemoveObject(
		context.Background(),
		s.bucket,
		objectKey,
		minio.RemoveObjectOptions{},
	)
}

func (s *MinIOService) GetObject(objectKey string) (io.ReadCloser, *minio.ObjectInfo, error) {
	obj, err := s.client.GetObject(context.Background(), s.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, nil, err
	}
	info, err := obj.Stat()
	if err != nil {
		return nil, nil, err
	}
	return obj, &info, nil
}
