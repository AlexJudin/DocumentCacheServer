package file_storage

import (
	"context"
)

type Document interface {
	Upload(ctx context.Context, documentUUID string, data []byte) error
	Download(ctx context.Context, documentUUID string) ([]byte, error)
	Delete(ctx context.Context, documentUUID string) error
}
