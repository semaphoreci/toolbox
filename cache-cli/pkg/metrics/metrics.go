package metrics

import "fmt"

type MetricsManager interface {
	Enabled() bool
	Publish(metric Metric) error
	PublishBatch(metrics []Metric) error
}

type Metric struct {
	Name  string
	Value string
}

func InitMetricsManager(backendType string) (MetricsManager, error) {
	switch backendType {
	case "local":
		return NewLocalMetricsBackend()
	default:
		return nil, fmt.Errorf("metrics backend type '%s' is not available", backendType)
	}
}
