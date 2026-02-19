package metric

import (
	"errors"
	"gorm.io/gorm"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	databaseQueryCount      = "database_query_count"
	databaseQueryDurationMs = "database_query_duration_ms"

	StatusSuccess  = "success"
	StatusFailed   = "failed"
	StatusNotFound = "not_found"
)

// QueryObserver - интерфейс для наблюдения за запросами
type QueryObserver interface {
	Observe(fn func() error, dataBaseType, queryName string) error
}

type DatabaseMetrics struct {
	queryCount    *prometheus.CounterVec
	queryDuration *prometheus.HistogramVec
}

// NewDatabaseMetrics создает новый экземпляр метрик для запроса
func NewDatabaseMetrics() *DatabaseMetrics {
	return &DatabaseMetrics{
		queryCount: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: databaseQueryCount,
				Help: "Total number of database queries",
			},
			[]string{"query_name", "database", "status"},
		),
		queryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    databaseQueryDurationMs,
				Help:    "Duration of database queries in milliseconds",
				Buckets: prometheus.DefBuckets, // можно настроить свои бакеты
			},
			[]string{"query_name", "database", "status"},
		),
	}
}

func (m *DatabaseMetrics) Observe(fn func() error, dataBaseType, queryName string) error {
	startTime := time.Now()

	err := fn()

	// Определяем статус
	status := m.getStatus(err)

	// Увеличиваем счетчик
	m.queryCount.With(prometheus.Labels{
		"query_name": queryName,
		"database":   dataBaseType,
		"status":     status,
	}).Inc()

	// Измеряем длительность в миллисекундах
	durationMs := time.Since(startTime).Seconds() * 1000
	m.queryDuration.With(prometheus.Labels{
		"query_name": queryName,
		"database":   dataBaseType,
		"status":     status,
	}).Observe(durationMs)

	return err
}

func (m *DatabaseMetrics) getStatus(err error) string {
	if err == nil {
		return StatusSuccess
	}

	if m.isNotFoundError(err) {
		return StatusNotFound
	}

	return StatusFailed
}

func (m *DatabaseMetrics) isNotFoundError(err error) bool {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return true
	case errors.Is(err, mongo.ErrNoDocuments):
		return true
	}

	return false
}
