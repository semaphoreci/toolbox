package archive

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
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

type directoryStat struct {
	name string
	mode fs.FileMode
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
	hadError := false
	delayedDirectoryStats := []directoryStat{}

	// We read all blocks in the tar archive.
	// If an error is found when processing a particular tar block,
	// it is logged, and we move to the next tar block.
	// The only reason we return an error, and stop execution,
	// is if we have issues reading the tar archive.
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

		i++
		switch header.Typeflag {
		case tar.TypeDir:
			mode := header.FileInfo().Mode()

			// Directories can be filled with files, but not be writable.
			// See: https://github.com/golang/go/issues/27161.
			// So, if we create the directory with the permissions on the tar header,
			// we are not able to create the files inside of it afterwards.
			// In those cases, we create the directory with 0770 permissions,
			// and delay setting the proper permissions on the directory after all files are extracted.
			if header.FileInfo().Mode()&0200 == 0 {
				mode = 0770
				delayedDirectoryStats = append(delayedDirectoryStats, directoryStat{
					name: header.Name,
					mode: header.FileInfo().Mode(),
				})
			}

			if err := os.MkdirAll(header.Name, mode); err != nil {
				log.Errorf("Error creating directory '%s': %v", header.Name, err)
				hadError = true
				continue
			}

		case tar.TypeSymlink:
			// we have to remove the symlink first, if it exists.
			// Otherwise os.Symlink will complain.
			if _, err := os.Lstat(header.Name); err == nil {
				_ = os.Remove(header.Name)
			}

			if err := os.Symlink(header.Linkname, header.Name); err != nil {
				log.Errorf("Error creating symlink '%s'-'%s': %v", header.Name, header.Linkname, err)
				hadError = true
				continue
			}

		case tar.TypeReg:
			outFile, err := a.openFile(header, tarReader)
			if err != nil {
				log.Errorf("Error opening file '%s': %v", header.Name, err)
				hadError = true
				continue
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				log.Errorf("Error writing to file '%s': %v", header.Name, err)
				hadError = true
				_ = outFile.Close()
				continue
			}

			if err := outFile.Close(); err != nil {
				log.Errorf("Error closing file handle for '%s': %v", header.Name, err)
				hadError = true
				continue
			}
		}
	}

	for _, d := range delayedDirectoryStats {
		if err := os.Chmod(d.name, d.mode); err != nil {
			log.Errorf("error changing mode of directory '%s': %v", d.name, err)
			hadError = true
		}
	}

	if hadError {
		return restorationPath, fmt.Errorf("tar archive was not completely decompressed without errors")
	}

	return restorationPath, nil
}

func (a *NativeArchiver) openFile(header *tar.Header, tarReader *tar.Reader) (*os.File, error) {
	outFile, err := os.OpenFile(header.Name, os.O_RDWR|os.O_CREATE|os.O_EXCL, header.FileInfo().Mode())

	// File was opened successfully, just return it.
	if err == nil {
		return outFile, nil
	}

	// Since we are using O_EXCL, this error could mean that the file already exists.
	// If that is the case, we attempt to remove it before attempting to open it again.
	if errors.Is(err, os.ErrExist) {
		if err := os.Remove(header.Name); err != nil {
			return nil, fmt.Errorf("file '%s' already exists and can't be removed: %v", header.Name, err)
		}
	}

	// If a ErrNotExist is returned, it means the parent directory does not exist.
	// That means the file was included in the archive, but not its parent directory.
	// If that happens, we create the parent directory here, and try to open the file again.
	if errors.Is(err, os.ErrNotExist) {
		parentDir := filepath.Dir(header.Name)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return nil, fmt.Errorf("error creating directory '%s' for '%s': %v", parentDir, header.Name, err)
		}
	}

	// Try to open file again now that we handled some possible errors.
	outFile, err = os.OpenFile(header.Name, os.O_RDWR|os.O_CREATE|os.O_EXCL, header.FileInfo().Mode())
	if err != nil {
		return nil, fmt.Errorf("error opening file '%s': %v", header.Name, err)
	}

	return outFile, nil
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
