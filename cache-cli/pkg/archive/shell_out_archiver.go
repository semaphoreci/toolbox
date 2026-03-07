package archive

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/metrics"
	log "github.com/sirupsen/logrus"
)

type ShellOutArchiver struct {
	metricsManager   metrics.MetricsManager
	ignoreCollisions bool
}

func NewShellOutArchiver(metricsManager metrics.MetricsManager) *ShellOutArchiver {
	return &ShellOutArchiver{metricsManager: metricsManager}
}

func NewShellOutArchiverWithOptions(metricsManager metrics.MetricsManager, opts ArchiverOptions) *ShellOutArchiver {
	return &ShellOutArchiver{
		metricsManager:   metricsManager,
		ignoreCollisions: opts.IgnoreCollisions,
	}
}

func (a *ShellOutArchiver) Compress(dst, src string) error {
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("error finding '%s': %v", src, err)
	}

	cmd := a.compressionCommand(dst, src)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error compressing %s: %s, %v", src, output, err)
	}

	return nil
}

func (a *ShellOutArchiver) Decompress(src string) (string, error) {
	restorationPath, err := a.findRestorationPath(src)
	if err != nil {
		if metricErr := a.metricsManager.LogEvent(metrics.CacheEvent{Command: metrics.CommandRestore, Corrupt: true}); metricErr != nil {
			log.Errorf("Error publishing corruption metric: %v", metricErr)
		}

		return "", fmt.Errorf("error finding restoration path: %v", err)
	}

	cmd := a.decompressionCmd(restorationPath, src)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if metricErr := a.metricsManager.LogEvent(metrics.CacheEvent{Command: metrics.CommandRestore, Corrupt: true}); metricErr != nil {
			log.Errorf("Error publishing corruption metric: %v", metricErr)
		}

		return "", fmt.Errorf("error executing decompression command: %s, %v", string(output), err)
	}

	return restorationPath, nil
}

func (a *ShellOutArchiver) compressionCommand(dst, src string) *exec.Cmd {
	if filepath.IsAbs(src) {
		return exec.Command("tar", "czPf", dst, src)
	}

	return exec.Command("tar", "czf", dst, src)
}

// decompressionCmd builds the tar extraction command.
// When ignoreCollisions is enabled, GNU tar uses --skip-old-files (silently skips, exit 0),
// while BSD tar uses -k (skips but may return non-zero on some systems).
func (a *ShellOutArchiver) decompressionCmd(dst, tempFile string) *exec.Cmd {
	if filepath.IsAbs(dst) {
		if a.ignoreCollisions {
			if isGNUTar() {
				return exec.Command("tar", "xzPf", tempFile, "-C", ".", "--skip-old-files")
			}
			return exec.Command("tar", "xzPf", tempFile, "-C", ".", "-k")
		}
		return exec.Command("tar", "xzPf", tempFile, "-C", ".")
	}

	if a.ignoreCollisions {
		if isGNUTar() {
			return exec.Command("tar", "xzf", tempFile, "-C", ".", "--skip-old-files")
		}
		return exec.Command("tar", "xzf", tempFile, "-C", ".", "-k")
	}
	return exec.Command("tar", "xzf", tempFile, "-C", ".")
}

var (
	gnuTarOnce   sync.Once
	gnuTarCached bool
)

// isGNUTar returns true if the system tar is GNU tar.
// GNU tar includes "GNU tar" in its --version output.
// The result is cached to avoid repeated subprocess calls.
// If tar --version fails, it defaults to false (assumes BSD tar).
func isGNUTar() bool {
	gnuTarOnce.Do(func() {
		cmd := exec.Command("tar", "--version")
		output, err := cmd.Output()
		if err != nil {
			log.Warnf("Could not determine tar version, assuming BSD tar: %v", err)
			gnuTarCached = false
			return
		}
		gnuTarCached = strings.Contains(string(output), "GNU tar")
	})
	return gnuTarCached
}

func (a *ShellOutArchiver) findRestorationPath(src string) (string, error) {
	// #nosec
	file, err := os.Open(src)
	if err != nil {
		log.Errorf("error opening %s: %v", src, err)
		return "", err
	}

	// #nosec
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		log.Errorf("error creating gzip reader: %v", err)
		return "", err
	}

	tr := tar.NewReader(gzipReader)
	header, err := tr.Next()
	if err == io.EOF {
		log.Warning("No files in archive.")
		_ = gzipReader.Close()
		return "", nil
	}

	if err != nil {
		_ = gzipReader.Close()
		return "", fmt.Errorf("error reading archive %s: %v", src, err)
	}

	return filepath.FromSlash(header.Name), gzipReader.Close()
}
