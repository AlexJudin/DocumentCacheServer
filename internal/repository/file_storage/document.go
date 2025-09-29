package file_storage

import (
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/AlexJudin/DocumentCacheServer/config"
	"github.com/AlexJudin/DocumentCacheServer/internal/entity"
)

var _ FileStorage = (*FileStorageRepo)(nil)

type FileStorageRepo struct {
	Cfg *config.Сonfig
}

func NewFileStorageRepo(cfg *config.Сonfig) *FileStorageRepo {
	return &FileStorageRepo{
		Cfg: cfg,
	}
}

func (r *FileStorageRepo) Create(document *entity.Document) (string, error) {
	log.Infof("start creating file [%s], document [%s]", document.File.Name, document)

	dirPath := filepath.Join(r.Cfg.MainDir, document.Meta.UUID)

	_, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		os.MkdirAll(dirPath, os.ModePerm)
	}

	filePath := filepath.Join(dirPath, document.File.Name)

	dst, err := os.Create(filePath)
	if err != nil {
		log.Debugf("error creating file [%s]: %+v", filePath, err)
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, document.File.Content)
	if err != nil {
		log.Debugf("error creating file [%s]: %+v", filePath, err)
		return "", err
	}

	log.Infof("end creating file [%s], document [%s]", document.File.Name, document)

	return filePath, nil
}

func (r *FileStorageRepo) Open(filePath string) ([]byte, error) {
	log.Infof("start opening file [%s]", filePath)

	file, err := os.ReadFile(filePath)
	if err != nil {
		log.Debugf("error opening file [%s]: %+v", filePath, err)
		return nil, err
	}

	log.Infof("end getting file [%s]", filePath)

	return file, nil
}

func (r *FileStorageRepo) Delete(uuid string) error {
	log.Infof("start deleting file [%s]", uuid)

	fullPath := filepath.Join(r.Cfg.MainDir, uuid)

	err := os.RemoveAll(fullPath)
	if err != nil {
		log.Debugf("error deleting file document [%s]: %+v", uuid, err)
		return err
	}

	log.Infof("end deleting file [%s]", uuid)

	return nil
}
