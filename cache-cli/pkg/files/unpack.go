package files

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/metrics"
)

func Unpack(metricsManager metrics.MetricsManager, path string) (string, error) {
	restorationPath, err := findRestorationPath(path)
	if err != nil {
		if metricErr := metricsManager.Publish(metrics.Metric{Name: metrics.CacheCorruptionRate, Value: "1"}); metricErr != nil {
			fmt.Printf("Error publishing %s metric: %v\n", metrics.CacheCorruptionRate, metricErr)
		}

		return "", err
	}

	cmd := unpackCommand(restorationPath, path)
	_, err = cmd.Output()
	if err != nil {
		if metricErr := metricsManager.Publish(metrics.Metric{Name: metrics.CacheCorruptionRate, Value: "1"}); metricErr != nil {
			fmt.Printf("Error publishing %s metric: %v\n", metrics.CacheCorruptionRate, metricErr)
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
		fmt.Printf("error opening %s: %v\n", path, err)
		return "", err
	}

	// #nosec
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		fmt.Printf("error creating gzip reader: %v\n", err)
		return "", err
	}

	tr := tar.NewReader(gzipReader)
	header, err := tr.Next()
	if err == io.EOF {
		fmt.Printf("No files in archive.\n")
		_ = gzipReader.Close()
		return "", nil
	}

	if err != nil {
		fmt.Printf("Error reading %s: %v\n", path, err)
		_ = gzipReader.Close()
		return "", err
	}

	return header.Name, gzipReader.Close()
}
