package metric

import (
	"runtime/debug"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const applicationInfo = "application_info"

type AppMetrics struct {
	info *prometheus.GaugeVec
}

// NewAppMetrics создает метрики с автоматическим сбором информации
func NewAppMetrics() *AppMetrics {
	info := promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: applicationInfo,
			Help: "Information about the application version, commit hash and build date",
		},
		[]string{"version", "commit", "date_build"},
	)

	version, commit, buildDate := extractBuildInfo()

	info.WithLabelValues(version, commit, buildDate).Set(1)

	return &AppMetrics{
		info: info,
	}
}

func extractBuildInfo() (version, commit, buildDate string) {
	version = "unknown"
	commit = "unknown"
	buildDate = time.Now().Format("2006-01-02_15:04:05")

	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" {
			version = info.Main.Version
		}

		// Ищем vcs информацию в настройках
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if len(setting.Value) >= 7 {
					commit = setting.Value[:7]
				} else {
					commit = setting.Value
				}
			case "vcs.time":
				if t, err := time.Parse(time.RFC3339, setting.Value); err == nil {
					buildDate = t.Format("2006-01-02_15:04:05")
				}
			}
		}
	}

	return version, commit, buildDate
}
