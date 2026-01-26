package saga

import (
	"go.temporal.io/sdk/workflow"

	"github.com/AlexJudin/DocumentCacheServer/internal/entity"
)

type Orchestrator interface {
	SaveDocument(ctxFlow workflow.Context, document *entity.Document) error
	DeleteDocument(ctxFlow workflow.Context, uuid string) error
}
