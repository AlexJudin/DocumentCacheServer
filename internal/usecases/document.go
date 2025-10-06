package usecases

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/AlexJudin/DocumentCacheServer/config"
	"github.com/AlexJudin/DocumentCacheServer/internal/entity"
	"github.com/AlexJudin/DocumentCacheServer/internal/model"
	"github.com/AlexJudin/DocumentCacheServer/internal/repository/cache"
	filestorage "github.com/AlexJudin/DocumentCacheServer/internal/repository/file_storage"
	"github.com/AlexJudin/DocumentCacheServer/internal/repository/mongodb"
	"github.com/AlexJudin/DocumentCacheServer/internal/repository/postgres"
	s3storage "github.com/AlexJudin/DocumentCacheServer/internal/repository/s3_storage"
)

var _ Document = (*DocumentUsecase)(nil)

type DocumentUsecase struct {
	Ctx          context.Context
	DB           postgres.Document
	Cache        cache.Client
	FileStorage  filestorage.FileStorage
	MongoDB      mongodb.Document
	MinioStorage s3storage.DocumentFile
}

func NewDocumentUsecase(cfg *config.Config, db postgres.Document, cache cache.Client, mgdb mongodb.Document, fileStorage s3storage.DocumentFile) *DocumentUsecase {
	return &DocumentUsecase{
		Ctx:          context.Background(),
		DB:           db,
		Cache:        cache,
		FileStorage:  filestorage.NewFileStorageRepo(cfg),
		MongoDB:      mgdb,
		MinioStorage: fileStorage,
	}
}

func (t *DocumentUsecase) SaveDocument(document *entity.Document) error {
	uuidDoc := uuid.New().String()

	document.Meta.UUID = uuidDoc

	if document.Meta.File {
		filePath, err := t.FileStorage.Create(document)
		if err != nil {
			return err
		}

		err = t.Cache.Set(t.Ctx, uuidDoc, document.Meta.Mime, filePath, true)
		if err != nil {
			return err
		}

		document.Meta.FilePath = filePath
	}

	if len(document.Json) != 0 {
		err := t.MongoDB.SaveDocument(t.Ctx, uuidDoc, document.Json)
		if err != nil {
			return err
		}

		err = t.Cache.Set(t.Ctx, uuidDoc, document.Meta.Mime, document.Json, false)
		if err != nil {
			return err
		}
	}

	err := t.DB.SaveDocument(document.Meta)
	if err != nil {
		return err
	}

	return nil
}

func (t *DocumentUsecase) GetDocumentsList(req entity.DocumentListRequest) ([]model.MetaDocument, error) {
	return t.DB.GetDocumentsList(req)
}

func (t *DocumentUsecase) GetDocumentById(uuid string) ([]byte, string, error) {
	data, mime, ok := t.Cache.Get(t.Ctx, uuid)
	if ok {
		return data, mime, nil
	}

	metaDoc, err := t.DB.GetDocumentById(uuid)
	if err != nil {
		return nil, "", err
	}

	if metaDoc.File {
		file, err := t.FileStorage.Open(metaDoc.FilePath)
		if err != nil {
			return nil, "", err
		}

		return file, metaDoc.Mime, nil
	}

	jsonDocMap, err := t.MongoDB.GetDocumentById(t.Ctx, uuid)
	if err != nil {
		return nil, "", err
	}

	result := entity.ApiResponse{
		Data: jsonDocMap,
	}

	jsonDoc, err := json.Marshal(result)
	if err != nil {
		return nil, "", err
	}

	return jsonDoc, metaDoc.Mime, nil
}

func (t *DocumentUsecase) DeleteDocumentById(uuid string) error {
	err := t.FileStorage.Delete(uuid)
	if err != nil {
		return err
	}

	err = t.MongoDB.DeleteDocumentById(t.Ctx, uuid)
	if err != nil {
		return err
	}

	err = t.Cache.Delete(t.Ctx, uuid)
	if err != nil {
		return err
	}

	return t.DB.DeleteDocumentById(uuid)
}
