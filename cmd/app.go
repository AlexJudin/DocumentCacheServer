package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"

	"github.com/AlexJudin/DocumentCacheServer/config"
	"github.com/AlexJudin/DocumentCacheServer/internal/api/domain"
	"github.com/AlexJudin/DocumentCacheServer/internal/cache"
	"github.com/AlexJudin/DocumentCacheServer/internal/repository/mongodb"
	"github.com/AlexJudin/DocumentCacheServer/internal/repository/postgres"
)

func startApp(cfg *config.Сonfig) {
	connStr := cfg.GetDataSourceName()
	db, err := postgres.ConnectDB(connStr)
	if err != nil {
		log.Fatal(err)
	}

	connMgDbStr := cfg.GetMongoDBSourse()
	mgDb, err := mongodb.NewMongoDBClient(connMgDbStr)
	if err != nil {
		log.Fatal(err)
	}
	defer mgDb.Close()

	redisClient, err := cache.ConnectToRedis(cfg)
	if err != nil {
		log.Error("error connecting to redis")
	}

	r := chi.NewRouter()
	domain.AddRoutes(cfg, db, mgDb.Client, redisClient, r)

	startHTTPServer(cfg, r)
}

func startHTTPServer(cfg *config.Сonfig, r *chi.Mux) {
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
