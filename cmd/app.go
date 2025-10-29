package main

import (
	"context"
	"fmt"
	"github.com/AlexJudin/DocumentCacheServer/internal/repository/client"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"

	"github.com/AlexJudin/DocumentCacheServer/config"
	"github.com/AlexJudin/DocumentCacheServer/internal/domain/controller"
)

func startApp(cfg *config.Config) {
	connStr := cfg.GetDataSourceName()
	db, err := client.ConnectDB(connStr)
	if err != nil {
		log.Fatal(err)
	}

	connMgDbStr := cfg.GetMongoDBSourse()
	mgDb, err := client.NewMongoDBClient(connMgDbStr)
	if err != nil {
		log.Fatal(err)
	}
	defer mgDb.Close()

	redisClient, err := client.ConnectToRedis(cfg)
	if err != nil {
		log.Error("error connecting to redis")
	}

	fileClient, err := client.NewFileStorageClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	controller.AddRoutes(cfg, db, mgDb.Client, redisClient, fileClient, r)

	startHTTPServer(cfg, r)
}

func startHTTPServer(cfg *config.Config, r *chi.Mux) {
	var err error

	log.Info("Start http server")

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

	stop := make(chan os.Signal)
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
