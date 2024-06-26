package storage

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__Restore(t *testing.T) {
	ctx := context.TODO()
	runTestForAllStorageTypes(t, SortByStoreTime, func(storageType string, storage Storage) {
		t.Run(fmt.Sprintf("%s key exists", storageType), func(t *testing.T) {
			_ = storage.Clear(ctx)

			file, _ := ioutil.TempFile(os.TempDir(), "*")
			file.WriteString("restore - key exists")

			err := storage.Store(ctx, "abc001", file.Name())
			assert.Nil(t, err)

			restoredFile, err := storage.Restore(ctx, "abc001")
			assert.Nil(t, err)

			content, err := ioutil.ReadFile(restoredFile.Name())
			assert.Nil(t, err)
			assert.Equal(t, "restore - key exists", string(content))

			os.Remove(file.Name())
			os.Remove(restoredFile.Name())
		})

		t.Run(fmt.Sprintf("%s key does not exist", storageType), func(t *testing.T) {
			_ = storage.Clear(ctx)

			_, err := storage.Restore(ctx, "abc002")
			assert.NotNil(t, err)
		})
	})
}
