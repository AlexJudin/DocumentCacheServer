package repository

import (
	"context"
	"encoding/json"
	"gorm.io/gorm"

	"github.com/minio/minio-go/v7"
	log "github.com/sirupsen/logrus"
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
	MetaStorage *postgres.DocumentMetaRepo
	JsonStorage *mongodb.DocumentJsonRepo
	FileStorage *filestorage.DocumentFileRepo
}

func NewDocumentRepo(db *gorm.DB, mongoClient *mongo.Client, minioClient *minio.Client) *DocumentRepo {
	return &DocumentRepo{
		MetaStorage: postgres.NewDocumentMetaRepo(db),
		JsonStorage: mongodb.NewDocumentJsonRepo(mongoClient),
		FileStorage: filestorage.NewDocumentFileRepo(minioClient),
	}
}

// Save - для сохранения документа используется паттерн SAGA
func (r *DocumentRepo) Save(ctx context.Context, document *entity.Document) error {
	uuidDoc := document.Meta.UUID

	err := r.MetaStorage.Save(document.Meta)
	if err != nil {
		return err
	}

	if document.Meta.File {
		if err = r.FileStorage.Upload(ctx, uuidDoc, document.File.Content); err != nil {
			if compErr := r.MetaStorage.DeleteById(uuidDoc); compErr != nil {
				log.Errorf("compensation failed: %+v", compErr)
			}
			return err
		}

		return nil
	}

	if err = r.JsonStorage.Save(ctx, uuidDoc, document.Json); err != nil {
		if compErr := r.MetaStorage.DeleteById(uuidDoc); compErr != nil {
			log.Errorf("compensation failed: %+v", compErr)
		}
		return err
	}

	return nil
}

func (r *DocumentRepo) GetList(req entity.DocumentListRequest) ([]model.MetaDocument, error) {
	return r.GetList(req)
}

func (r *DocumentRepo) GetById(ctx context.Context, uuid string) ([]byte, string, error) {
	metaDoc, err := r.MetaStorage.GetById(uuid)
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

// DeleteById - для удаления документа используется паттерн SAGA
func (r *DocumentRepo) DeleteById(ctx context.Context, uuid string) error {
	metaDoc, err := r.MetaStorage.GetById(uuid)
	if err != nil {
		return err
	}

	err = r.MetaStorage.DeleteById(uuid)
	if err != nil {
		return err
	}

	if metaDoc.File {
		if err = r.FileStorage.Delete(ctx, uuid); err != nil {
			if compErr := r.MetaStorage.Save(&metaDoc); compErr != nil {
				log.Errorf("compensation failed: %+v", compErr)
			}
			return err
		}

		return nil
	}

	if err = r.JsonStorage.DeleteById(ctx, uuid); err != nil {
		if compErr := r.MetaStorage.Save(&metaDoc); compErr != nil {
			log.Errorf("compensation failed: %+v", compErr)
		}
		return err
	}

	return nil
}
