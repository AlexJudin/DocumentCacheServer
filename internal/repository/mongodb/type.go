package mongodb

import "context"

type Document interface {
	SaveDocument(ctx context.Context, uuid string, jsonDoc map[string]interface{}) error
	GetDocumentById(ctx context.Context, uuid string) (map[string]interface{}, error)
	DeleteDocumentById(ctx context.Context, uuid string) error
}
