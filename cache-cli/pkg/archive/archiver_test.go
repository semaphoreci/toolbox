package archive

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/metrics"
	assert "github.com/stretchr/testify/assert"
)

func Test__Compress(t *testing.T) {
	runTestForAllArchiverTypes(t, false, func(archiverType string, archiver Archiver) {
		t.Run(archiverType+" file to compress is not present", func(t *testing.T) {
			err := archiver.Compress("???", "/tmp/this-file-does-not-exist")
			assert.NotNil(t, err)
		})

		t.Run(archiverType+" file to decompress is not present", func(t *testing.T) {
			_, err := archiver.Decompress("/tmp/this-file-does-not-exist")
			assert.NotNil(t, err)
		})

		t.Run(archiverType+" using absolute paths", func(t *testing.T) {
			tempDir, _ := ioutil.TempDir(os.TempDir(), "*")
			tempFile, _ := ioutil.TempFile(tempDir, "*")
			_ = tempFile.Close()

			assertCompressAndUnpack(t, archiver, tempDir, []fileAssertion{
				{
					name:    tempFile.Name(),
					mode:    fs.FileMode(0600),
					symlink: false,
				},
			})
		})

		t.Run(archiverType+" using relative paths", func(t *testing.T) {
			cwd, _ := os.Getwd()
			tempDir, _ := ioutil.TempDir(cwd, "*")
			tempFile, _ := ioutil.TempFile(tempDir, "*")
			tempDirBase := filepath.Base(tempDir)
			_ = tempFile.Close()

			assertCompressAndUnpack(t, archiver, tempDirBase, []fileAssertion{
				{
					name:    tempFile.Name(),
					mode:    fs.FileMode(0600),
					symlink: false,
				},
			})
		})

		t.Run(archiverType+" with symlink", func(t *testing.T) {
			cwd, _ := os.Getwd()
			tempDir, _ := ioutil.TempDir(cwd, "*")
			tempFile, _ := ioutil.TempFile(tempDir, "*")
			tempDirBase := filepath.Base(tempDir)
			_ = tempFile.Close()

			symlinkName := tempFile.Name() + "-link"
			assert.NoError(t, os.Symlink(tempFile.Name(), symlinkName))
			assertCompressAndUnpack(t, archiver, tempDirBase, []fileAssertion{
				{
					name:    tempFile.Name(),
					mode:    fs.FileMode(0600),
					symlink: false,
				},
				{
					name:    symlinkName,
					mode:    os.ModeSymlink,
					symlink: true,
				},
			})
		})

		t.Run(archiverType+" permissions bits are respected", func(t *testing.T) {
			cwd, _ := os.Getwd()
			tempDir, _ := ioutil.TempDir(cwd, "*")
			tempFile, _ := ioutil.TempFile(tempDir, "*")
			_ = tempFile.Close()
			assert.NoError(t, os.Chmod(tempFile.Name(), 0700))

			tempDirBase := filepath.Base(tempDir)
			assertCompressAndUnpack(t, archiver, tempDirBase, []fileAssertion{
				{
					name:    tempFile.Name(),
					mode:    fs.FileMode(0700),
					symlink: false,
				},
			})
		})

		t.Run(archiverType+" using read-only file", func(t *testing.T) {
			cwd, _ := os.Getwd()
			tempDir, _ := ioutil.TempDir(cwd, "*")
			tempFile, _ := ioutil.TempFile(tempDir, "*")

			// change file mode to read-only
			assert.NoError(t, os.Chmod(tempFile.Name(), 0444))

			tempDirBase := filepath.Base(tempDir)
			_ = tempFile.Close()

			// compressing
			compressedFileName := tmpFileNameWithPrefix("abc0003")
			assert.NoError(t, archiver.Compress(compressedFileName, tempDirBase))
			assert.Contains(t, compressedFileName, filepath.Join(os.TempDir(), "abc0003"))
			_, err := os.Stat(compressedFileName)
			assert.Nil(t, err)

			// make sure file and directory are deleted before trying to unpack.
			// Note: here we need to chmod before removing it.
			assert.NoError(t, os.Chmod(tempFile.Name(), 0755))
			assert.NoError(t, os.RemoveAll(tempDirBase))

			// unpacking
			unpackedAt, err := archiver.Decompress(compressedFileName)
			assert.Nil(t, err)
			assert.Equal(t, tempDirBase+string(os.PathSeparator), unpackedAt)

			files, _ := ioutil.ReadDir(unpackedAt)
			if assert.Len(t, files, 1) {
				file := files[0]
				assert.Equal(t, filepath.Base(tempFile.Name()), file.Name())
				if runtime.GOOS != "windows" {
					assert.Equal(t, fs.FileMode(0444), file.Mode())
				}

				assert.NoError(t, os.Chmod(unpackedAt, 0755))
				assert.NoError(t, os.RemoveAll(unpackedAt))
				assert.NoError(t, os.Remove(compressedFileName))
			}
		})

		t.Run(archiverType+" using read-only directory", func(t *testing.T) {
			cwd, _ := os.Getwd()
			tempDir, _ := ioutil.TempDir(cwd, "*")
			tempFile, _ := ioutil.TempFile(tempDir, "*")

			// change directory mode now that file is written
			assert.NoError(t, os.Chmod(tempDir, 0555))

			tempDirBase := filepath.Base(tempDir)
			_ = tempFile.Close()

			// compressing
			compressedFileName := tmpFileNameWithPrefix("abc0003")
			assert.NoError(t, archiver.Compress(compressedFileName, tempDirBase))
			assert.Contains(t, compressedFileName, filepath.Join(os.TempDir(), "abc0003"))
			_, err := os.Stat(compressedFileName)
			assert.Nil(t, err)

			// make sure file and directory are deleted before trying to unpack.
			// Note: here we need to chmod before removing it.
			assert.NoError(t, os.Chmod(tempDir, 0755))
			assert.NoError(t, os.RemoveAll(tempDirBase))

			// unpacking
			unpackedAt, err := archiver.Decompress(compressedFileName)
			if !assert.Nil(t, err) {
				return
			}

			assert.Equal(t, tempDirBase+string(os.PathSeparator), unpackedAt)

			// Assert directory is read-only
			dirInfo, err := os.Stat(unpackedAt)
			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, fs.ModeDir|fs.FileMode(0555), dirInfo.Mode())

			// Assert files inside read-only directory are correct
			files, _ := ioutil.ReadDir(unpackedAt)
			if assert.Len(t, files, 1) {
				file := files[0]
				assert.Equal(t, filepath.Base(tempFile.Name()), file.Name())
				assert.NoError(t, os.Chmod(unpackedAt, 0755))
				assert.NoError(t, os.RemoveAll(unpackedAt))
				assert.NoError(t, os.Remove(compressedFileName))
			}
		})

		t.Run(archiverType+" using single file", func(t *testing.T) {
			cwd, _ := os.Getwd()
			tempFile, _ := ioutil.TempFile(cwd, "*")
			_ = tempFile.Close()

			// compressing
			compressedFileName := tmpFileNameWithPrefix("abc0003")
			err := archiver.Compress(compressedFileName, tempFile.Name())
			assert.Nil(t, err)

			_, err = os.Stat(compressedFileName)
			assert.Nil(t, err)
			assert.NoError(t, os.Remove(tempFile.Name()))

			// unpacking
			unpackedAt, err := archiver.Decompress(compressedFileName)
			assert.Nil(t, err)
			assert.Equal(t, tempFile.Name(), unpackedAt)

			_, err = os.Stat(unpackedAt)
			assert.Nil(t, err)

			assert.NoError(t, os.Remove(tempFile.Name()))
			assert.NoError(t, os.Remove(compressedFileName))
		})

		t.Run(archiverType+" using single file in directory -> creates directory", func(t *testing.T) {
			cwd, _ := os.Getwd()
			tempDir, _ := ioutil.TempDir(cwd, "*")
			tempFile, _ := ioutil.TempFile(tempDir, "*")
			_ = tempFile.Close()

			// compressing
			compressedFileName := tmpFileNameWithPrefix("abc0007")
			err := archiver.Compress(compressedFileName, tempFile.Name())
			assert.Nil(t, err)

			// compressed file is created
			_, err = os.Stat(compressedFileName)
			assert.Nil(t, err)
			assert.NoError(t, os.RemoveAll(tempDir))

			// unpacking
			unpackedAt, err := archiver.Decompress(compressedFileName)
			assert.Nil(t, err)
			assert.Equal(t, tempFile.Name(), unpackedAt)

			_, err = os.Stat(unpackedAt)
			assert.Nil(t, err)

			assert.NoError(t, os.Remove(tempFile.Name()))
			assert.NoError(t, os.Remove(compressedFileName))
		})

		t.Run(archiverType+" respects timestamps", func(t *testing.T) {
			cwd, _ := os.Getwd()
			tempFile, _ := ioutil.TempFile(cwd, "*")
			_ = tempFile.Close()

			// change mtime to 1 hour ago
			originalTimestamp := time.Now().Add(time.Hour)
			_ = os.Chtimes(tempFile.Name(), originalTimestamp, originalTimestamp)

			// compressing
			compressedFileName := tmpFileNameWithPrefix("abc0007")
			err := archiver.Compress(compressedFileName, tempFile.Name())
			assert.Nil(t, err)

			// compressed file is created
			_, err = os.Stat(compressedFileName)
			assert.Nil(t, err)
			assert.NoError(t, os.Remove(tempFile.Name()))

			// unpacking
			unpackedAt, err := archiver.Decompress(compressedFileName)
			assert.Nil(t, err)
			assert.Equal(t, tempFile.Name(), unpackedAt)

			info, err := os.Stat(unpackedAt)
			assert.Nil(t, err)
			assert.Equal(t, info.ModTime().Unix(), originalTimestamp.Unix())

			assert.NoError(t, os.Remove(tempFile.Name()))
			assert.NoError(t, os.Remove(compressedFileName))
		})
	})
}

