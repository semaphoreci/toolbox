package storage

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__HasKey(t *testing.T) {
	ctx := context.TODO()
	runTestForAllStorageTypes(t, SortByStoreTime, func(storageType string, storage Storage) {
		t.Run(fmt.Sprintf("%s non-existing key", storageType), func(t *testing.T) {
			_ = storage.Clear(ctx)
			exists, err := storage.HasKey(ctx, "this-key-does-not-exist")
			assert.Nil(t, err)
			assert.False(t, exists)
		})

		t.Run(fmt.Sprintf("%s existing key", storageType), func(t *testing.T) {
			_ = storage.Clear(ctx)

			file, _ := ioutil.TempFile(os.TempDir(), "*")
			_ = storage.Store(ctx, "abc001", file.Name())

			exists, err := storage.HasKey(ctx, "abc001")
			assert.Nil(t, err)
			assert.True(t, exists)

			os.Remove(file.Name())
		})
	})
}
