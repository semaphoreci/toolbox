package storage

import (
	"io/ioutil"
	"os"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__S3__Restore(t *testing.T) {
	storage, err := NewS3Storage()
	assert.Nil(t, err)

	t.Run("object exists", func(t *testing.T) {
		_ = storage.Clear()

		file, _ := ioutil.TempFile("/tmp", "*")
		file.WriteString("S3 - restore - object exists")

		err = storage.Store("abc001", file.Name())
		assert.Nil(t, err)

		restoredFile, err := storage.Restore("abc001")
		assert.Nil(t, err)

		content, err := ioutil.ReadFile(restoredFile.Name())
		assert.Nil(t, err)
		assert.Equal(t, "S3 - restore - object exists", string(content))

		os.Remove(file.Name())
		os.Remove(restoredFile.Name())
	})

	t.Run("object does not exist", func(t *testing.T) {
		_ = storage.Clear()

		_, err := storage.Restore("abc002")
		assert.NotNil(t, err)
	})
}
