package file_storage

import (
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"

	"github.com/AlexJudin/DocumentCacheServer/config"
)

func NewFileStorageClient(cfg *config.Config) (*minio.Client, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create file_storage client: %w", err)
	}

	err = ensureBucketExists(client)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func ensureBucketExists(client *minio.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %+v", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %+v", err)
		}
		log.Infof("Bucket '%s' created successfully")
	}

	return nil
}
