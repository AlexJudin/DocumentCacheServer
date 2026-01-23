package usecases

import (
	"context"

	"github.com/google/uuid"
	tempClient "go.temporal.io/sdk/client"

	"github.com/AlexJudin/DocumentCacheServer/internal/entity"
	"github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository"
	"github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository/cache"
	"github.com/AlexJudin/DocumentCacheServer/internal/model"
	"github.com/AlexJudin/DocumentCacheServer/internal/temporal"
)

var _ Document = (*DocumentUsecase)(nil)

type DocumentUsecase struct {
	Ctx                context.Context
	DocumentRepository repository.Document
	Cache              cache.Document
	temporalClient     tempClient.Client
}

func NewDocumentUsecase(docRepo repository.Document, cache cache.Document, temporalClient tempClient.Client) *DocumentUsecase {
	return &DocumentUsecase{
		Ctx:                context.Background(),
		DocumentRepository: docRepo,
		Cache:              cache,
		temporalClient:     temporalClient,
	}
}

func (t *DocumentUsecase) SaveDocument(document *entity.Document) error {
	uuidDoc := uuid.New().String()

	document.Meta.UUID = uuidDoc

	opts := tempClient.StartWorkflowOptions{TaskQueue: temporal.SaveDocument}
	_, err := t.temporalClient.ExecuteWorkflow(t.Ctx, opts, t.DocumentRepository.SaveSagaWorkflow, document)
	if err != nil {
		return err
	}

	if document.Meta.File {
		go t.Cache.Set(t.Ctx, uuidDoc, document.Meta.Mime, document.File.Content, true)

		return nil
	}

	go t.Cache.Set(t.Ctx, uuidDoc, document.Meta.Mime, document.Json, false)

	return nil
}

func (t *DocumentUsecase) GetDocumentsList(req entity.DocumentListRequest) ([]model.MetaDocument, error) {
	return t.DocumentRepository.GetList(req)
}

func (t *DocumentUsecase) GetDocumentById(uuid string) ([]byte, string, error) {
	data, mime, ok := t.Cache.Get(t.Ctx, uuid)
	if ok {
		return data, mime, nil
	}

	return t.DocumentRepository.GetById(t.Ctx, uuid)
}

func (t *DocumentUsecase) DeleteDocumentById(uuid string) error {
	opts := tempClient.StartWorkflowOptions{TaskQueue: temporal.SaveDocument}
	_, err := t.temporalClient.ExecuteWorkflow(t.Ctx, opts, t.DocumentRepository.DeleteSagaWorkflow, uuid)
	if err != nil {
		return err
	}

	go t.Cache.Delete(t.Ctx, uuid)

	return nil
}
