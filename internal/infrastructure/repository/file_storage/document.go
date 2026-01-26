package file_storage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	log "github.com/sirupsen/logrus"
)

var _ FileRepository = (*FileRepo)(nil)

const BucketName = "saga-files"

type FileRepo struct {
	Client     *minio.Client
	bucketName string
}

func NewFileRepository(minioClient *minio.Client) *FileRepo {
	return &FileRepo{
		Client:     minioClient,
		bucketName: BucketName,
	}
}

func (r *FileRepo) Upload(ctx context.Context, documentId string, data []byte) error {
	log.Infof("uploading saga [%s] file", documentId)

	size := int64(len(data))

	reader := bytes.NewReader(data)

	_, err := r.Client.PutObject(ctx, r.bucketName, documentId, reader, size, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		log.Debugf("failed to upload saga file: %+v", err)
		return fmt.Errorf("failed to upload saga [%s] file", documentId)
	}

	log.Infof("saga [%s] file uploaded successfully", documentId)

	return nil
}

func (r *FileRepo) Download(ctx context.Context, documentId string) ([]byte, error) {
	log.Infof("downloading saga [%s] file", documentId)

	object, err := r.Client.GetObject(ctx, r.bucketName, documentId, minio.GetObjectOptions{})
	if err != nil {
		log.Debugf("failed to get saga file: %+v", err)
		return nil, fmt.Errorf("failed to get saga [%s] file", documentId)
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		log.Debugf("failed to read saga file: %+v", err)
		return nil, fmt.Errorf("failed to read saga [%s] file", documentId)
	}

	log.Infof("saga [%s] file downloaded successfully", documentId)

	return data, nil
}

func (r *FileRepo) Delete(ctx context.Context, documentId string) error {
	log.Infof("deleting saga [%s] file", documentId)

	err := r.Client.RemoveObject(ctx, r.bucketName, documentId, minio.RemoveObjectOptions{})
	if err != nil {
		log.Debugf("failed to delete saga file: %+v", err)
		return fmt.Errorf("failed to delete saga [%s] file", documentId)
	}

	log.Infof("saga [%s] file deleted successfully", documentId)

	return nil
}
