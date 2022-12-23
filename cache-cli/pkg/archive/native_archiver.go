package archive

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"

	pgzip "github.com/klauspost/pgzip"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/metrics"
	log "github.com/sirupsen/logrus"
)

type NativeArchiver struct {
	MetricsManager metrics.MetricsManager
	UseParallelism bool
}

func NewNativeArchiver(metricsManager metrics.MetricsManager, useParallelism bool) *NativeArchiver {
	return &NativeArchiver{
		MetricsManager: metricsManager,
		UseParallelism: useParallelism,
	}
}

func (a *NativeArchiver) Compress(dst, src string) error {
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR, os.FileMode(0644))
	if err != nil {
		log.Errorf("Error opening compression source '%s': %v", dst, err)
		return err
	}

	// The order is 'tar > gzip > file'
	gzipWriter := a.newGzipWriter(dstFile)
	tarWriter := tar.NewWriter(gzipWriter)

	// We walk through every file in the specified path, adding them to the tar archive.
	err = filepath.Walk(src, func(file string, fileInfo os.FileInfo, e error) error {
		header, err := tar.FileInfoHeader(fileInfo, file)
		if err != nil {
			return err
		}

		if fileInfo.IsDir() {
			header.Name = filepath.ToSlash(file + "/")
		} else {
			header.Name = filepath.ToSlash(file)
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if !fileInfo.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}

			if _, err := io.Copy(tarWriter, data); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	if err := tarWriter.Close(); err != nil {
		return err
	}

	return gzipWriter.Close()
}

func (a *NativeArchiver) Decompress(src string) (string, error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return "", err
	}

	uncompressedStream, err := a.newGzipReader(srcFile)
	if err != nil {
		log.Errorf("error creating gzip reader: %v", err)
		a.publishCorruptionMetric()
		return "", err
	}

	defer uncompressedStream.Close()
	tarReader := tar.NewReader(uncompressedStream)

	i := 0
	restorationPath := ""

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Errorf("Error reading tar stream: %v", err)
			a.publishCorruptionMetric()
			return "", err
		}

		// If it's the first file in archive, we keep track of its name.
		if i == 0 {
			restorationPath = header.Name
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(header.Name, 0755); err != nil {
				return "", err
			}

		case tar.TypeReg:
			outFile, err := os.Create(header.Name)
			if err != nil {
				return "", err
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				return "", err
			}

			outFile.Close()
		}

		i++
	}

	return restorationPath, nil
}

func (a *NativeArchiver) newGzipWriter(dstFile *os.File) io.WriteCloser {
	if a.UseParallelism {
		return pgzip.NewWriter(dstFile)
	}

	return gzip.NewWriter(dstFile)
}

func (a *NativeArchiver) newGzipReader(dstFile *os.File) (io.ReadCloser, error) {
	if a.UseParallelism {
		return pgzip.NewReader(dstFile)
	}

	return gzip.NewReader(dstFile)
}

func (a *NativeArchiver) publishCorruptionMetric() {
	err := a.MetricsManager.Publish(metrics.Metric{Name: metrics.CacheCorruptionRate, Value: "1"})
	if err != nil {
		log.Errorf("error publishing %s metric: %v", metrics.CacheCorruptionRate, err)
	}
}
