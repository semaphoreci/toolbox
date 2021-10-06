package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__Delete(t *testing.T) {
	runTestForAllStorageTypes(t, func(storageType string, storage Storage) {
		t.Run(fmt.Sprintf("%s non-existing key", storageType), func(t *testing.T) {
			_ = storage.Clear()
			err := storage.Delete("this-key-does-not-exist")
			assert.Nil(t, err)
		})

		t.Run(fmt.Sprintf("%s existing key", storageType), func(t *testing.T) {
			_ = storage.Clear()

			file, _ := ioutil.TempFile("/tmp", "*")
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
	})
}
