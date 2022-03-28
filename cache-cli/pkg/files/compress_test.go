package files

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/metrics"
	assert "github.com/stretchr/testify/assert"
)

func Test__CompressAndUnpack(t *testing.T) {
	metricsManager, err := metrics.InitMetricsManager(metrics.LocalBackend)
	assert.Nil(t, err)

	t.Run("file to compress is not present", func(t *testing.T) {
		compressedFileName, err := Compress("abc0001", "/tmp/this-file-does-not-exist")
		assert.NotNil(t, err)
		os.Remove(compressedFileName)
	})

	t.Run("file to unpack is not present", func(t *testing.T) {
		_, err := Unpack(metricsManager, "/tmp/this-file-does-not-exist")
		assert.NotNil(t, err)
	})

	t.Run("using absolute paths", func(t *testing.T) {
		tempDir, _ := ioutil.TempDir(os.TempDir(), "*")
		tempFile, _ := ioutil.TempFile(tempDir, "*")
		assertCompressAndUnpack(t, metricsManager, tempDir, tempFile)
	})

	t.Run("using relative paths", func(t *testing.T) {
		cwd, _ := os.Getwd()
		tempDir, _ := ioutil.TempDir(cwd, "*")
		tempFile, _ := ioutil.TempFile(tempDir, "*")
		tempDirBase := filepath.Base(tempDir)
		assertCompressAndUnpack(t, metricsManager, tempDirBase, tempFile)
	})

	t.Run("using single file", func(t *testing.T) {
		cwd, _ := os.Getwd()
		tempFile, _ := ioutil.TempFile(cwd, "*")
		_ = tempFile.Close()

		// compressing
		compressedFileName, err := Compress("abc0003", tempFile.Name())
		assert.Nil(t, err)
		assert.Contains(t, compressedFileName, filepath.FromSlash(fmt.Sprintf("%s/abc0003", os.TempDir())))

		_, err = os.Stat(compressedFileName)
		assert.Nil(t, err)

		err = os.Remove(tempFile.Name())
		assert.Nil(t, err)

		// unpacking
		unpackedAt, err := Unpack(metricsManager, compressedFileName)
		assert.Nil(t, err)
		assert.Equal(t, tempFile.Name(), unpackedAt)

		_, err = os.Stat(unpackedAt)
		assert.Nil(t, err)

		err = os.Remove(tempFile.Name())
		assert.Nil(t, err)
		err = os.Remove(compressedFileName)
		assert.Nil(t, err)
	})
}

func assertCompressAndUnpack(t *testing.T, metricsManager metrics.MetricsManager, tempDirectory string, tempFile *os.File) {
	// compressing
	compressedFileName, err := Compress("abc0003", tempDirectory)
	assert.Nil(t, err)
	assert.Contains(t, compressedFileName, filepath.FromSlash(fmt.Sprintf("%s/abc0003", os.TempDir())))

	_, err = os.Stat(compressedFileName)
	assert.Nil(t, err)

	// make sure file and directory are deleted
	// before trying to unpack.
	_ = tempFile.Close()
	err = os.Remove(tempFile.Name())
	assert.Nil(t, err)
	err = os.Remove(tempDirectory)
	assert.Nil(t, err)

	// unpacking
	unpackedAt, err := Unpack(metricsManager, compressedFileName)
	assert.Nil(t, err)
	assert.Equal(t, fmt.Sprintf("%s/", tempDirectory), unpackedAt)

	files, _ := ioutil.ReadDir(unpackedAt)
	assert.Len(t, files, 1)
	file := files[0]
	assert.Equal(t, filepath.Base(tempFile.Name()), file.Name())

	err = os.Remove(tempFile.Name())
	assert.Nil(t, err)
	err = os.Remove(unpackedAt)
	assert.Nil(t, err)
	err = os.Remove(compressedFileName)
	assert.Nil(t, err)
}
