package archive

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/klauspost/compress/zstd"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/metrics"
	log "github.com/sirupsen/logrus"
)

type ShellOutArchiver struct {
	metricsManager metrics.MetricsManager
}

func NewShellOutArchiver(metricsManager metrics.MetricsManager) *ShellOutArchiver {
	return &ShellOutArchiver{metricsManager: metricsManager}
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
		return exec.Command("tar", "cPf", dst, "--zstd", src)
	}

	return exec.Command("tar", "cf", dst, "--zstd", src)
}

func (a *ShellOutArchiver) decompressionCmd(dst, tempFile string) *exec.Cmd {
	if filepath.IsAbs(dst) {
		return exec.Command("tar", "xPf", tempFile, "-C", ".")
	}

	return exec.Command("tar", "xf", tempFile, "-C", ".")
}

func openDecompressingReader(file *os.File) (io.ReadCloser, error) {
	is_zstd, err := IsZstdCompressed(file)
	if err != nil {
		return nil, err
	}

	if is_zstd {
		var reader *zstd.Decoder
		reader, err = zstd.NewReader(file)
		if err != nil {
			return nil, err
		}
		return reader.IOReadCloser(), nil
	}

	return gzip.NewReader(file)
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

	reader, err := openDecompressingReader(file)
	if err != nil {
		log.Errorf("error creating gzip reader: %v", err)
		return "", err
	}

	tr := tar.NewReader(reader)
	header, err := tr.Next()
	if err == io.EOF {
		log.Warning("No files in archive.")
		_ = reader.Close()
		return "", nil
	}

	if err != nil {
		_ = reader.Close()
		return "", fmt.Errorf("error reading archive %s: %v", src, err)
	}

	return filepath.FromSlash(header.Name), reader.Close()
}
