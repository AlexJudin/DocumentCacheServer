package postgres

import (
	"errors"
	"fmt"
	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"

	"github.com/AlexJudin/DocumentCacheServer/internal/custom_error"
	"github.com/AlexJudin/DocumentCacheServer/internal/entity"
	"github.com/AlexJudin/DocumentCacheServer/internal/model"
)

var _ MetadataRepository = (*MetadataRepo)(nil)

type MetadataRepo struct {
	Db *gorm.DB
}

func NewMetadataRepository(db *gorm.DB) *MetadataRepo {
	return &MetadataRepo{Db: db}
}

func (r *MetadataRepo) Save(document *model.MetaDocument) error {
	log.Infof("saving saga [%s] metadata to database", document.UUID)

	err := r.Db.Create(&document).Error
	if err != nil {
		log.Debugf("failed to save saga metadata: %+v", err)
		return fmt.Errorf("failed to save saga [%s] metadata", document.UUID)
	}

	log.Infof("saga [%s] metadata saved successfully", document.UUID)

	return nil
}

func (r *MetadataRepo) GetList(req entity.DocumentListRequest) ([]model.MetaDocument, error) {
	log.Info("retrieving documents list from database")

	documents := make([]model.MetaDocument, req.Limit)

	err := r.Db.Model(&model.MetaDocument{}).
		Where(fmt.Sprintf("%s = ?", req.Key), req.Value).
		Where("? = ANY(meta_documents.grant)", req.Login).
		Limit(req.Limit).
		Offset(req.Offset).
		Find(&documents).
		Order("name asc").
		Order("created_at desc").Error
	if err != nil {
		log.Debugf("failed to retrieve documents list: %+v", err)
		return nil, fmt.Errorf("failed to retrieve documents list")
	}

	log.Info("documents list retrieved successfully")

	return documents, nil
}

func (r *MetadataRepo) GetById(uuid string) (model.MetaDocument, error) {
	log.Infof("retrieving saga [%s] metadata", uuid)

	var document model.MetaDocument

	err := r.Db.Model(&document).
		Where("uuid = ?", uuid).
		First(&document).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Infof("saga [%s] metadata not found", uuid)
			return document, custom_error.ErrDocumentNotFound
		}

		log.Debugf("failed to retrieve saga: %+v", err)
		return document, fmt.Errorf("failed to retrieve saga [%s] metadata", uuid)
	}

	log.Infof("saga [%s] metadata retrieved successfully", uuid)

	return document, nil
}

func (r *MetadataRepo) DeleteById(id string) error {
	log.Infof("deleting saga [%s] metadata", id)

	err := r.Db.Model(&model.MetaDocument{}).
		Where("uuid = ?", id).
		Delete(&model.MetaDocument{}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Infof("saga [%s] metadata not found", id)
			return custom_error.ErrDocumentNotFound
		}

		log.Debugf("failed to delete saga metadata: %+v", err)
		return fmt.Errorf("failed to delete saga [%s] metadata", id)
	}

	log.Infof("saga [%s] metadata deleted successfully", id)

	return nil
}
