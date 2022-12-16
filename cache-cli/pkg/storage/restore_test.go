package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__Restore(t *testing.T) {
	runTestForAllStorageTypes(t, func(storageType string, storage Storage) {
		t.Run(fmt.Sprintf("%s key exists", storageType), func(t *testing.T) {
			_ = storage.Clear()

			file, _ := ioutil.TempFile(os.TempDir(), "*")
			file.WriteString("restore - key exists")

			err := storage.Store("abc001", file.Name())
			assert.Nil(t, err)

			restoredFile, err := storage.Restore("abc001")
			assert.Nil(t, err)

			content, err := ioutil.ReadFile(restoredFile.Name())
			assert.Nil(t, err)
			assert.Equal(t, "restore - key exists", string(content))

			os.Remove(file.Name())
			os.Remove(restoredFile.Name())
		})

		t.Run(fmt.Sprintf("%s key does not exist", storageType), func(t *testing.T) {
			_ = storage.Clear()

			_, err := storage.Restore("abc002")
			assert.NotNil(t, err)
		})
	})
}
