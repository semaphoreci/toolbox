//revive:disable-next-line:var-naming
package metrics

type NoOpMetricsManager struct {
}

func NewNoOpMetricsManager() *NoOpMetricsManager {
	return &NoOpMetricsManager{}
}

func (b *NoOpMetricsManager) Enabled() bool {
	return false
}

func (b *NoOpMetricsManager) LogEvent(event CacheEvent) error {
	return nil
}
