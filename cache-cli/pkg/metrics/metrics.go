package metrics

import "fmt"

const LocalBackend = "local"
const CacheDownloadSize = "cache_download_size"
const CacheDownloadTime = "cache_download_time"
const CacheUser = "cache_user"
const CacheServer = "cache_server"
const CacheTotalRate = "cache_total_rate"
const CacheCorruptionRate = "cache_corruption_rate"

type MetricsManager interface {
	Enabled() bool
	Publish(metric Metric) error
	PublishBatch(metrics []Metric) error
}

type Metric struct {
	Name  string
	Value string
}

func InitMetricsManager(backend string) (MetricsManager, error) {
	switch backend {
	case LocalBackend:
		return NewLocalMetricsBackend()
	default:
		return nil, fmt.Errorf("metrics backend '%s' is not available", backend)
	}
}
