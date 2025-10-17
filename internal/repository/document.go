package repository

import (
	"context"
	"encoding/json"
	"gorm.io/gorm"

	"github.com/minio/minio-go/v7"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/AlexJudin/DocumentCacheServer/internal/entity"
	"github.com/AlexJudin/DocumentCacheServer/internal/model"
	filestorage "github.com/AlexJudin/DocumentCacheServer/internal/repository/file_storage"
	"github.com/AlexJudin/DocumentCacheServer/internal/repository/mongodb"
	"github.com/AlexJudin/DocumentCacheServer/internal/repository/postgres"
)

var _ Document = (*DocumentRepo)(nil)

// DocumentRepo - используется паттерн SAGA
type DocumentRepo struct {
	DB          *postgres.DocumentMetaRepo
	JsonStorage *mongodb.DocumentJsonRepo
	FileStorage *filestorage.DocumentFileRepo
}

func NewDocumentRepo(db *gorm.DB, mongoClient *mongo.Client, minioClient *minio.Client) *DocumentRepo {
	return &DocumentRepo{
		DB:          postgres.NewDocumentMetaRepo(db),
		JsonStorage: mongodb.NewDocumentJsonRepo(mongoClient),
		FileStorage: filestorage.NewDocumentFileRepo(minioClient),
	}
}

func (r *DocumentRepo) Save(ctx context.Context, document *entity.Document) error {
	uuidDoc := document.Meta.UUID

	err := r.DB.Save(document.Meta)
	if err != nil {
		return err
	}

	if document.Meta.File {
		err = r.FileStorage.Upload(ctx, uuidDoc, document.File.Content)
		if err != nil {
			return err
		}
	}

	if len(document.Json) != 0 {
		err = r.JsonStorage.Save(ctx, uuidDoc, document.Json)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *DocumentRepo) GetList(req entity.DocumentListRequest) ([]model.MetaDocument, error) {
	return r.GetList(req)
}

func (r *DocumentRepo) GetById(ctx context.Context, uuid string) ([]byte, string, error) {
	metaDoc, err := r.DB.GetById(uuid)
	if err != nil {
		return nil, "", err
	}

	if metaDoc.File {
		file, err := r.FileStorage.Download(ctx, metaDoc.UUID)
		if err != nil {
			return nil, "", err
		}

		return file, metaDoc.Mime, nil
	}

	jsonDocMap, err := r.JsonStorage.GetById(ctx, uuid)
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

func (r *DocumentRepo) DeleteById(ctx context.Context, uuid string) error {
	err := r.DB.DeleteById(uuid)
	if err != nil {
		return err
	}

	err = r.FileStorage.Delete(ctx, uuid)
	if err != nil {
		return err
	}

	err = r.JsonStorage.DeleteById(ctx, uuid)
	if err != nil {
		return err
	}

	return nil
}
