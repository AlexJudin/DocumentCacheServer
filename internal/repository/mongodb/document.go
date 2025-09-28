package mongodb

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/AlexJudin/DocumentCacheServer/internal/model"
)

type MongoDBRepo struct {
	Client *mongo.Client
}

func NewMongoDBRepo(client *mongo.Client) *MongoDBRepo {
	return &MongoDBRepo{
		Client: client,
	}
}

func (r *MongoDBRepo) SaveDocument(ctx context.Context, uuid string, jsonDoc map[string]interface{}) error {
	collection := r.Client.Database(model.MongoDbName).Collection(model.MongoCollectionName)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	jsonDoc["_id"] = uuid

	_, err := collection.InsertOne(ctx, jsonDoc)
	if err != nil {
		log.Debugf("error save json document: %+v", err)
		return err
	}

	return nil
}

func (r *MongoDBRepo) GetDocumentById(ctx context.Context, uuid string) (map[string]interface{}, error) {
	collection := r.Client.Database(model.MongoDbName).Collection(model.MongoCollectionName)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var result map[string]interface{}
	err := collection.FindOne(ctx, bson.M{"_id": uuid}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Debugf("no document found by id [%s]", uuid)
			return nil, fmt.Errorf("document not found")
		}
		log.Debugf("error get json document[%s]: %+v", uuid, err)
		return nil, fmt.Errorf("failed to get document: %+v", err)
	}

	return result, nil
}

func (r *MongoDBRepo) DeleteDocumentById(ctx context.Context, uuid string) error {
	collection := r.Client.Database(model.MongoDbName).Collection(model.MongoCollectionName)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := collection.DeleteOne(ctx, bson.M{"_id": uuid})
	if err != nil {
		log.Debugf("error delete document by id [%s]: %+v", uuid, err)
		return err
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("document not found")
	}

	return nil
}
