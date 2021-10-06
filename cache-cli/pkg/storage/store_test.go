package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

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

			if assert.Len(t, keys, 1) {
				key := keys[0]
				assert.Equal(t, key.Name, "abc001")
				assert.NotNil(t, key.StoredAt)
				assert.NotNil(t, key.Size)
			}

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

	runTestForSingleStorageType("sftp", t, func(storage Storage) {
		t.Run("sftp storage deletes old keys if no space left to store", func(t *testing.T) {
			_ = storage.Clear()

			file1, _ := ioutil.TempFile("/tmp", "*")
			file1.WriteString(strings.Repeat("x", 400))
			storage.Store("abc001", file1.Name())

			time.Sleep(time.Second)

			file2, _ := ioutil.TempFile("/tmp", "*")
			file2.WriteString(strings.Repeat("x", 400))
			storage.Store("abc002", file2.Name())

			time.Sleep(time.Second)

			file3, _ := ioutil.TempFile("/tmp", "*")
			file3.WriteString(strings.Repeat("x", 400))
			storage.Store("abc003", file3.Name())

			keys, _ := storage.List()
			assert.Len(t, keys, 2)

			firstKey := keys[0]
			assert.Equal(t, "abc003", firstKey.Name)
			secondKey := keys[1]
			assert.Equal(t, "abc002", secondKey.Name)

			os.Remove(file1.Name())
			os.Remove(file2.Name())
			os.Remove(file3.Name())
		})
	})
}
