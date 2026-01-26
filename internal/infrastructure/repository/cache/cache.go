package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"

	"github.com/AlexJudin/DocumentCacheServer/config"
	"github.com/AlexJudin/DocumentCacheServer/internal/entity"
)

var _ Document = (*DocumentRepo)(nil)

type DocumentRepo struct {
	Cfg         *config.Config
	RedisClient *redis.Client
}

func NewDocumentRepo(cfg *config.Config, redisClient *redis.Client) *DocumentRepo {
	return &DocumentRepo{
		Cfg:         cfg,
		RedisClient: redisClient,
	}
}

func (r *DocumentRepo) Set(ctx context.Context, uuid, mime string, data interface{}, isFile bool) {
	log.Infof("setting saga [%s] to cache", uuid)

	var (
		file       []byte
		jsonDocMap map[string]interface{}
		err        error
	)

	metadata := make(map[string]interface{})

	switch value := data.(type) {
	case []byte:
		file = value
	case map[string]interface{}:
		jsonDocMap = value
	}

	if isFile {
		metadata["size"] = len(file)
	} else {
		result := entity.ApiResponse{
			Data: jsonDocMap,
		}

		file, err = json.Marshal(result)
		if err != nil {
			log.Debugf("failed to marshal saga json: %+v", err)
			return
		}
	}

	err = r.RedisClient.Set(ctx, "file:data:"+uuid, file, r.Cfg.CacheTTL).Err()
	if err != nil {
		log.Debugf("failed to store saga data in cache: %+v", err)
		return
	}

	metadata["type"] = mime
	metadata["created"] = time.Now().Unix()

	err = r.RedisClient.HSet(ctx, "file:meta:"+uuid, metadata).Err()
	if err != nil {
		log.Debugf("failed to store saga metadata in cache: %+v", err)
		return
	}

	log.Infof("saga [%s] successfully cached", uuid)
}

func (r *DocumentRepo) Get(ctx context.Context, uuid string) ([]byte, string, bool) {
	log.Infof("retrieving saga [%s] from cache", uuid)

	file, err := r.RedisClient.Get(ctx, "file:data:"+uuid).Bytes()
	if err != nil {
		log.Debugf("failed to retrieve saga data from cache: %+v", err)
		return nil, "", false
	}

	meta, err := r.RedisClient.HGetAll(ctx, "file:meta:"+uuid).Result()
	if err != nil {
		log.Debugf("failed to retrieve saga metadata from cache: %+v", err)
		return nil, "", false
	}

	mime, ok := meta["type"]
	if !ok {
		log.Debug("saga metadata missing MIME type")
		return nil, "", false
	}

	log.Infof("saga [%s] successfully retrieved from cache", uuid)

	return file, mime, true
}

func (r *DocumentRepo) Delete(ctx context.Context, uuid string) {
	log.Infof("deleting saga [%s] from cache", uuid)

	err := r.RedisClient.Del(ctx, uuid).Err()
	if err != nil {
		log.Debugf("failed to delete saga from cache: %+v", err)
		return
	}

	log.Infof("saga [%s] successfully deleted from cache", uuid)
}
