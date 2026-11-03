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
	metricsManager metrics.MetricsManager
	opts           ArchiverOptions
}

func NewShellOutArchiver(metricsManager metrics.MetricsManager) *ShellOutArchiver {
	return NewShellOutArchiverWithOptions(metricsManager, ArchiverOptions{})
}

func NewShellOutArchiverWithOptions(metricsManager metrics.MetricsManager, opts ArchiverOptions) *ShellOutArchiver {
	return &ShellOutArchiver{metricsManager: metricsManager, opts: opts}
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
	// Only GNU tar can skip existing paths without failing or rewriting their metadata.
	if a.opts.IgnoreCollisions && !isGNUTar() {
		return NewNativeArchiverWithOptions(a.metricsManager, true, a.opts).Decompress(src)
	}

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
		return exec.Command("tar", "czPf", dst, src) // #nosec G204 -- command is literal "tar"; dst/src are cache paths, not user-controlled commands
	}

	return exec.Command("tar", "czf", dst, src) // #nosec G204 -- command is literal "tar"; dst/src are cache paths, not user-controlled commands
}

func (a *ShellOutArchiver) decompressionCmd(dst, tempFile string) *exec.Cmd {
	flags := "xzf"
	if filepath.IsAbs(dst) {
		flags = "xzPf"
	}

	args := []string{flags, tempFile, "-C", "."}
	if a.opts.IgnoreCollisions {
		args = append(args, "--skip-old-files")
	}

	return exec.Command("tar", args...) // #nosec G204 -- command is literal "tar"; tempFile is internal temp path
}

var isGNUTar = sync.OnceValue(func() bool {
	output, err := exec.Command("tar", "--version").Output()
	if err != nil {
		log.Warnf("Could not determine tar version, assuming non-GNU tar: %v", err)
		return false
	}

	return strings.Contains(string(output), "GNU tar")
})

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
