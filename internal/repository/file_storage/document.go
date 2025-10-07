package file_storage

import (
	"bytes"
	"context"
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

func NewDocumentFileRepo(minioClient *minio.Client) *DocumentFileRepo {
	return &DocumentFileRepo{
		Client:     minioClient,
		bucketName: bucketName,
	}
}

func (m *DocumentFileRepo) Upload(ctx context.Context, documentName string, data []byte) error {
	log.Infof("start upload file document [%s]", documentName)

	size := int64(len(data))

	reader := bytes.NewReader(data)

	_, err := m.Client.PutObject(ctx, m.bucketName, documentName, reader, size, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		log.Debugf("error upload file document [%s]: %+v", documentName, err)
		return err
	}

	log.Infof("end upload file document [%s]", documentName)

	return nil
}

func (m *DocumentFileRepo) Download(ctx context.Context, documentName string) ([]byte, error) {
	log.Infof("start download file [%s]", documentName)

	object, err := m.Client.GetObject(ctx, m.bucketName, documentName, minio.GetObjectOptions{})
	if err != nil {
		log.Debugf("error download file [%s]: %+v", documentName, err)
		return nil, err
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		log.Debugf("error download file [%s]: %+v", documentName, err)
		return nil, err
	}

	log.Infof("end download file [%s]", documentName)

	return data, nil
}

func (m *DocumentFileRepo) Delete(ctx context.Context, documentName string) error {
	log.Infof("start deleting file [%s]", documentName)

	err := m.Client.RemoveObject(ctx, m.bucketName, documentName, minio.RemoveObjectOptions{})
	if err != nil {
		log.Debugf("error deleting file [%s]: %+v", documentName, err)
		return err
	}

	log.Infof("end deleting file [%s]", documentName)

	return nil
}
