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

	log.Info("Successfully connected to database")

	log.Info("Running migration")
	err = db.AutoMigrate(
		&model.MetaDocument{},
		&model.User{},
		&model.Token{},
	)
	if err != nil {
		return nil, err
	}

	log.Info("Successfully migrated")

	return db, nil
}
