package postgres

import (
	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"

	"github.com/AlexJudin/DocumentCacheServer/internal/model"
)

var _ User = (*UserRepo)(nil)

type UserRepo struct {
	Db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{Db: db}
}

func (r *UserRepo) SaveUser(user model.User) error {
	log.Infof("start saving user with login [%s]", user.Login)

	err := r.Db.Create(&user).Error
	if err != nil {
		log.Debugf("error create user: %+v", err)
		return err
	}

	return nil
}

func (r *UserRepo) GetUserByLogin(login string) (model.User, error) {
	log.Infof("start getting user by login [%s]", login)

	var user model.User

	err := r.Db.Model(user).
		Where("login = ?", login).
		Find(&user).Error
	if err != nil {
		log.Debugf("error getting user by login [%s]: %+v", login, err)
		return user, err
	}

	return user, nil
}
