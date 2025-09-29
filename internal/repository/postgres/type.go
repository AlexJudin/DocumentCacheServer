package postgres

import (
	"github.com/AlexJudin/DocumentCacheServer/internal/entity"
	"github.com/AlexJudin/DocumentCacheServer/internal/model"
)

type Document interface {
	SaveDocument(document *model.MetaDocument) error
	GetDocumentsList(req entity.DocumentListRequest) ([]model.MetaDocument, error)
	GetDocumentById(uuid string) (model.MetaDocument, error)
	DeleteDocumentById(uuid string) error
}

type User interface {
	GetUserByLogin(login string) (model.User, error)
	SaveUser(user model.User) error
}

type TokenStorage interface {
	Save(token model.Token) error
	Get(accessTokenID string) (string, error)
	Delete(accessTokenID string) error
}
