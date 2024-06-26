package archive

import (
	"context"
	"os"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/metrics"
)

type Archiver interface {
	Compress(ctx context.Context, dst, src string) error
	Decompress(ctx context.Context, src string) (string, error)
}

func NewArchiver(metricsManager metrics.MetricsManager) Archiver {
	method := os.Getenv("SEMAPHORE_CACHE_ARCHIVE_METHOD")
	switch method {
	case "native":
		return NewNativeArchiver(metricsManager, false)
	case "native-parallel":
		return NewNativeArchiver(metricsManager, true)
	default:
		return NewShellOutArchiver(metricsManager)
	}
}
