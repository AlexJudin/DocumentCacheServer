package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
)

const httpServerRequestsSecondsCount = "http_server_requests_seconds_count"

type HTTPMetrics struct {
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

func NewHTTPMetrics() *HTTPMetrics {
	return &HTTPMetrics{
		requestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    httpServerRequestsSecondsCount,
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "uri", "status"},
		),
	}
}

func (m *HTTPMetrics) HTTPMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapper := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapper, r)

		duration := time.Since(start).Seconds()

		path := r.URL.Path

		statusStr := strconv.Itoa(wrapper.statusCode)

		m.requestDuration.WithLabelValues(r.Method, path, statusStr).Observe(duration)

		log.Printf("Method: %s, URI: %s, Status: %d, Duration: %f",
			r.Method, path, wrapper.statusCode, duration)
	})
}
