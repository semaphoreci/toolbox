package storage

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__Clear(t *testing.T) {
	ctx := context.TODO()
	runTestForAllStorageTypes(t, SortByStoreTime, func(storageType string, storage Storage) {
		setup := func(storage Storage) []string {
			_ = storage.Clear(ctx)

			file1, _ := ioutil.TempFile(os.TempDir(), "*")
			file1.WriteString("something, something")

			file2, _ := ioutil.TempFile(os.TempDir(), "*")
			file2.WriteString("else, else")

			_ = storage.Store(ctx, "abc001", file1.Name())
			_ = storage.Store(ctx, "abc002", file2.Name())

			return []string{file1.Name(), file2.Name()}
		}

		cleanup := func(files []string) {
			for _, file := range files {
				os.Remove(file)
			}
		}

		t.Run(fmt.Sprintf("%s no keys", storageType), func(t *testing.T) {
			err := storage.Clear(ctx)
			assert.Nil(t, err)

			keys, err := storage.List(ctx)
			assert.Nil(t, err)
			assert.Len(t, keys, 0)

			err = storage.Clear(ctx)
			assert.Nil(t, err)
		})

		t.Run(fmt.Sprintf("%s with keys", storageType), func(t *testing.T) {
			filesToCleanup := setup(storage)

			keys, err := storage.List(ctx)
			assert.Nil(t, err)
			assert.Len(t, keys, 2)

			err = storage.Clear(ctx)
			assert.Nil(t, err)

			keys, err = storage.List(ctx)
			assert.Nil(t, err)
			assert.Len(t, keys, 0)

			cleanup(filesToCleanup)
		})
	})
}
