package repository

import (
	"gorm.io/gorm"

	"github.com/minio/minio-go/v7"
	"go.mongodb.org/mongo-driver/mongo"

	filestorage "github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository/file_storage"
	"github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository/mongodb"
	"github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository/postgres"
)

type DocumentRepo struct {
	*postgres.MetadataRepo
	*mongodb.ContentRepo
	*filestorage.FileRepo
}

func NewDocumentRepository(db *gorm.DB, mongoClient *mongo.Client, minioClient *minio.Client) *DocumentRepo {
	return &DocumentRepo{
		MetadataRepo: postgres.NewMetadataRepository(db),
		ContentRepo:  mongodb.NewContentRepository(mongoClient),
		FileRepo:     filestorage.NewFileRepository(minioClient),
	}
}
