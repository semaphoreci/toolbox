package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__Store(t *testing.T) {
	runTestForAllStorageTypes(t, func(storageType string, storage Storage) {
		t.Run(fmt.Sprintf("%s stored objects can be listed", storageType), func(t *testing.T) {
			_ = storage.Clear()

			file, _ := ioutil.TempFile("/tmp", "*")
			err := storage.Store("abc001", file.Name())
			assert.Nil(t, err)

			keys, err := storage.List()
			assert.Nil(t, err)
			assert.Len(t, keys, 1)

			key := keys[0]
			assert.Equal(t, key.Name, "abc001")
			assert.NotNil(t, key.StoredAt)
			assert.NotNil(t, key.Size)

			os.Remove(file.Name())
		})

		t.Run(fmt.Sprintf("%s stored objects can be restored", storageType), func(t *testing.T) {
			_ = storage.Clear()

			file, _ := ioutil.TempFile("/tmp", "*")
			file.WriteString("stored objects can be restored")

			err := storage.Store("abc002", file.Name())
			assert.Nil(t, err)

			restoredFile, err := storage.Restore("abc002")
			assert.Nil(t, err)

			content, err := ioutil.ReadFile(restoredFile.Name())
			assert.Nil(t, err)
			assert.Equal(t, "stored objects can be restored", string(content))

			os.Remove(file.Name())
			os.Remove(restoredFile.Name())
		})
	})
}
