package usecases

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	tempClient "go.temporal.io/sdk/client"

	"github.com/AlexJudin/DocumentCacheServer/internal/entity"
	"github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository"
	"github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository/cache"
	"github.com/AlexJudin/DocumentCacheServer/internal/model"
	"github.com/AlexJudin/DocumentCacheServer/internal/temporal"
	"github.com/AlexJudin/DocumentCacheServer/internal/temporal/saga"
)

var _ Document = (*DocumentUsecase)(nil)

type DocumentUsecase struct {
	Ctx                context.Context
	DocumentRepository repository.DocumentRepository
	Cache              cache.Document
	temporalClient     tempClient.Client
	sagaOrchestrator   saga.Orchestrator
}

func NewDocumentUsecase(docRepo repository.DocumentRepository, cache cache.Document, temporalClient tempClient.Client, sagaOrchestrator *saga.DocumentOrchestrator) *DocumentUsecase {
	return &DocumentUsecase{
		Ctx:                context.Background(),
		DocumentRepository: docRepo,
		Cache:              cache,
		temporalClient:     temporalClient,
		sagaOrchestrator:   sagaOrchestrator,
	}
}

func (t *DocumentUsecase) SaveDocument(document *entity.Document) error {
	uuidDoc := uuid.New().String()

	document.Meta.UUID = uuidDoc

	opts := tempClient.StartWorkflowOptions{TaskQueue: temporal.SaveDocument}
	_, err := t.temporalClient.ExecuteWorkflow(t.Ctx, opts, t.sagaOrchestrator.SaveDocument, document)
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

	metaDoc, err := t.DocumentRepository.GetById(uuid)
	if err != nil {
		return nil, entity.DefaultMimeType, err
	}

	if metaDoc.File {
		file, err := t.DocumentRepository.Download(t.Ctx, metaDoc.UUID)
		if err != nil {
			return nil, entity.DefaultMimeType, err
		}

		return file, metaDoc.Mime, nil
	}

	jsonDocMap, err := t.DocumentRepository.GetByDocumentId(t.Ctx, uuid)
	if err != nil {
		return nil, entity.DefaultMimeType, err
	}

	result := entity.ApiResponse{
		Data: jsonDocMap,
	}

	jsonDoc, err := json.Marshal(result)
	if err != nil {
		return nil, entity.DefaultMimeType, err
	}

	return jsonDoc, metaDoc.Mime, nil
}

func (t *DocumentUsecase) DeleteDocumentById(uuid string) error {
	opts := tempClient.StartWorkflowOptions{TaskQueue: temporal.SaveDocument}
	_, err := t.temporalClient.ExecuteWorkflow(t.Ctx, opts, t.sagaOrchestrator.DeleteDocument, uuid)
	if err != nil {
		return err
	}

	go t.Cache.Delete(t.Ctx, uuid)

	return nil
}
