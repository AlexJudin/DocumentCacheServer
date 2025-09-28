package middleware

import (
	"context"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/AlexJudin/DocumentCacheServer/internal/api/common"
	"github.com/AlexJudin/DocumentCacheServer/internal/service"
)

type AuthMiddleware struct {
	authService service.AuthService
}

func NewAuthMiddleware(authService service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

func (a *AuthMiddleware) CheckToken(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		accessToken, err := r.Cookie("accessToken")
		if err != nil {
			common.ApiError(http.StatusUnauthorized, "access token not found", w)
			return
		}

		userLogin, err := a.authService.VerifyUser(accessToken.Value)
		if err != nil {
			common.ApiError(http.StatusUnauthorized, err.Error(), w)
			return
		}

		log.Infof("Пользователь %s сделал запрос %s", userLogin, r.URL.Path)

		ctx := context.WithValue(r.Context(), "currentUser", userLogin)
		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}
