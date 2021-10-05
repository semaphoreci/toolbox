package storage

import (
	"io/ioutil"
	"os"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__S3__Usage(t *testing.T) {
	storage, err := NewS3Storage()
	assert.Nil(t, err)

	t.Run("no usage", func(t *testing.T) {
		_ = storage.Clear()
		usage, err := storage.Usage()
		assert.Nil(t, err)
		assert.Equal(t, int64(0), usage.Used)
		assert.Equal(t, int64(-1), usage.Free)
	})

	t.Run("some usage", func(t *testing.T) {
		_ = storage.Clear()

		file, _ := ioutil.TempFile("/tmp", "*")
		file.WriteString("s3 - usage - some usage")
		_ = storage.Store("abc001", file.Name())

		usage, err := storage.Usage()
		assert.Nil(t, err)
		assert.Greater(t, usage.Used, int64(0))
		assert.Equal(t, int64(-1), usage.Free)

		os.Remove(file.Name())
	})
}
