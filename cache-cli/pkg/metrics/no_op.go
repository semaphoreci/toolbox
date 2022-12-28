package metrics

type NoOpMetricsManager struct {
}

func NewNoOpMetricsManager() *NoOpMetricsManager {
	return &NoOpMetricsManager{}
}

func (b *NoOpMetricsManager) Enabled() bool {
	return false
}

func (b *NoOpMetricsManager) PublishBatch(metrics []Metric) error {
	return nil
}

func (b *NoOpMetricsManager) Publish(metric Metric) error {
	return nil
}
