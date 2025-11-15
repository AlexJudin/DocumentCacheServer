package client

import (
	"context"
	"strings"

	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"

	"github.com/AlexJudin/DocumentCacheServer/config"
)

func ConnectToRedis(cfg *config.Config) (*redis.Client, error) {
	log.Info("Start connection to Redis")

	var connStr strings.Builder

	connStr.WriteString(cfg.ConfigRedis.Host)
	connStr.WriteString(":")
	connStr.WriteString(cfg.ConfigRedis.Port)

	client := redis.NewClient(&redis.Options{
		Addr:     connStr.String(),
		Password: cfg.ConfigRedis.Password,
		DB:       0,
	})

	status := client.Ping(context.Background())
	if err := status.Err(); err != nil {
		return nil, err
	}

	log.Info("Successfully connected to Redis")

	return client, nil
}
