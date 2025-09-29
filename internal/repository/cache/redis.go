package cache

import (
	"context"
	"strings"

	"github.com/redis/go-redis/v9"

	"github.com/AlexJudin/DocumentCacheServer/config"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func ConnectToRedis(cfg *config.Сonfig) (*redis.Client, error) {
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

	return client, nil
}
