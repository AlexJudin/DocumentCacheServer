package saga

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/AlexJudin/DocumentCacheServer/internal/entity"
	"github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository"
	"github.com/AlexJudin/DocumentCacheServer/internal/model"
)

var _ Orchestrator = (*DocumentOrchestrator)(nil)

type DocumentOrchestrator struct {
	DocumentRepository *repository.DocumentRepo
}

func NewDocumentOrchestrator(documentRepository *repository.DocumentRepo) *DocumentOrchestrator {
	return &DocumentOrchestrator{
		DocumentRepository: documentRepository,
	}
}

func (s *DocumentOrchestrator) SaveDocument(ctx context.Context, document *entity.Document) error {
	uuidDoc := document.Meta.UUID

	err := s.DocumentRepository.Save(document.Meta)
	if err != nil {
		log.Error("failed to save saga metadata",
			"uuid", uuidDoc,
			"error", err)
		return err
	}

	if document.Meta.File {
		if err = s.DocumentRepository.Upload(ctx, uuidDoc, document.File.Content); err != nil {
			log.Error("failed to upload file content",
				"uuid", uuidDoc,
				"error", err)

			if compErr := s.DocumentRepository.DeleteById(uuidDoc); compErr != nil {
				log.Error("compensation failed: failed to delete metadata after upload failure",
					"uuid", uuidDoc,
					"compensationError", compErr,
					"originalError", err)
			}

			return err
		}
		log.Info("saga saved successfully", "uuid", uuidDoc)

		return nil
	}

	if err = s.DocumentRepository.Store(ctx, uuidDoc, document.Json); err != nil {
		log.Error("failed to save JSON content",
			"uuid", uuidDoc,
			"error", err)

		if compErr := s.DocumentRepository.DeleteById(uuidDoc); compErr != nil {
			log.Error("compensation failed: failed to delete metadata after JSON save failure",
				"uuid", uuidDoc,
				"compensationError", compErr,
				"originalError", err)
		}

		return err
	}
	log.Info("saga saved successfully", "uuid", uuidDoc)

	return nil
}

func (s *DocumentOrchestrator) DeleteDocument(ctx context.Context, uuid string) error {
	var metaDoc model.MetaDocument
	metaDoc, err := s.DocumentRepository.GetById(uuid)
	if err != nil {
		log.Error("failed to get saga metadata", "uuid", uuid, "error", err)
		return err
	}

	err = s.DocumentRepository.DeleteById(uuid)
	if err != nil {
		log.Error("failed to delete saga metadata", "uuid", uuid, "error", err)
		return err
	}

	if metaDoc.File {
		if err = s.DocumentRepository.Delete(ctx, uuid); err != nil {
			log.Error("failed to delete file from storage", "uuid", uuid, "error", err)

			if compErr := s.DocumentRepository.Save(&metaDoc); compErr != nil {
				log.Error("compensation failed: unable to restore metadata",
					"uuid", uuid,
					"error", compErr)
			}

			return err
		}
		log.Info("saga delete successfully", "uuid", uuid)

		return nil
	}

	if err = s.DocumentRepository.DeleteByDocumentId(ctx, uuid); err != nil {
		log.Error("failed to delete JSON data", "uuid", uuid, "error", err)

		if compErr := s.DocumentRepository.Save(&metaDoc); compErr != nil {
			log.Error("compensation failed: unable to restore metadata",
				"uuid", uuid,
				"error", compErr)
		}

		return err
	}
	log.Info("saga delete successfully", "uuid", uuid)

	return nil
}
