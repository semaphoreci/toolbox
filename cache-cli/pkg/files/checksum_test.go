package files

import (
	"io/ioutil"
	"os"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__GeneratesChecksum(t *testing.T) {
	t.Run("file is present", func(t *testing.T) {
		tempFile, _ := ioutil.TempFile(os.TempDir(), "*")
		tempFile.WriteString("hello, hello\n")

		checksum, err := GenerateChecksum(tempFile.Name())
		assert.Nil(t, err)
		assert.Equal(t, "db243d472932e6e19fcb85468f962c46", checksum)

		os.Remove(tempFile.Name())
	})

	t.Run("file is not present", func(t *testing.T) {
		_, err := GenerateChecksum("/tmp/this-file-does-not-exist")
		assert.NotNil(t, err)
	})
}
