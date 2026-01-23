package repository

import (
	"context"
	
	"go.temporal.io/sdk/workflow"

	"github.com/AlexJudin/DocumentCacheServer/internal/entity"
	"github.com/AlexJudin/DocumentCacheServer/internal/model"
)

type Document interface {
	SaveSagaWorkflow(ctxFlow workflow.Context, document *entity.Document) error
	GetList(req entity.DocumentListRequest) ([]model.MetaDocument, error)
	GetById(ctx context.Context, uuid string) ([]byte, string, error)
	DeleteSagaWorkflow(ctxFlow workflow.Context, uuid string) error
}
