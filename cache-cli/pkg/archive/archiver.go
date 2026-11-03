package archive

import (
	"os"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/metrics"
)

type Archiver interface {
	Compress(dst, src string) error
	Decompress(src string) (string, error)
}

type ArchiverOptions struct {
	// IgnoreCollisions skips extracting entries whose paths already exist, keeping the local copy.
	IgnoreCollisions bool
}

func NewArchiver(metricsManager metrics.MetricsManager) Archiver {
	return NewArchiverWithOptions(metricsManager, ArchiverOptions{})
}

func NewArchiverWithOptions(metricsManager metrics.MetricsManager, opts ArchiverOptions) Archiver {
	method := os.Getenv("SEMAPHORE_CACHE_ARCHIVE_METHOD")
	switch method {
	case "native":
		return NewNativeArchiverWithOptions(metricsManager, false, opts)
	case "native-parallel":
		return NewNativeArchiverWithOptions(metricsManager, true, opts)
	default:
		return NewShellOutArchiverWithOptions(metricsManager, opts)
	}
}
