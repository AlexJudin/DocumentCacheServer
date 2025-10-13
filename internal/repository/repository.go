package repository

import (
	"context"
	"gorm.io/gorm"

	"go.mongodb.org/mongo-driver/mongo"
)

type DocumentRepo struct {
	DB          gorm.DB
	MongoClient mongo.Client
}

func NewDocumentRepo(db *gorm.DB, mongoClient mongo.Client) *DocumentRepo {
	return &DocumentRepo{
		DB:          *db,
		MongoClient: mongoClient,
	}
}

func (r *DocumentRepo) Save(ctx context.Context) error {
	return nil
}
