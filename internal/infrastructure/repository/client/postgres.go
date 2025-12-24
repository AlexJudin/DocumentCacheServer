package client

import (
	"database/sql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"

	"github.com/AlexJudin/DocumentCacheServer/internal/model"
)

type Database struct {
	*gorm.DB
	sqlDB *sql.DB
}

func NewDatabase(connStr string) (*Database, error) {
	log.Info("Start connection to database")

	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	log.Info("Successfully connected to database")

	return &Database{
		DB:    db,
		sqlDB: sqlDB,
	}, nil
}

func (d *Database) Migrate() error {
	log.Info("Running migration")

	err := d.DB.AutoMigrate(
		&model.MetaDocument{},
		&model.User{},
		&model.Token{},
	)
	if err != nil {
		return err
	}

	log.Info("Successfully migrated")

	return nil
}

func (d *Database) Close() error {
	return d.sqlDB.Close()
}
