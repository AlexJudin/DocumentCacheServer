package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
	tempClient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/AlexJudin/DocumentCacheServer/config"
	"github.com/AlexJudin/DocumentCacheServer/internal/controller/api"
	"github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository"
	"github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository/cache"
	"github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository/client"
	"github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository/postgres"
	"github.com/AlexJudin/DocumentCacheServer/internal/temporal"
)

func startApp(cfg *config.Config) {
	connStr := cfg.GetDataSourceName()
	db, err := client.NewDatabase(connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.AutoMigrate()
	if err != nil {
		log.Fatal(err)
	}

	tempConStr := cfg.GetTemporalSource()
	temporalClient, err := temporal.NewTemporalClient(tempConStr)
	if err != nil {
		log.Fatal(err)
	}
	defer temporalClient.Close()

	connMgDbStr := cfg.GetMongoDBSourse()
	mgDb, err := client.NewMongoDBClient(connMgDbStr)
	if err != nil {
		log.Fatal(err)
	}
	defer mgDb.Close()

	cacheManager, err := client.ConnectToRedis(cfg)
	if err != nil {
		log.Error("error connecting to redis")
	}
	defer cacheManager.Close()

	fileClient, err := client.NewFileStorageClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// init cacheClient
	cacheRepo := cache.NewDocumentRepo(cfg, cacheManager)

	// init repository
	documentRepo := repository.NewDocumentRepo(db.DB, mgDb.Client, fileClient)
	userRepo := postgres.NewUserRepo(db.DB)
	tokenRepo := postgres.NewTokenStorageRepo(db.DB)

	runWorkflow(temporalClient, documentRepo)

	r := chi.NewRouter()
	api.AddRoutes(cfg, documentRepo, userRepo, tokenRepo, cacheRepo, temporalClient, r)

	startPprofServer()

	startHTTPServer(cfg, r)
}

func startPprofServer() {
	go func() {
		err := http.ListenAndServe("localhost:6060", nil)
		if err != nil {
			log.Error(err)
		}
	}()
}

func startHTTPServer(cfg *config.Config, r *chi.Mux) {
	var err error

	log.Info("Start api server")

	serverAddress := fmt.Sprintf(":%s", cfg.Port)
	serverErr := make(chan error)

	httpServer := &http.Server{
		Addr:    serverAddress,
		Handler: r,
	}

	go func() {
		log.Infof("Listening on %s", serverAddress)
		if err = httpServer.ListenAndServe(); err != nil {
			serverErr <- err
		}
		close(serverErr)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case <-stop:
		log.Info("Stop signal received. Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err = httpServer.Shutdown(ctx); err != nil {
			log.Errorf("error terminating server: %+v", err)
		}
		log.Info("The server has been stopped successfully")
	case err = <-serverErr:
		log.Errorf("Server error: %+v", err)
	}
}

func runWorkflow(client tempClient.Client, documentRepo *repository.DocumentRepo) {
	w := worker.New(client, temporal.SaveDocument, worker.Options{})

	w.RegisterWorkflow(documentRepo.SaveSagaWorkflow)
	w.RegisterWorkflow(documentRepo.DeleteSagaWorkflow)
	w.RegisterActivity(documentRepo.MetaStorage.Save)
	w.RegisterActivity(documentRepo.MetaStorage.DeleteById)
	w.RegisterActivity(documentRepo.MetaStorage.GetById)
	w.RegisterActivity(documentRepo.FileStorage.Upload)
	w.RegisterActivity(documentRepo.FileStorage.Delete)
	w.RegisterActivity(documentRepo.JsonStorage.Save)
	w.RegisterActivity(documentRepo.JsonStorage.DeleteById)

	err := w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatal(err)
	}
}
