package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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

func (c *CacheClientRepo) Set(ctx context.Context, key, mime string, data interface{}, isFile bool) error {
	log.Infof("start set document [%s] to cache", key)

	var (
		filePath string
		file     []byte
		jsonFile map[string]interface{}
	)

	fileInfo := make(map[string]interface{})

	switch data.(type) {
	case string:
		filePath = data.(string)
	case map[string]interface{}:
		jsonFile = data.(map[string]interface{})
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

		fileInfo["filename"] = filepath.Base(filePath)
		fileInfo["size"] = info.Size()
		fileInfo["modified"] = info.ModTime().Unix()
		fileInfo["data"] = file
	} else {
		fileInfo["data"] = jsonFile
	}

	fileInfo["isFile"] = isFile
	fileInfo["mime"] = mime

	dataByte, err := json.Marshal(fileInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %v", err)
	}

	log.Infof("end set document [%s] to cache", key)

	return c.RedisClient.Set(ctx, key, dataByte, c.Cfg.CacheTTL).Err()
}

func (c *CacheClientRepo) Get(ctx context.Context, key string) ([]byte, string, bool) {
	log.Infof("start get document [%s] from cache", key)

	var fileInfo map[string]interface{}

	data, err := c.RedisClient.Get(ctx, key).Bytes()
	if err != nil {
		log.Debugf("failed to get file info: %+v", err)
		return nil, "", false
	}

	err = json.Unmarshal(data, &fileInfo)
	if err != nil {
		log.Debugf("failed to unmarshal fileInfo: %+v", err)
		return nil, "", false
	}

	mime, ok := fileInfo["mime"].(string)
	if !ok {
		log.Debugf("failed to get file info: %+v", fileInfo)
		return nil, "", false
	}

	switch fileInfo["data"].(type) {
	case string:
		file, ok := fileInfo["data"].(string)
		if !ok {
			log.Debugf("failed to get file info: %+v", fileInfo)
			return nil, "", false
		}

		return []byte(file), mime, true
	case map[string]interface{}:
		jsonDocMap, ok := fileInfo["data"].(map[string]interface{})
		if !ok {
			log.Debugf("failed to get file info: %+v", fileInfo)
		}

		result := entity.ApiResponse{
			Data: jsonDocMap,
		}

		jsonDoc, err := json.Marshal(result)
		if err != nil {
			return nil, "", false
		}

		return jsonDoc, mime, true
	}

	log.Infof("end get document [%s] from cache", key)

	return nil, "", false
}

func (c *CacheClientRepo) Delete(ctx context.Context, key string) error {
	log.Infof("start delete document [%s] from cache", key)

	err := c.RedisClient.Del(ctx, key).Err()
	if err != nil {
		log.Debugf("failed to delete document: %+v", err)
		return err
	}

	log.Infof("end delete document [%s] from cache", key)

	return err
}
