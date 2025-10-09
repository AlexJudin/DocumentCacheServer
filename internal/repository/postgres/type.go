package postgres

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

type User interface {
	GetByLogin(login string) (model.User, error)
	Save(user model.User) error
}

type TokenStorage interface {
	Save(token model.Token) error
	Get(accessTokenID string) (string, error)
	Delete(accessTokenID string) error
}
