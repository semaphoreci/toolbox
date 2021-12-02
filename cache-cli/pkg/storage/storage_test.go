package storage

import (
	"math"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

var storageTypes = map[string]func(int64) (Storage, error){
	"s3": func(storageSize int64) (Storage, error) {
		return NewS3Storage(S3StorageOptions{
			URL:     "http://s3:9000",
			Bucket:  "semaphore-cache",
			Project: "cache-cli",
			Config:  StorageConfig{MaxSpace: math.MaxInt64},
		})
	},
	"sftp": func(storageSize int64) (Storage, error) {
		return NewSFTPStorage(SFTPStorageOptions{
			URL:            "sftp-server:22",
			Username:       "tester",
			PrivateKeyPath: "/root/.ssh/semaphore_cache_key",
			Config:         StorageConfig{MaxSpace: storageSize},
		})
	},
}

func runTestForAllStorageTypes(t *testing.T, test func(string, Storage)) {
	for storageType, storageProvider := range storageTypes {
		storage, err := storageProvider(9 * 1024 * 1024 * 1024)
		if assert.Nil(t, err) {
			test(storageType, storage)
		}
	}
}

func runTestForSingleStorageType(storageType string, storageSize int64, t *testing.T, test func(Storage)) {
	storageProvider := storageTypes[storageType]
	storage, err := storageProvider(storageSize)
	if assert.Nil(t, err) {
		test(storage)
	}
}
