package metrics

import (
	"fmt"
	"os"
	"runtime"
)

type LocalMetricsManager struct {
	ToolboxMetricsPath string
	CacheMetricsPath   string
}

func NewLocalMetricsBackend() (*LocalMetricsManager, error) {
	basePath := "/tmp"
	if runtime.GOOS == "windows" {
		basePath = os.TempDir()
	}

	return &LocalMetricsManager{
		ToolboxMetricsPath: fmt.Sprintf("%s/toolbox_metrics", basePath),
		CacheMetricsPath:   fmt.Sprintf("%s/cache_metrics", basePath),
	}, nil
}

func (b *LocalMetricsManager) Enabled() bool {
	return os.Getenv("SEMAPHORE_TOOLBOX_METRICS_ENABLED") == "true"
}

func (b *LocalMetricsManager) PublishBatch(metrics []Metric) error {
	if !b.Enabled() {
		return nil
	}

	for _, metric := range metrics {
		err := b.Publish(metric)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *LocalMetricsManager) Publish(metric Metric) error {
	if !b.Enabled() {
		return nil
	}

	switch metric.Name {
	case CacheDownloadSize, CacheDownloadTime, CacheUser, CacheServer:
		return publishMetricToFile(b.CacheMetricsPath, metric.Name, metric.Value)
	case CacheTotalRate, CacheCorruptionRate:
		return publishMetricToFile(b.ToolboxMetricsPath, metric.Name, metric.Value)
	}

	fmt.Printf("Ignoring metric %s\n", metric.Name)
	return nil
}

func publishMetricToFile(file, metricName, metricValue string) error {
	// #nosec
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	line := fmt.Sprintf("%s %s\n", metricName, metricValue)

	_, err = f.WriteString(line)
	if err != nil {
		_ = f.Close()
		return err
	}

	return f.Close()
}
