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

	"github.com/semaphoreci/toolbox/cache-cli/pkg/metrics"
	log "github.com/sirupsen/logrus"
)

type ShellOutArchiver struct {
	metricsManager metrics.MetricsManager
	skipExisting   bool
}

func NewShellOutArchiver(metricsManager metrics.MetricsManager) *ShellOutArchiver {
	return &ShellOutArchiver{metricsManager: metricsManager}
}

func NewShellOutArchiverWithOptions(metricsManager metrics.MetricsManager, opts ArchiverOptions) *ShellOutArchiver {
	return &ShellOutArchiver{
		metricsManager: metricsManager,
		skipExisting:   opts.SkipExisting,
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

func (a *ShellOutArchiver) decompressionCmd(dst, tempFile string) *exec.Cmd {
	args := []string{}

	if filepath.IsAbs(dst) {
		args = append(args, "xzPf", tempFile, "-C", ".")
	} else {
		args = append(args, "xzf", tempFile, "-C", ".")
	}

	// When skipExisting is enabled, skip existing files without overwriting.
	// GNU tar uses --skip-old-files (silent, exit 0).
	// BSD tar uses -k (silent, exit 0).
	if a.skipExisting {
		if isGNUTar() {
			args = append(args, "--skip-old-files")
		} else {
			args = append(args, "-k")
		}
	}

	return exec.Command("tar", args...)
}

// isGNUTar returns true if the system tar is GNU tar.
// GNU tar includes "GNU tar" in its --version output.
func isGNUTar() bool {
	cmd := exec.Command("tar", "--version")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "GNU tar")
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
