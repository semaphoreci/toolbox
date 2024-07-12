package archive

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/metrics"
	log "github.com/sirupsen/logrus"
)

type ShellOutArchiver struct {
	metricsManager metrics.MetricsManager
}

func NewShellOutArchiver(metricsManager metrics.MetricsManager) *ShellOutArchiver {
	return &ShellOutArchiver{metricsManager: metricsManager}
}

func (a *ShellOutArchiver) Compress(ctx context.Context, dst, src string) error {
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("error finding '%s': %v", src, err)
	}

	cmd := a.compressionCommand(ctx, dst, src)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error compressing %s: %s, %v", src, output, err)
	}

	return nil
}

func (a *ShellOutArchiver) Decompress(ctx context.Context, src string) (string, error) {
	restorationPath, err := a.findRestorationPath(src)
	if err != nil {
		if metricErr := a.metricsManager.Publish(metrics.Metric{Name: metrics.CacheCorruptionRate, Value: "1"}); metricErr != nil {
			log.Errorf("Error publishing %s metric: %v", metrics.CacheCorruptionRate, metricErr)
		}

		return "", fmt.Errorf("error finding restoration path: %v", err)
	}

	cmd := a.decompressionCmd(ctx, restorationPath, src)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if metricErr := a.metricsManager.Publish(metrics.Metric{Name: metrics.CacheCorruptionRate, Value: "1"}); metricErr != nil {
			log.Errorf("Error publishing %s metric: %v", metrics.CacheCorruptionRate, metricErr)
		}

		return "", fmt.Errorf("error executing decompression command: %s, %v", string(output), err)
	}

	return restorationPath, nil
}

func (a *ShellOutArchiver) compressionCommand(ctx context.Context, dst, src string) *exec.Cmd {
	if filepath.IsAbs(src) {
		return exec.CommandContext(ctx, "tar", "czPf", dst, src)
	}

	return exec.CommandContext(ctx, "tar", "czf", dst, src)
}

func (a *ShellOutArchiver) decompressionCmd(ctx context.Context, dst, tempFile string) *exec.Cmd {
	if filepath.IsAbs(dst) {
		return exec.CommandContext(ctx, "tar", "xzPf", tempFile, "-C", ".")
	}

	return exec.CommandContext(ctx, "tar", "xzf", tempFile, "-C", ".")
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
