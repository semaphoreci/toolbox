package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__HasKey(t *testing.T) {
	runTestForAllStorageTypes(t, func(storageType string, storage Storage) {
		t.Run(fmt.Sprintf("%s non-existing key", storageType), func(t *testing.T) {
			_ = storage.Clear()
			exists, err := storage.HasKey("this-key-does-not-exist")
			assert.Nil(t, err)
			assert.False(t, exists)
		})

		t.Run(fmt.Sprintf("%s existing key", storageType), func(t *testing.T) {
			_ = storage.Clear()

			file, _ := ioutil.TempFile(os.TempDir(), "*")
			_ = storage.Store("abc001", file.Name())

			exists, err := storage.HasKey("abc001")
			assert.Nil(t, err)
			assert.True(t, exists)

			os.Remove(file.Name())
		})
	})
}
