package cache

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"

	"github.com/AlexJudin/DocumentCacheServer/config"
	"github.com/AlexJudin/DocumentCacheServer/internal/entity"
)

var _ Client = (*CacheClientRepo)(nil)

type CacheClientRepo struct {
	Cfg         *config.Сonfig
	RedisClient *redis.Client
}

func NewCacheClientRepo(cfg *config.Сonfig, redisClient *redis.Client) *CacheClientRepo {
	return &CacheClientRepo{
		Cfg:         cfg,
		RedisClient: redisClient,
	}
}

func (c *CacheClientRepo) Set(ctx context.Context, uuid, mime string, data interface{}, isFile bool) error {
	log.Infof("start set document [%s] to cache", uuid)

	var (
		filePath   string
		file       []byte
		jsonDocMap map[string]interface{}
		err        error
	)

	metadata := make(map[string]interface{})

	switch data.(type) {
	case string:
		filePath = data.(string)
	case map[string]interface{}:
		jsonDocMap = data.(map[string]interface{})
	default:
	}

	if isFile {
		info, err := os.Stat(filePath)
		if err != nil {
			log.Debugf("error getting file info: %+v", err)
			return err
		}

		file, err = os.ReadFile(filePath)
		if err != nil {
			log.Debugf("error reading file: %+v", err)
			return err
		}

		metadata["size"] = info.Size()
	} else {
		result := entity.ApiResponse{
			Data: jsonDocMap,
		}

		file, err = json.Marshal(result)
		if err != nil {
			log.Debugf("error marshalling file: %+v", err)
			return err
		}
	}

	err = c.RedisClient.Set(ctx, "file:data:"+uuid, file, c.Cfg.CacheTTL).Err()
	if err != nil {
		return err
	}

	metadata["type"] = mime
	metadata["created"] = time.Now().Unix()

	err = c.RedisClient.HSet(ctx, "file:meta:"+uuid, metadata).Err()
	if err != nil {
		log.Errorf("error setting metadata: %+v", err)
		return err
	}

	log.Infof("end set document [%s] to cache", uuid)

	return err
}

func (c *CacheClientRepo) Get(ctx context.Context, uuid string) ([]byte, string, bool) {
	log.Infof("start get document [%s] from cache", uuid)

	file, err := c.RedisClient.Get(ctx, "file:data:"+uuid).Bytes()
	if err != nil {
		log.Debugf("error getting file data: %+v", err)
		return nil, "", false
	}

	meta, err := c.RedisClient.HGetAll(ctx, "file:meta:"+uuid).Result()
	if err != nil {
		log.Debugf("error getting file meta: %+v", err)
		return nil, "", false
	}

	mime, ok := meta["type"]
	if !ok {
		log.Debugf("error to get mime: %+v", err)
		return nil, "", false
	}

	log.Infof("end get document [%s] from cache", uuid)

	return file, mime, true
}

func (c *CacheClientRepo) Delete(ctx context.Context, uuid string) error {
	log.Infof("start delete document [%s] from cache", uuid)

	err := c.RedisClient.Del(ctx, uuid).Err()
	if err != nil {
		log.Debugf("failed to delete document: %+v", err)
		return err
	}

	log.Infof("end delete document [%s] from cache", uuid)

	return err
}
