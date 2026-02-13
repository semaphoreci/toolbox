//revive:disable-next-line:var-naming
package metrics

import (
	"fmt"
	"time"
)

const (
	LocalBackend = "local"

	MeasurementName = "usercache"
	CommandStore    = "store"
	CommandRestore  = "restore"
)

type CacheEvent struct {
	Command   string
	Server    string
	User      string
	SizeBytes int64
	Duration  time.Duration
	Corrupt   bool
}

type MetricsManager interface {
	Enabled() bool
	LogEvent(event CacheEvent) error
}

func InitMetricsManager(backend string) (MetricsManager, error) {
	switch backend {
	case LocalBackend:
		return NewLocalMetricsBackend()
	default:
		return nil, fmt.Errorf("metrics backend '%s' is not available", backend)
	}
}
