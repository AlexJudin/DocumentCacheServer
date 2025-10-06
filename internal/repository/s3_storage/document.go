package s3_storage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	log "github.com/sirupsen/logrus"
)

var _ DocumentFile = (*DocumentFileRepo)(nil)

const bucketName = "document-files"

type DocumentFileRepo struct {
	Client     *minio.Client
	bucketName string
}

func NewDocumentFileRepo(clientS3 *minio.Client) *DocumentFileRepo {
	return &DocumentFileRepo{
		Client:     clientS3,
		bucketName: bucketName,
	}
}

func (m *DocumentFileRepo) Upload(ctx context.Context, documentName string, data []byte) error {
	reader := bytes.NewReader(data)

	_, err := m.Client.PutObject(ctx, m.bucketName, documentName, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return fmt.Errorf("failed to upload bytes: %w", err)
	}

	log.Printf("Data uploaded successfully as '%s'", documentName)
	return nil
}

func (m *DocumentFileRepo) Download(ctx context.Context, documentName string) ([]byte, error) {
	object, err := m.Client.GetObject(ctx, m.bucketName, documentName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		return nil, fmt.Errorf("failed to read object data: %w", err)
	}

	return data, nil
}

func (m *DocumentFileRepo) Delete(ctx context.Context, documentName string) error {
	err := m.Client.RemoveObject(ctx, m.bucketName, documentName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	log.Printf("Object '%s' deleted successfully", documentName)
	return nil
}
