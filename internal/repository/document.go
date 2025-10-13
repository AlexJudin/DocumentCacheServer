package repository

import (
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

type DocumentRepo struct {
	DB          *postgres.DocumentRepo
	MongoClient mongodb.Document
	FileStorage filestorage.Document
}

func NewDocumentRepo(db *gorm.DB, mongoClient *mongo.Client, minioClient *minio.Client) *DocumentRepo {
	return &DocumentRepo{
		DB:          postgres.NewDocumentRepo(db),
		MongoClient: mongodb.NewDocumentRepo(mongoClient),
		FileStorage: filestorage.NewDocumentRepo(minioClient),
	}
}

func (r *DocumentRepo) Save(document *model.MetaDocument) error {
	return nil
}

func (r *DocumentRepo) GetList(req entity.DocumentListRequest) ([]model.MetaDocument, error) {
	return []model.MetaDocument{}, nil
}

func (r *DocumentRepo) GetById(uuid string) (model.MetaDocument, error) {
	return model.MetaDocument{}, nil
}

func (r *DocumentRepo) DeleteById(uuid string) error {
	return nil
}
