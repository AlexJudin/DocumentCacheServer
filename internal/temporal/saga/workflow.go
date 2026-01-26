package saga

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/AlexJudin/DocumentCacheServer/internal/entity"
	"github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository"
	"github.com/AlexJudin/DocumentCacheServer/internal/model"
)

var _ Orchestrator = (*DocumentOrchestrator)(nil)

type DocumentOrchestrator struct {
	DocumentRepository repository.DocumentRepository
}

func NewDocumentOrchestrator(documentRepository repository.DocumentRepository) *DocumentOrchestrator {
	return &DocumentOrchestrator{
		DocumentRepository: documentRepository,
	}
}

func (s *DocumentOrchestrator) SaveDocument(ctxFlow workflow.Context, document *entity.Document) error {
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

	err := workflow.ExecuteActivity(ctxFlow, s.DocumentRepository.Save, document.Meta).Get(ctxFlow, nil)
	if err != nil {
		logger.Error("failed to save saga metadata",
			"uuid", uuidDoc,
			"error", err)
		return err
	}

	if document.Meta.File {
		if err = workflow.ExecuteActivity(ctxFlow, s.DocumentRepository.Upload, uuidDoc, document.File.Content).Get(ctxFlow, nil); err != nil {
			logger.Error("failed to upload file content",
				"uuid", uuidDoc,
				"error", err)

			if compErr := workflow.ExecuteActivity(ctxFlow, s.DocumentRepository.DeleteById, uuidDoc).Get(ctxFlow, nil); compErr != nil {
				logger.Error("compensation failed: failed to delete metadata after upload failure",
					"uuid", uuidDoc,
					"compensationError", compErr,
					"originalError", err)
			}

			return err
		}
		logger.Info("saga saved successfully", "uuid", uuidDoc)

		return nil
	}

	if err = workflow.ExecuteActivity(ctxFlow, s.DocumentRepository.Store, uuidDoc, document.Json).Get(ctxFlow, nil); err != nil {
		logger.Error("failed to save JSON content",
			"uuid", uuidDoc,
			"error", err)

		if compErr := workflow.ExecuteActivity(ctxFlow, s.DocumentRepository.DeleteById, uuidDoc).Get(ctxFlow, nil); compErr != nil {
			logger.Error("compensation failed: failed to delete metadata after JSON save failure",
				"uuid", uuidDoc,
				"compensationError", compErr,
				"originalError", err)
		}

		return err
	}
	logger.Info("saga saved successfully", "uuid", uuidDoc)

	return nil
}

func (s *DocumentOrchestrator) DeleteDocument(ctxFlow workflow.Context, uuid string) error {
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
	err := workflow.ExecuteActivity(ctxFlow, s.DocumentRepository.GetById, uuid).Get(ctxFlow, &metaDoc)
	if err != nil {
		logger.Error("failed to get saga metadata", "uuid", uuid, "error", err)
		return err
	}

	err = workflow.ExecuteActivity(ctxFlow, s.DocumentRepository.DeleteById, uuid).Get(ctxFlow, nil)
	if err != nil {
		logger.Error("failed to delete saga metadata", "uuid", uuid, "error", err)
		return err
	}

	if metaDoc.File {
		if err = workflow.ExecuteActivity(ctxFlow, s.DocumentRepository.Delete, uuid).Get(ctxFlow, nil); err != nil {
			logger.Error("failed to delete file from storage", "uuid", uuid, "error", err)

			if compErr := workflow.ExecuteActivity(ctxFlow, s.DocumentRepository.Save, &metaDoc).Get(ctxFlow, nil); compErr != nil {
				logger.Error("compensation failed: unable to restore metadata",
					"uuid", uuid,
					"error", compErr)
			}

			return err
		}
		logger.Info("saga delete successfully", "uuid", uuid)

		return nil
	}

	if err = workflow.ExecuteActivity(ctxFlow, s.DocumentRepository.DeleteByDocumentId, uuid).Get(ctxFlow, nil); err != nil {
		logger.Error("failed to delete JSON data", "uuid", uuid, "error", err)

		if compErr := workflow.ExecuteActivity(ctxFlow, s.DocumentRepository.Save, &metaDoc).Get(ctxFlow, nil); compErr != nil {
			logger.Error("compensation failed: unable to restore metadata",
				"uuid", uuid,
				"error", compErr)
		}

		return err
	}
	logger.Info("saga delete successfully", "uuid", uuid)

	return nil
}
