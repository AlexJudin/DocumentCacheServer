package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var _ Client = (*CacheClientRepo)(nil)

type CacheClientRepo struct {
	RedisClient *redis.Client
}

func NewCacheClientRepo(redisClient *redis.Client) *CacheClientRepo {
	return &CacheClientRepo{
		RedisClient: redisClient,
	}
}

func (c CacheClientRepo) SetValue(ctx context.Context, key string, value any) error {
	err := c.RedisClient.Set(ctx, key, value, 5*time.Minute).Err()
	if err != nil {
		return err
	}

	return nil
}

func (c CacheClientRepo) GetValue(ctx context.Context, key string) (string, error) {
	result, err := c.RedisClient.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return result, nil
}
