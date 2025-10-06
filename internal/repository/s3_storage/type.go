package s3_storage

import "context"

type DocumentFile interface {
	Upload(ctx context.Context, documentName string, data []byte) error
	Download(ctx context.Context, documentName string) ([]byte, error)
	Delete(ctx context.Context, documentName string) error
}
