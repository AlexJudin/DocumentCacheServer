package usecases

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/google/uuid"

	"github.com/AlexJudin/DocumentCacheServer/config"
	"github.com/AlexJudin/DocumentCacheServer/internal/api/entity"
	"github.com/AlexJudin/DocumentCacheServer/internal/cache"
	"github.com/AlexJudin/DocumentCacheServer/internal/model"
	filestorage "github.com/AlexJudin/DocumentCacheServer/internal/repository/file_storage"
	"github.com/AlexJudin/DocumentCacheServer/internal/repository/mongodb"
	"github.com/AlexJudin/DocumentCacheServer/internal/repository/postgres"
)

var _ Document = (*DocumentUsecase)(nil)

type DocumentUsecase struct {
	Ctx         context.Context
	DB          postgres.Document
	Cache       cache.Client
	FileStorage filestorage.FileStorage
	MongoDB     mongodb.Document
}

func NewDocumentUsecase(cfg *config.Сonfig, db postgres.Document, cache cache.Client, mgdb mongodb.Document) *DocumentUsecase {
	return &DocumentUsecase{
		Ctx:         context.Background(),
		DB:          db,
		Cache:       cache,
		FileStorage: filestorage.NewFileStorageRepo(cfg),
		MongoDB:     mgdb,
	}
}

func (t *DocumentUsecase) SaveDocument(document *entity.Document) error {
	document.Meta.UUID = uuid.New().String()

	if document.Meta.File {
		filePath, err := t.FileStorage.Create(document)
		if err != nil {
			return err
		}

		document.Meta.FilePath = filePath
	}

	err := t.MongoDB.SaveDocument(t.Ctx, document.Meta.UUID, document.Json)
	if err != nil {
		return err
	}

	err = t.DB.SaveDocument(document.Meta)
	if err != nil {
		return err
	}

	return nil
}

func (t *DocumentUsecase) GetDocumentsList(req entity.DocumentListRequest) ([]model.MetaDocument, error) {
	return t.DB.GetDocumentsList(req)
}

func (t *DocumentUsecase) GetDocumentById(uuid string) ([]byte, string, error) {
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

	return t.DB.DeleteDocumentById(uuid)
}

func (t *DocumentUsecase) SetCacheValue(ctx context.Context, walletUUID string, balance int64) error {
	return t.Cache.SetValue(ctx, walletUUID, balance)
}

func (t *DocumentUsecase) GetCacheValue(ctx context.Context, walletUUID string) (int64, error) {
	result, err := t.Cache.GetValue(ctx, walletUUID)
	if err != nil {
		return 0, err
	}

	balance, err := strconv.Atoi(result)
	if err != nil {
		return 0, err
	}

	return int64(balance), nil
}
