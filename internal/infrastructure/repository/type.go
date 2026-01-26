package repository

import (
	filestorage "github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository/file_storage"
	"github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository/mongodb"
	"github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository/postgres"
)

type DocumentRepository interface {
	postgres.MetadataRepository
	mongodb.ContentRepository
	filestorage.FileRepository
}
