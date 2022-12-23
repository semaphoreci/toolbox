package metrics

type NoOpMetricsManager struct {
}

func NewNoOpMetricsBackend() *LocalMetricsManager {
	return &LocalMetricsManager{}
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
