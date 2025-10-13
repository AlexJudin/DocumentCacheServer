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

var _ Document = (*DocumentRepo)(nil)

type DocumentRepo struct {
	Db *gorm.DB
}

func NewDocumentRepo(db *gorm.DB) *DocumentRepo {
	return &DocumentRepo{Db: db}
}

func (r *DocumentRepo) Save(document *model.MetaDocument) error {
	log.Infof("saving document [%s] metadata to database", document.UUID)

	err := r.Db.Create(&document).Error
	if err != nil {
		log.Debugf("failed to save document metadata: %+v", err)
		return fmt.Errorf("failed to save document [%s] metadata", document.UUID)
	}

	log.Infof("document [%s] metadata saved successfully", document.UUID)

	return nil
}

func (r *DocumentRepo) GetList(req entity.DocumentListRequest) ([]model.MetaDocument, error) {
	log.Info("retrieving documents list from database")

	documents := make([]model.MetaDocument, req.Limit)

	err := r.Db.Model(&model.MetaDocument{}).
		Where(fmt.Sprintf("%s = ?", req.Key), req.Value).
		Where("? = ANY(meta_documents.grant)", req.Login).
		Limit(req.Limit).
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

func (r *DocumentRepo) GetById(uuid string) (model.MetaDocument, error) {
	log.Infof("retrieving document [%s] metadata", uuid)

	var document model.MetaDocument

	err := r.Db.Model(&document).
		Where("uuid = ?", uuid).
		First(&document).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Infof("document [%s] metadata not found", uuid)
			return document, custom_error.ErrDocumentNotFound
		}

		log.Debugf("failed to retrieve document: %+v", err)
		return document, fmt.Errorf("failed to retrieve document [%s] metadata", uuid)
	}

	log.Infof("document [%s] metadata retrieved successfully", uuid)

	return document, nil
}

func (r *DocumentRepo) DeleteById(id string) error {
	log.Infof("deleting document [%s] metadata", id)

	err := r.Db.Model(&model.MetaDocument{}).
		Where("uuid = ?", id).
		Delete(&model.MetaDocument{}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Infof("document [%s] metadata not found", id)
			return custom_error.ErrDocumentNotFound
		}

		log.Debugf("failed to delete document metadata: %+v", err)
		return fmt.Errorf("failed to delete document [%s] metadata", id)
	}

	log.Infof("document [%s] metadata deleted successfully", id)

	return nil
}
