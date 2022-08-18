package files

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/metrics"
	log "github.com/sirupsen/logrus"
)

func Unpack(metricsManager metrics.MetricsManager, path string) (string, error) {
	restorationPath, err := findRestorationPath(path)
	if err != nil {
		log.Errorf("Could not find restoration path: %v", err)
		if metricErr := metricsManager.Publish(metrics.Metric{Name: metrics.CacheCorruptionRate, Value: "1"}); metricErr != nil {
			log.Errorf("Error publishing %s metric: %v", metrics.CacheCorruptionRate, metricErr)
		}

		return "", err
	}

	cmd := unpackCommand(restorationPath, path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("Unpacking command failed: %s", string(output))
		if metricErr := metricsManager.Publish(metrics.Metric{Name: metrics.CacheCorruptionRate, Value: "1"}); metricErr != nil {
			log.Errorf("Error publishing %s metric: %v", metrics.CacheCorruptionRate, metricErr)
		}

		return "", err
	}

	return restorationPath, nil
}

func unpackCommand(restorationPath, tempFile string) *exec.Cmd {
	if filepath.IsAbs(restorationPath) {
		return exec.Command("tar", "xzPf", tempFile, "-C", ".")
	}

	return exec.Command("tar", "xzf", tempFile, "-C", ".")
}

func findRestorationPath(path string) (string, error) {
	// #nosec
	file, err := os.Open(path)
	if err != nil {
		log.Errorf("error opening %s: %v", path, err)
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
		log.Errorf("Error reading %s: %v", path, err)
		_ = gzipReader.Close()
		return "", err
	}

	return filepath.FromSlash(header.Name), gzipReader.Close()
}
