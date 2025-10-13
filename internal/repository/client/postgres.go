package client

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"

	"github.com/AlexJudin/DocumentCacheServer/internal/model"
)

func ConnectDB(connStr string) (*gorm.DB, error) {
	log.Info("Start connection to database")

	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	log.Info("Connected to database")

	log.Info("Running migration")
	db.AutoMigrate(
		&model.MetaDocument{},
		&model.User{},
		&model.Token{},
	)

	return db, nil
}
