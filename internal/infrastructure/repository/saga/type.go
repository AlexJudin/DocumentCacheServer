package saga

import (
	"context"

	"github.com/AlexJudin/DocumentCacheServer/internal/entity"
)

type Orchestrator interface {
	SaveDocument(ctx context.Context, document *entity.Document) error
	DeleteDocument(ctx context.Context, uuid string) error
}