func Test__Decompress(t *testing.T) {
	os.Setenv("SEMAPHORE_TOOLBOX_METRICS_ENABLED", "true")
	runTestForAllArchiverTypes(t, true, func(archiverType string, archiver Archiver) {
		t.Run(archiverType+" sends metric on failure", func(t *testing.T) {
			tempFile, _ := ioutil.TempFile(os.TempDir(), "*")
			tempFile.WriteString("this is not a proper archive")
			_ = tempFile.Close()

			_, err := archiver.Decompress(tempFile.Name())
			assert.NotNil(t, err)

			metricsFile := filepath.Join(os.TempDir(), "toolbox_metrics")
			bytes, err := ioutil.ReadFile(metricsFile)
			assert.Nil(t, err)
			assert.Contains(t, string(bytes), "usercache")
			assert.Contains(t, string(bytes), "command=restore,corrupt=1")

			os.Remove(tempFile.Name())
			os.Remove(metricsFile)
		})
	})
}

func tmpFileNameWithPrefix(prefix string) string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("%s-%d", prefix, time.Now().Nanosecond()))
}

type fileAssertion struct {
	name    string
	mode    fs.FileMode
	symlink bool
}

func assertCompressAndUnpack(t *testing.T, archiver Archiver, tempDirectory string, assertions []fileAssertion) {
	// compressing
	compressedFileName := tmpFileNameWithPrefix("abc0003")
	assert.NoError(t, archiver.Compress(compressedFileName, tempDirectory))
	assert.Contains(t, compressedFileName, filepath.Join(os.TempDir(), "abc0003"))
	_, err := os.Stat(compressedFileName)
	assert.Nil(t, err)

	// make sure file and directory are deleted before trying to unpack.
	assert.NoError(t, os.RemoveAll(tempDirectory))

	// unpacking
	unpackedAt, err := archiver.Decompress(compressedFileName)
	assert.Nil(t, err)
	assert.Equal(t, tempDirectory+string(os.PathSeparator), unpackedAt)

	files, _ := ioutil.ReadDir(unpackedAt)
	if assert.Len(t, files, len(assertions)) {
		for i, a := range assertions {
			f := files[i]
			assert.Equal(t, filepath.Base(a.name), f.Name())
			assert.Equal(t, a.symlink, f.Mode()&os.ModeSymlink == os.ModeSymlink)
			if !a.symlink && runtime.GOOS != "windows" {
				assert.Equal(t, a.mode, f.Mode())
			}
		}
	}

	assert.NoError(t, os.RemoveAll(unpackedAt))
	assert.NoError(t, os.Remove(compressedFileName))
}

type archiverInitFn func(metricsManager metrics.MetricsManager) Archiver

var testArchiverTypes = map[string]archiverInitFn{
	"shell-out": func(metricsManager metrics.MetricsManager) Archiver {
		return NewShellOutArchiver(metricsManager)
	},
	"native": func(metricsManager metrics.MetricsManager) Archiver {
		return NewNativeArchiver(metricsManager, false)
	},
	"native-parallel": func(metricsManager metrics.MetricsManager) Archiver {
		return NewNativeArchiver(metricsManager, true)
	},
}

func runTestForAllArchiverTypes(t *testing.T, realMetrics bool, test func(string, Archiver)) {
	for archiverType, initFn := range testArchiverTypes {
		var metricsManager metrics.MetricsManager
		if realMetrics {
			metricsManager, _ = metrics.NewLocalMetricsBackend()
		} else {
			metricsManager = metrics.NewNoOpMetricsManager()
		}

		archiver := initFn(metricsManager)
		test(archiverType, archiver)
	}
}
