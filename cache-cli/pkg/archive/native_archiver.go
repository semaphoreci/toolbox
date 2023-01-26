package archive

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
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
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("error finding '%s': %v", src, err)
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR, os.FileMode(0644))
	if err != nil {
		return err
	}

	// The order is 'tar > gzip > file'
	gzipWriter := a.newGzipWriter(dstFile)
	tarWriter := tar.NewWriter(gzipWriter)

	// We walk through every file in the specified path, adding them to the tar archive.
	err = filepath.Walk(src, func(fileName string, fileInfo os.FileInfo, e error) error {
		var link string
		if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
			if link, err = os.Readlink(fileName); err != nil {
				return fmt.Errorf("error reading symlink for '%s': %v", fileName, err)
			}
		}

		header, err := tar.FileInfoHeader(fileInfo, link)
		if err != nil {
			return fmt.Errorf("error creating tar header for '%s': %v", fileName, err)
		}

		if fileInfo.IsDir() {
			header.Name = fileName + string(os.PathSeparator)
		} else {
			header.Name = fileName
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("error writing tar header: %v", err)
		}

		// If the file is not a regular file, nothing else to do for it
		if !fileInfo.Mode().IsRegular() {
			return nil
		}

		// If it is a regular file, we need to copy its contents to the archive
		file, err := os.Open(fileName)
		if err != nil {
			return fmt.Errorf("error opening file '%s': %v", fileName, err)
		}

		if _, err := io.Copy(tarWriter, file); err != nil {
			return fmt.Errorf("error writing file '%s' to tar archive: %v", fileName, err)
		}

		_ = file.Close()

		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking tar archive: %v", err)
	}

	if err := tarWriter.Close(); err != nil {
		return fmt.Errorf("error closing tar writer: %v", err)
	}

	if err := gzipWriter.Close(); err != nil {
		return fmt.Errorf("error closing gzip writer: %v", err)
	}

	if err := dstFile.Close(); err != nil {
		return fmt.Errorf("error closing destination file '%s', %v", dst, err)
	}

	return nil
}

func (a *NativeArchiver) Decompress(src string) (string, error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return "", fmt.Errorf("error opening '%s': %v", src, err)
	}

	defer srcFile.Close()

	uncompressedStream, err := a.newGzipReader(srcFile)
	if err != nil {
		log.Errorf("error creating gzip reader: %v", err)
		a.publishCorruptionMetric()
		return "", err
	}

	defer uncompressedStream.Close()

	i := 0
	tarReader := tar.NewReader(uncompressedStream)
	restorationPath := ""

	// Sometimes, directories can have read-only permissions,
	// but be filled with files. See: https://github.com/golang/go/issues/27161.
	delayedDirectoryModes := []*tar.Header{}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			a.publishCorruptionMetric()
			return "", fmt.Errorf("error reading tar stream: %v", err)
		}

		// If it's the first file in archive, we keep track of its name.
		if i == 0 {
			restorationPath = header.Name
		}

		switch header.Typeflag {
		case tar.TypeDir:
			mode := header.FileInfo().Mode()

			// If the directory is not writable, we create it with 0770 permissions,
			// and when we are done writing the files inside of it, we apply the actual permissions.
			if header.FileInfo().Mode()&0200 == 0 {
				delayedDirectoryModes = append(delayedDirectoryModes, header)
				mode = 0770
			}

			if err := os.MkdirAll(header.Name, mode); err != nil {
				return "", fmt.Errorf("error creating directory '%s': %v", header.Name, err)
			}

		case tar.TypeSymlink:
			// we have to remove the symlink first, if it exists.
			// Otherwise os.Symlink will complain.
			if _, err := os.Lstat(header.Name); err == nil {
				if err := os.Remove(header.Name); err != nil {
					return "", fmt.Errorf("error removing symlink '%s': %v", header.Name, err)
				}
			}

			if err := os.Symlink(header.Linkname, header.Name); err != nil {
				return "", fmt.Errorf("error creating symlink '%s': %v", header.Name, err)
			}

		case tar.TypeReg:
			outFile, err := a.openFile(header, tarReader)
			if err != nil {
				return "", err
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				return "", fmt.Errorf("error copying to file '%s': %v", header.Name, err)
			}

			if err := outFile.Close(); err != nil {
				return "", fmt.Errorf("error closing file '%s': %v", header.Name, err)
			}
		}

		i++
	}

	for _, d := range delayedDirectoryModes {
		if err := os.Chmod(d.Name, d.FileInfo().Mode()); err != nil {
			return "", fmt.Errorf("error changing mode of directory '%s': %v", d.Name, err)
		}
	}

	return restorationPath, nil
}

func (a *NativeArchiver) openFile(header *tar.Header, tarReader *tar.Reader) (*os.File, error) {
	outFile, err := os.OpenFile(header.Name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, header.FileInfo().Mode())

	// File was opened successfully, just return it.
	if err == nil {
		return outFile, nil
	}

	// Check if this is a permission error.
	// This error could happen when trying to open an existing read-only file.
	// If that's the case, we try to remove the file, and open it again.
	// If the file can't be removed, then there's nothing we can do.
	if errors.Is(err, os.ErrPermission) {
		if err := os.Remove(header.Name); err != nil {
			return nil, fmt.Errorf("error removing file '%s': %v", header.Name, err)
		}

		outFile, err = os.OpenFile(header.Name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, header.FileInfo().Mode())
		if err != nil {
			return nil, fmt.Errorf("error opening file '%s': %v", header.Name, err)
		}

		return outFile, nil
	}

	// If we're dealing with a different error, just return it
	return nil, fmt.Errorf("error opening file '%s': %v", header.Name, err)
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
