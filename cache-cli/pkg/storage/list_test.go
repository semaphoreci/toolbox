package storage

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
)

func Test__List(t *testing.T) {
	ctx := context.TODO()
	runTestForAllStorageTypes(t, SortByStoreTime, func(storageType string, storage Storage) {
		t.Run(fmt.Sprintf("%s no keys", storageType), func(t *testing.T) {
			_ = storage.Clear(ctx)
			keys, err := storage.List(ctx)
			assert.Nil(t, err)
			assert.Len(t, keys, 0)
		})

		t.Run(fmt.Sprintf("%s keys are ordered by store time", storageType), func(t *testing.T) {
			err := storage.Clear(ctx)
			assert.Nil(t, err)

			file1, _ := ioutil.TempFile(os.TempDir(), "*")
			err = storage.Store(ctx, "abc001", file1.Name())
			assert.Nil(t, err)

			time.Sleep(time.Second)

			file2, _ := ioutil.TempFile(os.TempDir(), "*")
			err = storage.Store(ctx, "abc002", file2.Name())
			assert.Nil(t, err)

			keys, err := storage.List(ctx)
			assert.Nil(t, err)

			if assert.Len(t, keys, 2) {
				firstObject := keys[0]
				assert.Equal(t, firstObject.Name, "abc002")
				assert.NotNil(t, firstObject.StoredAt)
				assert.NotNil(t, firstObject.LastAccessedAt)
				assert.NotNil(t, firstObject.Size)

				secondObject := keys[1]
				assert.Equal(t, secondObject.Name, "abc001")
				assert.NotNil(t, secondObject.StoredAt)
				assert.NotNil(t, secondObject.LastAccessedAt)
				assert.NotNil(t, secondObject.Size)
			}

			os.Remove(file1.Name())
			os.Remove(file2.Name())
		})
	})

	runTestForAllStorageTypes(t, SortBySize, func(storageType string, storage Storage) {
		t.Run(fmt.Sprintf("%s keys are ordered by size", storageType), func(t *testing.T) {
			err := storage.Clear(ctx)
			assert.Nil(t, err)

			biggerFile := fmt.Sprintf("%s/bigger.tmp", os.TempDir())
			err = createBigTempFile(biggerFile, 100*1000*1000) // 100M
			assert.Nil(t, err)
			err = storage.Store(ctx, "bigger", biggerFile)
			assert.Nil(t, err)

			// Just to make sure things are really being sorted by size
			time.Sleep(2 * time.Second)

			smallerFile := fmt.Sprintf("%s/smaller.tmp", os.TempDir())
			err = createBigTempFile(smallerFile, 50*1000*1000) // 50M
			assert.Nil(t, err)
			err = storage.Store(ctx, "smaller", smallerFile)
			assert.Nil(t, err)

			keys, err := storage.List(ctx)
			assert.Nil(t, err)

			if assert.Len(t, keys, 2) {
				firstObject := keys[0]
				assert.Equal(t, firstObject.Name, "bigger")
				assert.NotNil(t, firstObject.StoredAt)
				assert.NotNil(t, firstObject.LastAccessedAt)
				assert.NotNil(t, firstObject.Size)

				secondObject := keys[1]
				assert.Equal(t, secondObject.Name, "smaller")
				assert.NotNil(t, secondObject.StoredAt)
				assert.NotNil(t, secondObject.LastAccessedAt)
				assert.NotNil(t, secondObject.Size)
			}

			os.Remove(biggerFile)
			os.Remove(smallerFile)
		})
	})

	if runtime.GOOS != "windows" {
		// s3 does not support access time sorting
		runTestForSingleStorageType("sftp", 1024, SortByAccessTime, t, func(storage Storage) {
			t.Run("sftp keys are ordered by access time", func(t *testing.T) {
				err := storage.Clear(ctx)
				assert.Nil(t, err)

				// store first key
				tmpFile, _ := ioutil.TempFile(os.TempDir(), "*")
				err = storage.Store(ctx, "abc001", tmpFile.Name())
				assert.Nil(t, err)

				// wait a little bit, and then store second key
				time.Sleep(2 * time.Second)
				err = storage.Store(ctx, "abc002", tmpFile.Name())
				assert.Nil(t, err)

				// wait a little bit, and then restore first key (access time will be updated)
				time.Sleep(2 * time.Second)
				_, err = storage.Restore(ctx, "abc001")
				assert.Nil(t, err)

				keys, err := storage.List(ctx)
				assert.Nil(t, err)

				if assert.Len(t, keys, 2) {
					firstObject := keys[0]
					assert.Equal(t, firstObject.Name, "abc001")
					assert.NotNil(t, firstObject.StoredAt)
					assert.NotNil(t, firstObject.LastAccessedAt)
					assert.NotNil(t, firstObject.Size)

					secondObject := keys[1]
					assert.Equal(t, secondObject.Name, "abc002")
					assert.NotNil(t, secondObject.StoredAt)
					assert.NotNil(t, secondObject.LastAccessedAt)
					assert.NotNil(t, secondObject.Size)
				}

				os.Remove(tmpFile.Name())
			})
		})
	}
}
