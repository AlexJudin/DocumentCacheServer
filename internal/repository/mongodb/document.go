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

var _ Document = (*DocumentRepo)(nil)

type DocumentRepo struct {
	Client *mongo.Client
}

func NewDocumentRepo(client *mongo.Client) *DocumentRepo {
	return &DocumentRepo{
		Client: client,
	}
}

func (r *DocumentRepo) Save(ctx context.Context, uuid string, jsonDoc map[string]interface{}) error {
	log.Infof("saving document [%s] json to database", uuid)

	collection := r.Client.Database(model.MongoDbName).Collection(model.MongoCollectionName)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	jsonDoc["_id"] = uuid

	_, err := collection.InsertOne(ctx, jsonDoc)
	if err != nil {
		log.Debugf("failed to save document json: %+v", err)
		return fmt.Errorf("failed to save document [%s] json", uuid)
	}

	log.Infof("document [%s] json saved successfully", uuid)

	return nil
}

func (r *DocumentRepo) GetById(ctx context.Context, uuid string) (map[string]interface{}, error) {
	log.Infof("retrieving document [%s] json from database", uuid)

	collection := r.Client.Database(model.MongoDbName).Collection(model.MongoCollectionName)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var result map[string]interface{}
	err := collection.FindOne(ctx, bson.M{"_id": uuid}).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, custom_error.ErrDocumentNotFound
		}

		log.Debugf("failed to retrieve document json: %+v", err)
		return nil, fmt.Errorf("failed to retrieve document [%s] json", uuid)
	}

	log.Infof("document [%s] json retrieved successfully", uuid)

	return result, nil
}

func (r *DocumentRepo) DeleteById(ctx context.Context, uuid string) error {
	log.Infof("deleting document [%s] json from database", uuid)

	collection := r.Client.Database(model.MongoDbName).Collection(model.MongoCollectionName)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := collection.DeleteOne(ctx, bson.M{"_id": uuid})
	if err != nil {
		log.Debugf("failed to delete document json: %+v", err)
		return fmt.Errorf("failed to delete document [%s] json", uuid)
	}

	if result.DeletedCount == 0 {
		return custom_error.ErrDocumentNotFound
	}

	log.Infof("document [%s] json deleted successfully", uuid)

	return nil
}
