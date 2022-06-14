package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
)

func Test__List(t *testing.T) {
	runTestForAllStorageTypes(t, func(storageType string, storage Storage) {
		t.Run(fmt.Sprintf("%s no keys", storageType), func(t *testing.T) {
			_ = storage.Clear()
			keys, err := storage.List()
			assert.Nil(t, err)
			assert.Len(t, keys, 0)
		})

		t.Run(fmt.Sprintf("%s keys are ordered by creation date", storageType), func(t *testing.T) {
			err := storage.Clear()
			assert.Nil(t, err)

			file1, _ := ioutil.TempFile(os.TempDir(), "*")
			err = storage.Store("abc001", file1.Name())
			assert.Nil(t, err)

			time.Sleep(time.Second)

			file2, _ := ioutil.TempFile(os.TempDir(), "*")
			err = storage.Store("abc002", file2.Name())
			assert.Nil(t, err)

			keys, err := storage.List()
			assert.Nil(t, err)

			if assert.Len(t, keys, 2) {
				firstObject := keys[0]
				assert.Equal(t, firstObject.Name, "abc002")
				assert.NotNil(t, firstObject.StoredAt)
				assert.NotNil(t, firstObject.LastAccessedAt)
				assert.NotNil(t, firstObject.Size)

				secondObject := keys[1]
				assert.Equal(t, secondObject.Name, "abc001")
				assert.NotNil(t, secondObject.StoredAt)
				assert.NotNil(t, secondObject.LastAccessedAt)
				assert.NotNil(t, secondObject.Size)
			}

			os.Remove(file1.Name())
			os.Remove(file2.Name())
		})
	})
}
