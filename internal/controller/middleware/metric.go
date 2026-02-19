package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
)

type MetricsHTTP struct {
	requestDuration *prometheus.HistogramVec
}

// Обертка для ResponseWriter
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func NewMetricsHTTP() *MetricsHTTP {
	return &MetricsHTTP{
		// Гистограмма длительности запросов
		requestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_server_requests_seconds_count",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "uri", "status"},
		),
	}
}

func (m *MetricsHTTP) MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Замеряем время начала
		start := time.Now()

		// Создаем обертку для ResponseWriter, чтобы перехватить статус и размер
		wrapper := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Обрабатываем запрос
		next.ServeHTTP(wrapper, r)

		// Вычисляем длительность
		duration := time.Since(start).Seconds()

		// Получаем путь запроса
		path := r.URL.Path

		// Записываем метрики
		statusStr := strconv.Itoa(wrapper.statusCode)

		// Длительность запроса
		m.requestDuration.WithLabelValues(r.Method, path, statusStr).Observe(duration)

		// Логирование для отладки
		log.Printf("Method: %s, URI: %s, Status: %d, Duration: %f",
			r.Method, path, wrapper.statusCode, duration)
	})
}
