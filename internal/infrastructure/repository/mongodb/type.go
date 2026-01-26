package mongodb

import "context"

type ContentRepository interface {
	Store(ctx context.Context, uuid string, jsonDoc map[string]interface{}) error
	GetByDocumentId(ctx context.Context, uuid string) (map[string]interface{}, error)
	DeleteByDocumentId(ctx context.Context, uuid string) error
}
