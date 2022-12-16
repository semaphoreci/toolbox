package files

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	gzip "github.com/klauspost/pgzip"
	log "github.com/sirupsen/logrus"
)

// TODO: assert absolute paths work as well
func Compress(key, path string) (string, error) {
	epochNanos := time.Now().Nanosecond()
	tempFileName := filepath.Join(os.TempDir(), fmt.Sprintf("%s-%d", key, epochNanos))

	tempFile, err := os.OpenFile(tempFileName, os.O_CREATE|os.O_RDWR, os.FileMode(0644))
	if err != nil {
		log.Errorf("Error creating temporary file '%s' for archive: %v", tempFile, err)
		return "", err
	}

	// The order is 'tar > gzip > file'
	gzipWriter := gzip.NewWriter(tempFile)
	tarWriter := tar.NewWriter(gzipWriter)

	// We walk through every file in the
	// specified path, adding them to the tar archive.
	filepath.Walk(path, func(file string, fi os.FileInfo, e error) error {
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(file)
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if !fi.IsDir() {
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
		log.Errorf("Error archiving files for '%s': %v", path, err)
		return tempFileName, err
	}

	if err := tarWriter.Close(); err != nil {
		return tempFileName, err
	}

	if err := gzipWriter.Close(); err != nil {
		return tempFileName, err
	}

	return tempFileName, nil
}
