package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/AlexJudin/DocumentCacheServer/internal/entity"
	"github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository"
	"github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository/cache"
	"github.com/AlexJudin/DocumentCacheServer/internal/model"
)

var _ Document = (*DocumentUsecase)(nil)

type DocumentUsecase struct {
	Ctx                context.Context
	DocumentRepository repository.Document
	Cache              cache.Document
}

func NewDocumentUsecase(docRepo repository.Document, cache cache.Document) *DocumentUsecase {
	return &DocumentUsecase{
		Ctx:                context.Background(),
		DocumentRepository: docRepo,
		Cache:              cache,
	}
}

func (t *DocumentUsecase) SaveDocument(document *entity.Document) error {
	uuidDoc := uuid.New().String()

	document.Meta.UUID = uuidDoc

	err := t.DocumentRepository.SaveSagaWorkflow(t.Ctx, document)
	if err != nil {
		return err
	}

	if document.Meta.File {
		go t.Cache.Set(t.Ctx, uuidDoc, document.Meta.Mime, document.File.Content, true)
	}

	if len(document.Json) != 0 {
		go t.Cache.Set(t.Ctx, uuidDoc, document.Meta.Mime, document.Json, false)
	}

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
	err := t.DocumentRepository.DeleteByIdSagaWorkflow(t.Ctx, uuid)
	if err != nil {
		return err
	}

	go t.Cache.Delete(t.Ctx, uuid)

	return nil
}
