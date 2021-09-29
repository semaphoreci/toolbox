package storage

import (
	"io/ioutil"
	"os"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__S3__Delete(t *testing.T) {
	storage, err := NewS3Storage()
	assert.Nil(t, err)

	t.Run("non-existing key", func(t *testing.T) {
		_ = storage.Clear()
		err := storage.Delete("this-key-does-not-exist")

		// if the object does not exist, S3 still responds with success: https://docs.aws.amazon.com/AmazonS3/latest/API/API_DeleteObject.html
		assert.Nil(t, err)
	})

	t.Run("existing key", func(t *testing.T) {
		_ = storage.Clear()

		file, _ := ioutil.TempFile("/tmp", "*")
		file.WriteString("s3 - delete - objects")
		_ = storage.Store("abc001", file.Name())

		keys, err := storage.List()
		assert.Nil(t, err)
		assert.Len(t, keys, 1)

		err = storage.Delete("abc001")
		assert.Nil(t, err)

		keys, err = storage.List()
		assert.Nil(t, err)
		assert.Len(t, keys, 0)

		os.Remove(file.Name())
	})
}
