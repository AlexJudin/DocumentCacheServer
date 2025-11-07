package http

import (
	"gorm.io/gorm"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/AlexJudin/DocumentCacheServer/config"
	"github.com/AlexJudin/DocumentCacheServer/internal/controller/http/auth"
	"github.com/AlexJudin/DocumentCacheServer/internal/controller/http/document"
	"github.com/AlexJudin/DocumentCacheServer/internal/controller/http/register"
	"github.com/AlexJudin/DocumentCacheServer/internal/controller/middleware"
	"github.com/AlexJudin/DocumentCacheServer/internal/repository"
	"github.com/AlexJudin/DocumentCacheServer/internal/repository/cache"
	"github.com/AlexJudin/DocumentCacheServer/internal/repository/postgres"
	"github.com/AlexJudin/DocumentCacheServer/internal/service"
	"github.com/AlexJudin/DocumentCacheServer/internal/usecases"
)

func AddRoutes(config *config.Config,
	db *gorm.DB,
	mgDbClient *mongo.Client,
	redisClient *redis.Client,
	fileStorageClient *minio.Client,
	r *chi.Mux) {
	// init services
	authService := service.NewAuthService(config, db)

	// init postgres repository
	repoDocument := repository.NewDocumentRepo(db, mgDbClient, fileStorageClient)
	repoUser := postgres.NewUserRepo(db)

	// init cacheClient
	cacheClient := cache.NewDocumentRepo(config, redisClient)

	// init usecases
	docsUC := usecases.NewDocumentUsecase(repoDocument, cacheClient)
	docsHandler := document.NewDocumentHandler(docsUC)

	registerUC := usecases.NewRegisterUsecase(repoUser, authService)
	registerHandler := register.NewRegisterHandler(registerUC)

	authUC := usecases.NewAuthUsecase(repoUser, authService)
	authHandler := auth.NewAuthHandler(authUC)

	// init middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	r.Post("/api/register", registerHandler.RegisterUser)
	r.Post("/api/auth", authHandler.AuthorizationUser)
	r.Post("/api/refresh-token", authHandler.RefreshToken)
	r.Delete("/api/auth", authHandler.DeleteToken)

	r.Group(func(r chi.Router) {
		r.Use(httprate.LimitByIP(5000, time.Second))
		r.Use(authMiddleware.CheckToken)
		r.Post("/http/docs", docsHandler.SaveDocument)

		r.Get("/http/docs", docsHandler.GetDocumentsList)
		r.Head("/http/docs", docsHandler.GetDocumentsList)

		r.Get("/http/docs/", docsHandler.GetDocumentById)
		r.Head("/http/docs/", docsHandler.GetDocumentById)

		r.Delete("/http/docs/", docsHandler.DeleteDocumentById)
	})
}
