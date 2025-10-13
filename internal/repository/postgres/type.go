package postgres

import (
	"github.com/AlexJudin/DocumentCacheServer/internal/model"
)

type User interface {
	GetByLogin(login string) (model.User, error)
	Save(user model.User) error
}

type TokenStorage interface {
	Save(token model.Token) error
	Get(accessTokenID string) (string, error)
	Delete(accessTokenID string) error
}
