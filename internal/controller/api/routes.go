package api

import (
	"gorm.io/gorm"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/AlexJudin/DocumentCacheServer/config"
	"github.com/AlexJudin/DocumentCacheServer/internal/controller/api/auth"
	"github.com/AlexJudin/DocumentCacheServer/internal/controller/api/document"
	"github.com/AlexJudin/DocumentCacheServer/internal/controller/api/register"
	"github.com/AlexJudin/DocumentCacheServer/internal/controller/middleware"
	"github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository"
	"github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository/cache"
	"github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository/postgres"
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

	// init auth middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	//init timeout middleware
	timeoutMiddleware := middleware.NewTimeoutMiddleware(time.Second)

	r.Post("/api/register", registerHandler.RegisterUser)
	r.Post("/api/auth", authHandler.AuthorizationUser)
	r.Post("/api/refresh-token", authHandler.RefreshToken)
	r.Delete("/api/auth", authHandler.DeleteToken)

	r.Group(func(r chi.Router) {
		r.Use(httprate.LimitByIP(5000, time.Second))
		r.Use(
			authMiddleware.CheckToken,
			timeoutMiddleware.WithTimeout,
		)
		r.Post("/api/docs", docsHandler.SaveDocument)

		r.Get("/api/docs", docsHandler.GetDocumentsList)
		r.Head("/api/docs", docsHandler.GetDocumentsList)

		r.Get("/api/docs/", docsHandler.GetDocumentById)
		r.Head("/api/docs/", docsHandler.GetDocumentById)

		r.Delete("/api/docs/", docsHandler.DeleteDocumentById)
	})
}
