package repository

import (
	"context"

	"github.com/AlexJudin/DocumentCacheServer/internal/entity"
	"github.com/AlexJudin/DocumentCacheServer/internal/model"
)

type Document interface {
	SaveSagaWorkflow(ctx context.Context, document *entity.Document) error
	GetList(req entity.DocumentListRequest) ([]model.MetaDocument, error)
	GetById(ctx context.Context, uuid string) ([]byte, string, error)
	DeleteByIdSagaWorkflow(ctx context.Context, uuid string) error
}
