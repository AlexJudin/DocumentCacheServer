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
	"github.com/AlexJudin/DocumentCacheServer/internal/metric"
	"github.com/AlexJudin/DocumentCacheServer/internal/model"
)

const (
	dataBaseType = "mongodb"

	saveDocumentContent       = "save_document_content"
	getDocumentContentById    = "get_document_content_by_id"
	deleteDocumentContentById = "delete_document_content_by_id"
)

var _ ContentRepository = (*ContentRepo)(nil)

type ContentRepo struct {
	Client       *mongo.Client
	QueryObserve metric.QueryObserver
}

func NewContentRepository(client *mongo.Client) *ContentRepo {
	return &ContentRepo{
		Client:       client,
		QueryObserve: metric.NewDatabaseMetrics(),
	}
}

func (r *ContentRepo) Save(ctx context.Context, uuid string, jsonDoc map[string]interface{}) error {
	log.Infof("saving document [%s] content", uuid)

	collection := r.Client.Database(model.MongoDbName).Collection(model.MongoCollectionName)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	jsonDoc["_id"] = uuid

	fn := func() error {
		_, err := collection.InsertOne(ctx, jsonDoc)
		if err != nil {
			return err
		}

		return nil
	}

	err := r.QueryObserve.Observe(fn, dataBaseType, saveDocumentContent)
	if err != nil {
		log.Debugf("failed to save document content: %+v", err)
		return fmt.Errorf("failed to save document [%s] content", uuid)
	}

	log.Infof("document [%s] content saved successfully", uuid)

	return nil
}

func (r *ContentRepo) GetById(ctx context.Context, uuid string) (map[string]interface{}, error) {
	log.Infof("retrieving document [%s] content from database", uuid)

	collection := r.Client.Database(model.MongoDbName).Collection(model.MongoCollectionName)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var result map[string]interface{}

	fn := func() error {
		err := collection.FindOne(ctx, bson.M{"_id": uuid}).Decode(&result)
		if err != nil {
			return err
		}

		return nil
	}

	err := r.QueryObserve.Observe(fn, dataBaseType, getDocumentContentById)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, custom_error.ErrDocumentNotFound
		}

		log.Debugf("failed to retrieve document content: %+v", err)
		return nil, fmt.Errorf("failed to retrieve document [%s] content", uuid)
	}

	log.Infof("document [%s] content retrieved successfully", uuid)

	return result, nil
}

func (r *ContentRepo) DeleteById(ctx context.Context, uuid string) error {
	log.Infof("deleting document [%s] content from database", uuid)

	collection := r.Client.Database(model.MongoDbName).Collection(model.MongoCollectionName)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	fn := func() error {
		result, err := collection.DeleteOne(ctx, bson.M{"_id": uuid})
		if err != nil {
			return fmt.Errorf("failed to delete document [%s] content", uuid)
		}

		if result.DeletedCount == 0 {
			return custom_error.ErrDocumentNotFound
		}

		return nil
	}

	err := r.QueryObserve.Observe(fn, dataBaseType, deleteDocumentContentById)
	if err != nil {
		log.Debugf("failed to delete document content: %+v", err)
		return err
	}

	log.Infof("document [%s] content deleted successfully", uuid)

	return nil
}
