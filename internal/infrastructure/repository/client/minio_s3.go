package client

import (
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"

	"github.com/AlexJudin/DocumentCacheServer/config"
	filestorage "github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository/file_storage"
)

var buckets = []string{
	filestorage.BucketName,
}

func NewFileStorageClient(cfg *config.Config) (*minio.Client, error) {
	log.Info("Start connection to S3")

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create file_storage client: %w", err)
	}

	log.Info("Successfully connected to S3")

	err = ensureBucketExists(client)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func ensureBucketExists(client *minio.Client) error {
	log.Info("Check if bucket exists")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, bucket := range buckets {
		exists, err := client.BucketExists(ctx, bucket)
		if err != nil {
			return fmt.Errorf("failed to check bucket existence: %+v", err)
		}

		if !exists {
			err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
			if err != nil {
				return fmt.Errorf("failed to create bucket: %+v", err)
			}
			log.Infof("bucket '%s' created successfully", bucket)
		}
	}

	return nil
}
