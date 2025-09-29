package domain

import (
	"gorm.io/gorm"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/AlexJudin/DocumentCacheServer/config"
	"github.com/AlexJudin/DocumentCacheServer/internal/api/domain/auth"
	"github.com/AlexJudin/DocumentCacheServer/internal/api/domain/document"
	"github.com/AlexJudin/DocumentCacheServer/internal/api/domain/register"
	"github.com/AlexJudin/DocumentCacheServer/internal/repository/cache"
	"github.com/AlexJudin/DocumentCacheServer/internal/repository/mongodb"
	"github.com/AlexJudin/DocumentCacheServer/internal/repository/postgres"
	"github.com/AlexJudin/DocumentCacheServer/internal/service"
	"github.com/AlexJudin/DocumentCacheServer/internal/usecases"
)

func AddRoutes(config *config.Сonfig,
	db *gorm.DB,
	mgDbClient *mongo.Client,
	redisClient *redis.Client,
	r *chi.Mux) {
	// init services
	authService := service.NewAuthService(config, db)

	// init postgres repository
	repoDocument := postgres.NewDocumentRepo(db)
	repoUser := postgres.NewUserRepo(db)

	// init cacheClient
	cacheClient := cache.NewCacheClientRepo(redisClient)

	// init mongoDB repository
	repoJson := mongodb.NewMongoDBRepo(mgDbClient)

	// init usecases
	docsUC := usecases.NewDocumentUsecase(config, repoDocument, cacheClient, repoJson)
	docsHandler := document.NewDocumentHandler(docsUC)

	registerUC := usecases.NewRegisterUsecase(repoUser, authService)
	registerHandler := register.NewRegisterHandler(registerUC)

	authUC := usecases.NewAuthUsecase(repoUser, authService)
	authHandler := auth.NewAuthHandler(authUC)

	// init middleware
	//authMiddleware := middleware.NewAuthMiddleware(authService)

	r.Post("/api/register", registerHandler.RegisterUser)
	r.Post("/api/auth", authHandler.AuthorizationUser)
	r.Post("/api/refresh-token", authHandler.RefreshToken)
	r.Delete("/api/auth", authHandler.DeleteToken)

	r.Group(func(r chi.Router) {
		r.Use(httprate.LimitByIP(5000, time.Second))
		//r.Use(authMiddleware.CheckToken)
		r.Post("/api/docs", docsHandler.SaveDocument)

		r.Get("/api/docs", docsHandler.GetDocumentsList)
		r.Head("/api/docs", docsHandler.GetDocumentsList)

		r.Get("/api/docs/", docsHandler.GetDocumentById)
		r.Head("/api/docs/", docsHandler.GetDocumentById)

		r.Delete("/api/docs/", docsHandler.DeleteDocumentById)
	})
}
