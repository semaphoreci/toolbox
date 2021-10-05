package storage

import (
	"io/ioutil"
	"os"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__S3__Clear(t *testing.T) {
	storage, err := NewS3Storage()
	assert.Nil(t, err)

	setup := func(storage S3Storage) []string {
		_ = storage.Clear()

		file1, _ := ioutil.TempFile("/tmp", "*")
		file1.WriteString("something, something")

		file2, _ := ioutil.TempFile("/tmp", "*")
		file2.WriteString("else, else")

		_ = storage.Store("abc001", file1.Name())
		_ = storage.Store("abc002", file2.Name())

		return []string{file1.Name(), file2.Name()}
	}

	cleanup := func(files []string) {
		for _, file := range files {
			os.Remove(file)
		}
	}

	t.Run("bucket is empty", func(t *testing.T) {
		_ = storage.Clear()

		keys, err := storage.List()
		assert.Nil(t, err)
		assert.Len(t, keys, 0)

		err = storage.Clear()
		assert.Nil(t, err)
	})

	t.Run("bucket has objects", func(t *testing.T) {
		filesToCleanup := setup(*storage)

		keys, err := storage.List()
		assert.Nil(t, err)
		assert.Len(t, keys, 2)

		err = storage.Clear()
		assert.Nil(t, err)

		keys, err = storage.List()
		assert.Nil(t, err)
		assert.Len(t, keys, 0)

		cleanup(filesToCleanup)
	})
}
