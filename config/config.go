package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

const (
	accessTokenTTLDefault  = 5
	refreshTokenTTLDefault = 60

	cacheTTLDefault = 15
)

type Config struct {
	Host     string
	Port     string
	LogLevel log.Level
	*ConfigDB
	*ConfigMongoDB
	*ConfigAuth
	*ConfigRedis
	*ConfigFileStorage
	*ConfigMinio
}

type ConfigDB struct {
	Port     string
	User     string
	Password string
	DBName   string
	Sslmode  string
}

type ConfigMongoDB struct {
	Host string
	Port string
}

type ConfigAuth struct {
	AdminToken      string
	PasswordSalt    []byte
	TokenSalt       []byte
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type ConfigRedis struct {
	Host     string
	Port     string
	Password string
	CacheTTL time.Duration
}

type ConfigFileStorage struct {
	MainDir string
}

type ConfigMinio struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
}

func New() (*Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, err
	}

	cfg := Config{
		Host: os.Getenv("HOST"),
		Port: os.Getenv("PORT"),
	}

	logLevel, err := log.ParseLevel(os.Getenv("LOGLEVEL"))
	if err != nil {
		return nil, err
	}

	cfg.LogLevel = logLevel

	dbCfg := ConfigDB{
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
	}
	cfg.ConfigDB = &dbCfg

	mgDbCfg := ConfigMongoDB{
		Host: os.Getenv("MGDB_HOST"),
		Port: os.Getenv("MGDB_PORT"),
	}
	cfg.ConfigMongoDB = &mgDbCfg

	accessTokenTTL, err := time.ParseDuration(os.Getenv("ACCESS_TOKEN_TTL"))
	if err != nil {
		accessTokenTTL = accessTokenTTLDefault
	}

	refreshTokenTTL, err := time.ParseDuration(os.Getenv("REFRESH_TOKEN_TTL"))
	if err != nil {
		refreshTokenTTL = refreshTokenTTLDefault
	}

	authCfg := ConfigAuth{
		AdminToken:      os.Getenv("ADMIN_TOKEN"),
		PasswordSalt:    []byte(os.Getenv("PASSWORD_SALT")),
		TokenSalt:       []byte(os.Getenv("TOKEN_SALT")),
		AccessTokenTTL:  accessTokenTTL * time.Minute,
		RefreshTokenTTL: refreshTokenTTL * time.Minute,
	}
	cfg.ConfigAuth = &authCfg

	cacheTTL, err := time.ParseDuration(os.Getenv("CACHE_TTL"))
	if err != nil {
		cacheTTL = cacheTTLDefault
	}

	redisCfg := ConfigRedis{
		Host:     os.Getenv("REDIS_HOST"),
		Port:     os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		CacheTTL: cacheTTL * time.Minute,
	}
	cfg.ConfigRedis = &redisCfg

	fileCfg := ConfigFileStorage{
		MainDir: os.Getenv("FILE_MAIN_DIR"),
	}
	cfg.ConfigFileStorage = &fileCfg

	cfg.ConfigMinio = &ConfigMinio{
		Endpoint:        getMinioEndpoint(),
		AccessKeyID:     os.Getenv("MINIO_ROOT_USER"),
		SecretAccessKey: os.Getenv("MINIO_ROOT_PASSWORD"),
		UseSSL:          false,
	}

	return &cfg, nil
}

func (c *Config) GetDataSourceName() string {
	str := fmt.Sprintf("host=localhost port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.ConfigDB.Port, c.ConfigDB.User, c.ConfigDB.Password, c.ConfigDB.DBName)

	return str
}

func (c *Config) GetMongoDBSourse() string {
	str := fmt.Sprintf("mongodb://%s:%s", c.ConfigMongoDB.Host, c.ConfigMongoDB.Port)

	return str
}

func getMinioEndpoint() string {
	return fmt.Sprintf("%s:%s", os.Getenv("MINIO_ROOT_HOST"), os.Getenv("MINIO_ROOT_PORT"))
}
