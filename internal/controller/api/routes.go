package api

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"

	"github.com/AlexJudin/DocumentCacheServer/config"
	"github.com/AlexJudin/DocumentCacheServer/internal/controller/api/auth"
	"github.com/AlexJudin/DocumentCacheServer/internal/controller/api/document"
	"github.com/AlexJudin/DocumentCacheServer/internal/controller/api/register"
	"github.com/AlexJudin/DocumentCacheServer/internal/controller/middleware"
	"github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository"
	"github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository/cache"
	"github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository/postgres"
	"github.com/AlexJudin/DocumentCacheServer/internal/infrastructure/repository/saga"
	"github.com/AlexJudin/DocumentCacheServer/internal/service"
	"github.com/AlexJudin/DocumentCacheServer/internal/usecases"
)

func AddRoutes(cfg *config.Config,
	documentRepo *repository.DocumentRepo,
	userRepo *postgres.UserRepo,
	tokenRepo *postgres.TokenStorageRepo,
	cacheRepo *cache.DocumentRepo,
	sagaOrchestrator *saga.DocumentOrchestrator,
	r *chi.Mux) {
	// init services
	authService := service.NewAuthService(cfg, tokenRepo)

	// init usecases
	docsUC := usecases.NewDocumentUsecase(documentRepo, cacheRepo, sagaOrchestrator)
	docsHandler := document.NewDocumentHandler(docsUC)

	registerUC := usecases.NewRegisterUsecase(userRepo, authService)
	registerHandler := register.NewRegisterHandler(registerUC)

	authUC := usecases.NewAuthUsecase(userRepo, authService)
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
