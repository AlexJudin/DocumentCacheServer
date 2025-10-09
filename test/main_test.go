package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/AlexJudin/wallet_java_code/internal/api/controller/wallet"
	"github.com/AlexJudin/wallet_java_code/internal/repository"
	"github.com/AlexJudin/wallet_java_code/internal/usecases"
)

var (
	walletTest WalletTest
)

type WalletTest struct {
	db      *gorm.DB
	handler wallet.WalletHandler
}

func TestMain(m *testing.M) {
	log.Info("Start initializing test environment")

	if err := initialize(); err != nil {
		log.Panicf("error initializing test environment: %+v", err)
	}
	os.Exit(m.Run())
}

func initialize() error {
	err := godotenv.Load("../config/test_config.env")
	if err != nil {
		return err
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))

	db, err := repository.ConnectDB(connStr)
	if err != nil {
		return err
	}
	walletTest.db = db

	// init repository
	repo := repository.NewWalletRepo(db)

	// init usecases
	walletUC := usecases.NewWalletUsecase(repo)
	walletTest.handler = wallet.NewWalletHandler(walletUC)

	return nil
}

func truncateTable(db *gorm.DB) {
	err := db.Exec(`TRUNCATE payment_operations`).Error
	if err != nil {
		log.Fatalf("error truncate table: %+v", err)
	}
}

func closeDB() error {
	dbInstance, err := walletTest.db.DB()
	if err != nil {
		return err
	}
	err = dbInstance.Close()
	if err != nil {
		return err
	}

	return nil
}
