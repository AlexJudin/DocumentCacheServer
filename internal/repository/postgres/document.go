package postgres

import (
	"fmt"
	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"

	"github.com/AlexJudin/DocumentCacheServer/internal/api/entity"
	"github.com/AlexJudin/DocumentCacheServer/internal/model"
)

var _ Document = (*DocumentRepo)(nil)

type DocumentRepo struct {
	Db *gorm.DB
}

func NewDocumentRepo(db *gorm.DB) *DocumentRepo {
	return &DocumentRepo{Db: db}
}

func (r *DocumentRepo) SaveDocument(document *model.MetaDocument) error {
	log.Infof("start saving document [%s]", document.Name)

	err := r.Db.Create(&document).Error
	if err != nil {
		log.Debugf("error save document: %+v", err)
		return err
	}

	return nil
}

func (r *DocumentRepo) GetDocumentsList(req entity.DocumentListRequest) ([]model.MetaDocument, error) {
	log.Info("start getting documents list")

	documents := make([]model.MetaDocument, req.Limit)

	err := r.Db.Model(&model.MetaDocument{}).
		Where(fmt.Sprintf("%s = ?", req.Key), req.Value).
		Where("? = ANY(meta_documents.grant)", req.Login).
		Limit(req.Limit).
		Find(&documents).
		Order("name asc").
		Order("created_at desc").Error
	if err != nil {
		log.Debugf("error getting documents list: %+v", err)
		return nil, err
	}

	return documents, nil
}

func (r *DocumentRepo) GetDocumentById(uuid string) (model.MetaDocument, error) {
	log.Infof("start getting document by uuid[%s]", uuid)

	var document model.MetaDocument

	err := r.Db.Model(&document).
		Where("uuid = ?", uuid).
		First(&document).Error
	if err != nil {
		log.Debugf("error getting document by uuid[%s]: %+v", uuid, err)
		return document, err
	}

	return document, nil
}

func (r *DocumentRepo) DeleteDocumentById(id string) error {
	log.Infof("start deleting document by id[%s]", id)

	err := r.Db.Model(&model.MetaDocument{}).
		Where("uuid = ?", id).
		Delete(&model.MetaDocument{}).Error
	if err != nil {
		log.Debugf("error deleting document by id[%s]: %+v", id, err)
		return err
	}

	return nil
}
