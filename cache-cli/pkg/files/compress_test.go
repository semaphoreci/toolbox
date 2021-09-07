package files

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__CompressAndUnpack(t *testing.T) {
	t.Run("file to compress is not present", func(t *testing.T) {
		compressedFileName, err := Compress("abc0001", "/tmp/this-file-does-not-exist")
		assert.NotNil(t, err)
		os.Remove(compressedFileName)
	})

	t.Run("file to unpack is not present", func(t *testing.T) {
		_, err := Unpack("/tmp/this-file-does-not-exist")
		assert.NotNil(t, err)
	})

	t.Run("using absolute paths", func(t *testing.T) {
		tempDir, _ := ioutil.TempDir("/tmp", "*")
		tempFile, _ := ioutil.TempFile(tempDir, "*")

		// compressing
		compressedFileName, err := Compress("abc0002", tempDir)
		assert.Nil(t, err)
		assert.True(t, strings.HasPrefix(compressedFileName, "/tmp/abc0002"))

		_, err = os.Stat(compressedFileName)
		assert.Nil(t, err)

		os.Remove(tempDir)

		// unpacking
		unpackedAt, err := Unpack(compressedFileName)
		assert.Nil(t, err)
		assert.Equal(t, fmt.Sprintf("%s/", tempDir), unpackedAt)

		files, _ := ioutil.ReadDir(unpackedAt)
		assert.Len(t, files, 1)
		file := files[0]
		assert.Equal(t, filepath.Base(tempFile.Name()), file.Name())

		os.Remove(tempFile.Name())
		os.Remove(unpackedAt)
		os.Remove(compressedFileName)
	})

	t.Run("using relative paths", func(t *testing.T) {
		// TODO
	})

	t.Run("using single file", func(t *testing.T) {
		// TODO
	})
}
