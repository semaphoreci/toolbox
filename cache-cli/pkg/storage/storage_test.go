package storage

import (
	"testing"

	assert "github.com/stretchr/testify/assert"
)

var storageTypes = map[string]func() (Storage, error){
	"s3": func() (Storage, error) {
		return NewS3Storage("http://s3:9000", "semaphore-cache", "cache-cli")
	},
	"sftp": func() (Storage, error) {
		return NewSFTPStorage("sftp-server", "tester", "/root/.ssh/semaphore_cache_key")
	},
}

func runTestForAllStorageTypes(t *testing.T, test func(string, Storage)) {
	for storageType, storageProvider := range storageTypes {
		storage, err := storageProvider()
		assert.Nil(t, err)
		test(storageType, storage)
	}
}
