package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/AlexJudin/DocumentCacheServer/internal/custom_error"
	"github.com/AlexJudin/DocumentCacheServer/internal/model"
)

var _ ContentRepository = (*ContentRepo)(nil)

type ContentRepo struct {
	Client *mongo.Client
}

func NewContentRepository(client *mongo.Client) *ContentRepo {
	return &ContentRepo{
		Client: client,
	}
}

func (r *ContentRepo) Store(ctx context.Context, uuid string, jsonDoc map[string]interface{}) error {
	log.Infof("saving saga [%s] json to database", uuid)

	collection := r.Client.Database(model.MongoDbName).Collection(model.MongoCollectionName)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	jsonDoc["_id"] = uuid

	_, err := collection.InsertOne(ctx, jsonDoc)
	if err != nil {
		log.Debugf("failed to save saga json: %+v", err)
		return fmt.Errorf("failed to save saga [%s] json", uuid)
	}

	log.Infof("saga [%s] json saved successfully", uuid)

	return nil
}

func (r *ContentRepo) GetByDocumentId(ctx context.Context, uuid string) (map[string]interface{}, error) {
	log.Infof("retrieving saga [%s] json from database", uuid)

	collection := r.Client.Database(model.MongoDbName).Collection(model.MongoCollectionName)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var result map[string]interface{}
	err := collection.FindOne(ctx, bson.M{"_id": uuid}).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, custom_error.ErrDocumentNotFound
		}

		log.Debugf("failed to retrieve saga json: %+v", err)
		return nil, fmt.Errorf("failed to retrieve saga [%s] json", uuid)
	}

	log.Infof("saga [%s] json retrieved successfully", uuid)

	return result, nil
}

func (r *ContentRepo) DeleteByDocumentId(ctx context.Context, uuid string) error {
	log.Infof("deleting saga [%s] json from database", uuid)

	collection := r.Client.Database(model.MongoDbName).Collection(model.MongoCollectionName)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := collection.DeleteOne(ctx, bson.M{"_id": uuid})
	if err != nil {
		log.Debugf("failed to delete saga json: %+v", err)
		return fmt.Errorf("failed to delete saga [%s] json", uuid)
	}

	if result.DeletedCount == 0 {
		return custom_error.ErrDocumentNotFound
	}

	log.Infof("saga [%s] json deleted successfully", uuid)

	return nil
}
