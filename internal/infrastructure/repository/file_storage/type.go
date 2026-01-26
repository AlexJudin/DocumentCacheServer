package file_storage

import "context"

type FileRepository interface {
	Upload(ctx context.Context, documentId string, data []byte) error
	Download(ctx context.Context, documentId string) ([]byte, error)
	Delete(ctx context.Context, documentId string) error
}
