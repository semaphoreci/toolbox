package storage

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__Delete(t *testing.T) {
	ctx := context.TODO()
	runTestForAllStorageTypes(t, SortByStoreTime, func(storageType string, storage Storage) {
		t.Run(fmt.Sprintf("%s non-existing key", storageType), func(t *testing.T) {
			_ = storage.Clear(ctx)
			err := storage.Delete(ctx, "this-key-does-not-exist")
			assert.Nil(t, err)
		})

		t.Run(fmt.Sprintf("%s existing key", storageType), func(t *testing.T) {
			_ = storage.Clear(ctx)

			file, _ := ioutil.TempFile(os.TempDir(), "*")
			_ = storage.Store(ctx, "abc001", file.Name())

			keys, err := storage.List(ctx)
			assert.Nil(t, err)
			assert.Len(t, keys, 1)

			err = storage.Delete(ctx, "abc001")
			assert.Nil(t, err)

			keys, err = storage.List(ctx)
			assert.Nil(t, err)
			assert.Len(t, keys, 0)

			os.Remove(file.Name())
		})
	})
}
