package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/AlexJudin/DocumentCacheServer/internal/controller/common"
)

func WithTimeout(timeout time.Duration, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		// Создаем новый запрос с контекстом с таймаутом
		r = r.WithContext(ctx)

		// Создаем канал для отслеживания завершения обработки
		done := make(chan bool, 1)

		go func() {
			next.ServeHTTP(w, r)
			done <- true
		}()

		select {
		case <-done:
			// Обработчик завершился вовремя
			return
		case <-ctx.Done():
			// Таймаут истек
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				log.Errorf("request timeout: %s %s", r.Method, r.URL.Path)
				common.ApiError(http.StatusRequestTimeout, "Превышено время ожидания выполнения запроса", w)
			}
		}
	}
}
