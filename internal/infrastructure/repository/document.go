package repository

import (
	"context"
	"encoding/json"
	"gorm.io/gorm"
	"time"

	"github.com/minio/minio-go/v7"
	"go.mongodb.org/mongo-driver/mongo"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/AlexJudin/DocumentCacheServer/internal/entity"
	filestorage "github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository/file_storage"
	"github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository/mongodb"
	"github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository/postgres"
	"github.com/AlexJudin/DocumentCacheServer/internal/model"
)

var _ Document = (*DocumentRepo)(nil)

// DocumentRepo - используется паттерн SAGA
type DocumentRepo struct {
	MetaStorage *postgres.DocumentMetaRepo
	JsonStorage *mongodb.DocumentJsonRepo
	FileStorage *filestorage.DocumentFileRepo
}

func NewDocumentRepo(db *gorm.DB, mongoClient *mongo.Client, minioClient *minio.Client) *DocumentRepo {
	return &DocumentRepo{
		MetaStorage: postgres.NewDocumentMetaRepo(db),
		JsonStorage: mongodb.NewDocumentJsonRepo(mongoClient),
		FileStorage: filestorage.NewDocumentFileRepo(minioClient),
	}
}

// SaveSagaWorkflow - для сохранения документа используется паттерн SAGA
func (r *DocumentRepo) SaveSagaWorkflow(ctxFlow workflow.Context, document *entity.Document) error {
	ctxFlow = workflow.WithActivityOptions(ctxFlow, workflow.ActivityOptions{
		StartToCloseTimeout: 15 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:        10 * time.Minute,
			BackoffCoefficient:     1,
			MaximumInterval:        time.Hour,
			MaximumAttempts:        1,
			NonRetryableErrorTypes: nil,
		},
	})
	logger := workflow.GetLogger(ctxFlow)

	uuidDoc := document.Meta.UUID

	err := workflow.ExecuteActivity(ctxFlow, r.MetaStorage.Save, document.Meta).Get(ctxFlow, nil)
	if err != nil {
		logger.Error("failed to save document metadata",
			"uuid", uuidDoc,
			"error", err)
		return err
	}

	if document.Meta.File {
		if err = workflow.ExecuteActivity(ctxFlow, r.FileStorage.Upload, uuidDoc, document.File.Content).Get(ctxFlow, nil); err != nil {
			logger.Error("failed to upload file content",
				"uuid", uuidDoc,
				"error", err)

			if compErr := workflow.ExecuteActivity(ctxFlow, r.MetaStorage.DeleteById, uuidDoc).Get(ctxFlow, nil); compErr != nil {
				logger.Error("compensation failed: failed to delete metadata after upload failure",
					"uuid", uuidDoc,
					"compensationError", compErr,
					"originalError", err)
			}

			return err
		}
		logger.Info("document saved successfully", "uuid", uuidDoc)

		return nil
	}

	if err = workflow.ExecuteActivity(ctxFlow, r.JsonStorage.Save, uuidDoc, document.Json).Get(ctxFlow, nil); err != nil {
		logger.Error("failed to save JSON content",
			"uuid", uuidDoc,
			"error", err)

		if compErr := workflow.ExecuteActivity(ctxFlow, r.MetaStorage.DeleteById, uuidDoc).Get(ctxFlow, nil); compErr != nil {
			logger.Error("compensation failed: failed to delete metadata after JSON save failure",
				"uuid", uuidDoc,
				"compensationError", compErr,
				"originalError", err)
		}

		return err
	}
	logger.Info("document saved successfully", "uuid", uuidDoc)

	return nil
}

func (r *DocumentRepo) GetList(req entity.DocumentListRequest) ([]model.MetaDocument, error) {
	return r.MetaStorage.GetList(req)
}

func (r *DocumentRepo) GetById(ctx context.Context, uuid string) ([]byte, string, error) {
	metaDoc, err := r.MetaStorage.GetById(uuid)
	if err != nil {
		return nil, entity.DefaultMimeType, err
	}

	if metaDoc.File {
		file, err := r.FileStorage.Download(ctx, metaDoc.UUID)
		if err != nil {
			return nil, entity.DefaultMimeType, err
		}

		return file, metaDoc.Mime, nil
	}

	jsonDocMap, err := r.JsonStorage.GetById(ctx, uuid)
	if err != nil {
		return nil, entity.DefaultMimeType, err
	}

	result := entity.ApiResponse{
		Data: jsonDocMap,
	}

	jsonDoc, err := json.Marshal(result)
	if err != nil {
		return nil, entity.DefaultMimeType, err
	}

	return jsonDoc, metaDoc.Mime, nil
}

// DeleteSagaWorkflow - для удаления документа используется паттерн SAGA
func (r *DocumentRepo) DeleteSagaWorkflow(ctxFlow workflow.Context, uuid string) error {
	ctxFlow = workflow.WithActivityOptions(ctxFlow, workflow.ActivityOptions{
		StartToCloseTimeout: 15 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:        10 * time.Minute,
			BackoffCoefficient:     1,
			MaximumInterval:        time.Hour,
			MaximumAttempts:        1,
			NonRetryableErrorTypes: nil,
		},
	})
	logger := workflow.GetLogger(ctxFlow)

	var metaDoc *model.MetaDocument
	err := workflow.ExecuteActivity(ctxFlow, r.MetaStorage.GetById, uuid).Get(ctxFlow, &metaDoc)
	if err != nil {
		logger.Error("failed to get document metadata", "uuid", uuid, "error", err)
		return err
	}

	err = workflow.ExecuteActivity(ctxFlow, r.MetaStorage.DeleteById, uuid).Get(ctxFlow, nil)
	if err != nil {
		logger.Error("failed to delete document metadata", "uuid", uuid, "error", err)
		return err
	}

	if metaDoc.File {
		if err = workflow.ExecuteActivity(ctxFlow, r.FileStorage.Delete, uuid).Get(ctxFlow, nil); err != nil {
			logger.Error("failed to delete file from storage", "uuid", uuid, "error", err)

			if compErr := workflow.ExecuteActivity(ctxFlow, r.MetaStorage.Save, &metaDoc).Get(ctxFlow, nil); compErr != nil {
				logger.Error("compensation failed: unable to restore metadata",
					"uuid", uuid,
					"error", compErr)
			}

			return err
		}
		logger.Info("document delete successfully", "uuid", uuid)

		return nil
	}

	if err = workflow.ExecuteActivity(ctxFlow, r.JsonStorage.DeleteById, uuid).Get(ctxFlow, nil); err != nil {
		logger.Error("failed to delete JSON data", "uuid", uuid, "error", err)

		if compErr := workflow.ExecuteActivity(ctxFlow, r.MetaStorage.Save, &metaDoc).Get(ctxFlow, nil); compErr != nil {
			logger.Error("compensation failed: unable to restore metadata",
				"uuid", uuid,
				"error", compErr)
		}

		return err
	}
	logger.Info("document delete successfully", "uuid", uuid)

	return nil
}
