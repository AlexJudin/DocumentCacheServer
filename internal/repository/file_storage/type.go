package file_storage

import (
	"context"
)

type DocumentFile interface {
	Upload(ctx context.Context, documentUUID string, data []byte) (string, error)
	Download(ctx context.Context, documentUUID string) ([]byte, error)
	Delete(ctx context.Context, documentUUID string) error
}
