package postgres

import (
	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"

	"github.com/AlexJudin/DocumentCacheServer/internal/model"
)

var _ TokenStorage = (*TokenStorageRepo)(nil)

type TokenStorageRepo struct {
	Db *gorm.DB
}

func NewTokenStorageRepo(db *gorm.DB) *TokenStorageRepo {
	return &TokenStorageRepo{Db: db}
}

func (r *TokenStorageRepo) Save(token model.Token) error {
	log.Infof("start saving token with user login [%s]", token.Login)

	err := r.Db.Create(&token).Error
	if err != nil {
		log.Debugf("error create token: %+v", err)
		return err
	}

	return nil
}

func (r *TokenStorageRepo) Get(accessTokenID string) (string, error) {
	log.Infof("start getting user login with access token [%s]", accessTokenID)

	var token model.Token

	err := r.Db.Model(&token).
		Where("access_token_id = ?", accessTokenID).
		Find(&token).Error
	if err != nil {
		log.Debugf("error getting token by id[%s]: %+v", accessTokenID, err)
		return "", err
	}

	return token.Login, nil
}

func (r *TokenStorageRepo) Delete(accessTokenID string) error {
	log.Infof("start deleting user login with access token [%s]", accessTokenID)

	var token model.Token

	err := r.Db.Model(&token).
		Where("access_token_id = ?", accessTokenID).
		Delete(&token).Error
	if err != nil {
		log.Debugf("error deleting token by id[%s]: %+v", accessTokenID, err)
		return err
	}

	return nil
}
