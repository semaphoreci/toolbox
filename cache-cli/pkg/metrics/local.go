package metrics

import (
	"fmt"
	"os"
)

type LocalMetricsManager struct {
	ToolboxMetricsPath string
	CacheMetricsPath   string
}

func NewLocalMetricsBackend() (*LocalMetricsManager, error) {
	return &LocalMetricsManager{
		ToolboxMetricsPath: "/tmp/toolbox_metrics",
		CacheMetricsPath:   "/tmp/cache_metrics",
	}, nil
}

func (b *LocalMetricsManager) Enabled() bool {
	return os.Getenv("SEMAPHORE_EXECUTION_ENVIRONMENT") == "hosted"
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
	case "cache_download_size", "cache_download_time", "cache_user", "cache_server":
		return publishMetricToFile(b.CacheMetricsPath, metric.Name, metric.Value)
	case "cache_total_rate", "cache_corruption_rate":
		return publishMetricToFile(b.ToolboxMetricsPath, metric.Name, metric.Value)
	}

	fmt.Printf("Ignoring metric %s\n", metric.Name)
	return nil
}

func publishMetricToFile(file, metricName, metricValue string) error {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	line := fmt.Sprintf("%s %s\n", metricName, metricValue)

	_, err = f.WriteString(line)
	return err
}