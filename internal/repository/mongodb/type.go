package mongodb

import "context"

type Document interface {
	Save(ctx context.Context, uuid string, jsonDoc map[string]interface{}) error
	GetById(ctx context.Context, uuid string) (map[string]interface{}, error)
	DeleteById(ctx context.Context, uuid string) error
}
