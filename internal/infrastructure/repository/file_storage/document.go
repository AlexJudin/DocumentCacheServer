package file_storage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	log "github.com/sirupsen/logrus"
)

const BucketName = "document-files"

type DocumentFileRepo struct {
	Client     *minio.Client
	bucketName string
}

func NewDocumentFileRepo(minioClient *minio.Client) *DocumentFileRepo {
	return &DocumentFileRepo{
		Client:     minioClient,
		bucketName: BucketName,
	}
}

func (r *DocumentFileRepo) Upload(ctx context.Context, documentName string, data []byte) error {
	log.Infof("uploading document [%s] file", documentName)

	size := int64(len(data))

	reader := bytes.NewReader(data)

	_, err := r.Client.PutObject(ctx, r.bucketName, documentName, reader, size, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		log.Debugf("failed to upload document file: %+v", err)
		return fmt.Errorf("failed to upload document [%s] file", documentName)
	}

	log.Infof("document [%s] file uploaded successfully", documentName)

	return nil
}

func (r *DocumentFileRepo) Download(ctx context.Context, documentName string) ([]byte, error) {
	log.Infof("downloading document [%s] file", documentName)

	object, err := r.Client.GetObject(ctx, r.bucketName, documentName, minio.GetObjectOptions{})
	if err != nil {
		log.Debugf("failed to get document file: %+v", err)
		return nil, fmt.Errorf("failed to get document [%s] file", documentName)
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		log.Debugf("failed to read document file: %+v", err)
		return nil, fmt.Errorf("failed to read document [%s] file", documentName)
	}

	log.Infof("document [%s] file downloaded successfully", documentName)

	return data, nil
}

func (r *DocumentFileRepo) Delete(ctx context.Context, documentName string) error {
	log.Infof("deleting document [%s] file", documentName)

	err := r.Client.RemoveObject(ctx, r.bucketName, documentName, minio.RemoveObjectOptions{})
	if err != nil {
		log.Debugf("failed to delete document file: %+v", err)
		return fmt.Errorf("failed to delete document [%s] file", documentName)
	}

	log.Infof("document [%s] file deleted successfully", documentName)

	return nil
}
