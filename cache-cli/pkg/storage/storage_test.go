package storage

import (
	"math"
	"os"
	"runtime"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

type TestStorageType struct {
	runInWindows bool
	initializer  func(storageSize int64, sortBy string) (Storage, error)
}

var testStorageTypes = map[string]TestStorageType{
	"s3": {
		runInWindows: true,
		initializer: func(storageSize int64, sortBy string) (Storage, error) {
			return NewS3Storage(S3StorageOptions{
				URL:     os.Getenv("SEMAPHORE_CACHE_S3_URL"),
				Bucket:  "semaphore-cache",
				Project: "cache-cli",
				Config:  StorageConfig{MaxSpace: math.MaxInt64, SortKeysBy: sortBy},
			})
		},
	},
	"sftp": {
		runInWindows: false,
		initializer: func(storageSize int64, sortBy string) (Storage, error) {
			return NewSFTPStorage(SFTPStorageOptions{
				URL:            "sftp-server:22",
				Username:       "tester",
				PrivateKeyPath: "/root/.ssh/semaphore_cache_key",
				Config:         StorageConfig{MaxSpace: storageSize, SortKeysBy: sortBy},
			})
		},
	},
	"gcs": {
		runInWindows: true,
		initializer: func(storageSize int64, sortBy string) (Storage, error) {
			return NewGCSStorage(GCSStorageOptions{
				Bucket:  "semaphore-cache",
				Config:  StorageConfig{MaxSpace: math.MaxInt64, SortKeysBy: sortBy},
			})
		},
	},
}

func runTestForAllStorageTypes(t *testing.T, sortBy string, test func(string, Storage)) {
	for storageType, testStorage := range testStorageTypes {
		if runtime.GOOS == "windows" && !testStorage.runInWindows {
			continue
		}

		storage, err := testStorage.initializer(9*1024*1024*1024, sortBy)
		if assert.Nil(t, err) {
			test(storageType, storage)
		}
	}
}

func runTestForSingleStorageType(storageType string, storageSize int64, sortBy string, t *testing.T, test func(Storage)) {
	storageProvider := testStorageTypes[storageType]
	storage, err := storageProvider.initializer(storageSize, sortBy)
	if assert.Nil(t, err) {
		test(storage)
	}
}
