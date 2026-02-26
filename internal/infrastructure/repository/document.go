package repository

import (
	"github.com/AlexJudin/DocumentCacheServer/internal/app/metric"
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

func NewDocumentRepository(db *gorm.DB, mongoClient *mongo.Client, minioClient *minio.Client, metrics *metric.DatabaseMetrics) *DocumentRepo {
	return &DocumentRepo{
		MetadataRepo: postgres.NewMetadataRepository(db, metrics),
		ContentRepo:  mongodb.NewContentRepository(mongoClient, metrics),
		FileRepo:     filestorage.NewFileRepository(minioClient),
	}
}
