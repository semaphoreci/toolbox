package archive

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
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
			assertCompressAndUnpack(t, archiver, tempDir, tempFile)
		})

		t.Run(archiverType+" using relative paths", func(t *testing.T) {
			cwd, _ := os.Getwd()
			tempDir, _ := ioutil.TempDir(cwd, "*")
			tempFile, _ := ioutil.TempFile(tempDir, "*")
			tempDirBase := filepath.Base(tempDir)
			assertCompressAndUnpack(t, archiver, tempDirBase, tempFile)
		})

		t.Run(archiverType+" using single file", func(t *testing.T) {
			cwd, _ := os.Getwd()
			tempFile, _ := ioutil.TempFile(cwd, "*")
			_ = tempFile.Close()

			// compressing
			compressedFileName := tmpFileNameWithPrefix("abc0003")
			err := archiver.Compress(compressedFileName, tempFile.Name())
			assert.Nil(t, err)
			assert.Contains(t, compressedFileName, filepath.FromSlash(fmt.Sprintf("%s/abc0003", os.TempDir())))

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
	})
}

func Test__Decompress(t *testing.T) {
	runTestForAllArchiverTypes(t, true, func(archiverType string, archiver Archiver) {
		t.Run(archiverType+" sends metric on failure", func(t *testing.T) {
			tempFile, _ := ioutil.TempFile(os.TempDir(), "*")
			tempFile.WriteString("this is not a proper archive")
			_ = tempFile.Close()

			_, err := archiver.Decompress(tempFile.Name())
			assert.NotNil(t, err)

			metricsFile := path.Join(os.TempDir(), "toolbox_metrics")
			bytes, err := ioutil.ReadFile(metricsFile)
			assert.Nil(t, err)
			assert.Contains(t, string(bytes), fmt.Sprintf("%s 1", metrics.CacheCorruptionRate))

			os.Remove(tempFile.Name())
			os.Remove(metricsFile)
		})
	})
}

func tmpFileNameWithPrefix(prefix string) string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("%s-%d", prefix, time.Now().Nanosecond()))
}

func assertCompressAndUnpack(t *testing.T, archiver Archiver, tempDirectory string, tempFile *os.File) {
	_ = tempFile.Close()

	// compressing
	compressedFileName := tmpFileNameWithPrefix("abc0003")
	assert.NoError(t, archiver.Compress(compressedFileName, tempDirectory))
	assert.Contains(t, compressedFileName, path.Join(os.TempDir(), "abc0003"))

	_, err := os.Stat(compressedFileName)
	assert.Nil(t, err)

	// make sure file and directory are deleted
	// before trying to unpack.
	assert.NoError(t, os.Remove(tempFile.Name()))
	assert.NoError(t, os.Remove(tempDirectory))

	// unpacking
	unpackedAt, err := archiver.Decompress(compressedFileName)
	assert.Nil(t, err)
	assert.Equal(t, filepath.FromSlash(fmt.Sprintf("%s/", tempDirectory)), unpackedAt)

	files, _ := ioutil.ReadDir(unpackedAt)
	assert.Len(t, files, 1)
	file := files[0]
	assert.Equal(t, filepath.Base(tempFile.Name()), file.Name())

	assert.NoError(t, os.Remove(tempFile.Name()))
	assert.NoError(t, os.Remove(unpackedAt))
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
