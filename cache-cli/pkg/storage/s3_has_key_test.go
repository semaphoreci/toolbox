package storage

import (
	"io/ioutil"
	"os"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__S3__HasKey(t *testing.T) {
	storage, err := NewS3Storage()
	assert.Nil(t, err)

	t.Run("non-existing key", func(t *testing.T) {
		_ = storage.Clear()
		exists, err := storage.HasKey("this-key-does-not-exist")
		assert.Nil(t, err)
		assert.False(t, exists)
	})

	t.Run("existing key", func(t *testing.T) {
		_ = storage.Clear()

		file, _ := ioutil.TempFile("/tmp", "*")
		file.WriteString("s3 - has_key - objects")
		_ = storage.Store("abc001", file.Name())

		exists, err := storage.HasKey("abc001")
		assert.Nil(t, err)
		assert.True(t, exists)

		os.Remove(file.Name())
	})
}
