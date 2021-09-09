package storage

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
)

func Test__S3__List(t *testing.T) {
	storage, err := NewS3Storage()
	assert.Nil(t, err)

	t.Run("empty bucket", func(t *testing.T) {
		_ = storage.Clear()
		keys, err := storage.List()
		assert.Nil(t, err)
		assert.Len(t, keys, 0)
	})

	t.Run("objects are ordered by creation date", func(t *testing.T) {
		_ = storage.Clear()

		file1, _ := ioutil.TempFile("/tmp", "*")
		file1.WriteString("s3 - list - no objects")
		_ = storage.Store("abc001", file1.Name())

		time.Sleep(time.Second)

		file2, _ := ioutil.TempFile("/tmp", "*")
		file2.WriteString("s3 - list - no objects")
		_ = storage.Store("abc002", file2.Name())

		keys, err := storage.List()
		assert.Nil(t, err)
		assert.Len(t, keys, 2)

		firstObject := keys[0]
		assert.Equal(t, firstObject.Name, "abc002")
		assert.NotNil(t, firstObject.StoredAt)
		assert.NotNil(t, firstObject.Size)

		secondObject := keys[1]
		assert.Equal(t, secondObject.Name, "abc001")
		assert.NotNil(t, secondObject.StoredAt)
		assert.NotNil(t, secondObject.Size)

		os.Remove(file1.Name())
		os.Remove(file2.Name())
	})
}
