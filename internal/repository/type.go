package repository

import (
	"github.com/AlexJudin/DocumentCacheServer/internal/entity"
	"github.com/AlexJudin/DocumentCacheServer/internal/model"
)

type Document interface {
	Save(document *model.MetaDocument) error
	GetList(req entity.DocumentListRequest) ([]model.MetaDocument, error)
	GetById(uuid string) (model.MetaDocument, error)
	DeleteById(uuid string) error
}
