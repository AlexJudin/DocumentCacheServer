package usecases

import (
	entity "github.com/AlexJudin/DocumentCacheServer/internal/entity"

	"github.com/AlexJudin/DocumentCacheServer/internal/model"
)

type Document interface {
	SaveDocument(document *entity.Document) error
	GetDocumentsList(req entity.DocumentListRequest) ([]model.MetaDocument, error)
	GetDocumentById(uuid string) ([]byte, string, error)
	DeleteDocumentById(uuid string) error
}

type Register interface {
	RegisterUser(login, password, token string) error
}

type Authorization interface {
	AuthorizationUser(login string, password string) (entity.Tokens, error)
	RefreshToken(refreshToken string) (entity.Tokens, error)
	DeleteToken(token string) error
}
