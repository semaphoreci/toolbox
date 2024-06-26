package storage

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__IsNotEmpty(t *testing.T) {
	ctx := context.TODO()
	runTestForAllStorageTypes(t, SortByStoreTime, func(storageType string, storage Storage) {
		t.Run(fmt.Sprintf("%s empty cache", storageType), func(t *testing.T) {
			_ = storage.Clear(ctx)
			isNotEmpty, err := storage.IsNotEmpty(ctx)
			assert.Nil(t, err)
			assert.False(t, isNotEmpty)
		})

		t.Run(fmt.Sprintf("%s non-empty cache", storageType), func(t *testing.T) {
			_ = storage.Clear(ctx)

			file, _ := ioutil.TempFile(os.TempDir(), "*")
			_ = storage.Store(ctx, "abc001", file.Name())

			isNotEmpty, err := storage.IsNotEmpty(ctx)
			assert.Nil(t, err)
			assert.True(t, isNotEmpty)

			os.Remove(file.Name())
		})
	})
}
