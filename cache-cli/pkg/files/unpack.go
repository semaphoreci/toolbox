package files

import (
	"archive/tar"
	"io"
	"os"

	gzip "github.com/klauspost/pgzip"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/metrics"
	log "github.com/sirupsen/logrus"
)

func Unpack(metricsManager metrics.MetricsManager, reader io.Reader) error {
	uncompressedStream, err := gzip.NewReader(reader)
	if err != nil {
		log.Errorf("error creating gzip reader: %v", err)
		return err
	}

	defer uncompressedStream.Close()

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Errorf("Error reading tar stream: %v", err)
			if metricErr := metricsManager.Publish(metrics.Metric{Name: metrics.CacheCorruptionRate, Value: "1"}); metricErr != nil {
				log.Errorf("Error publishing %s metric: %v", metrics.CacheCorruptionRate, metricErr)
			}

			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(header.Name, 0755); err != nil {
				return err
			}

		case tar.TypeReg:
			outFile, err := os.Create(header.Name)
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}

			outFile.Close()
		}
	}

	return nil
}
