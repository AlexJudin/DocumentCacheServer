package test

import (
	"fmt"
	"gorm.io/gorm"
	"os"
	"testing"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"

	"github.com/AlexJudin/DocumentCacheServer/internal/controller/http/document"
	"github.com/AlexJudin/DocumentCacheServer/internal/repository/client"
	"github.com/AlexJudin/DocumentCacheServer/internal/repository/postgres"
	"github.com/AlexJudin/DocumentCacheServer/internal/usecases"
)

var (
	documentTest DocumentTest
)

type DocumentTest struct {
	db      *gorm.DB
	handler document.DocumentHandler
}

func TestMain(m *testing.M) {
	log.Info("Start initializing test environment")

	if err := initialize(); err != nil {
		log.Panicf("error initializing test environment: %+v", err)
	}
	os.Exit(m.Run())
}

func initialize() error {
	err := godotenv.Load("../config/config_test.env")
	if err != nil {
		return err
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))

	db, err := client.ConnectDB(connStr)
	if err != nil {
		return err
	}
	documentTest.db = db

	// init repository
	repo := postgres.NewDocumentMetaRepo(db)

	// init usecases
	docsUC := usecases.NewDocumentUsecase(repo)
	documentTest.handler = document.NewDocumentHandler(docsUC)

	return nil
}

func truncateTable(db *gorm.DB) {
	err := db.Exec(`TRUNCATE meta_documents`).Error
	if err != nil {
		log.Fatalf("error truncate table: %+v", err)
	}
}

func closeDB() error {
	dbInstance, err := documentTest.db.DB()
	if err != nil {
		return err
	}
	err = dbInstance.Close()
	if err != nil {
		return err
	}

	return nil
}
