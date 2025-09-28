package file_storage

import (
	"github.com/AlexJudin/DocumentCacheServer/internal/api/entity"
)

type FileStorage interface {
	Create(document *entity.Document) (string, error)
	Open(filePath string) ([]byte, error)
	Delete(uuid string) error
}
