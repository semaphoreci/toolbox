package cmd

import (
	"os"
	"testing"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	assert "github.com/stretchr/testify/assert"
)

var testBackend = map[string]map[string]string{
	"s3": {
		"SEMAPHORE_PROJECT_NAME":    "cache-cli",
		"SEMAPHORE_CACHE_BACKEND":   "s3",
		"SEMAPHORE_CACHE_S3_URL":    "http://s3:9000",
		"SEMAPHORE_CACHE_S3_BUCKET": "semaphore-cache",
	},
	"sftp": {
		"SEMAPHORE_CACHE_BACKEND":          "sftp",
		"SEMAPHORE_CACHE_URL":              "sftp-server:22",
		"SEMAPHORE_CACHE_USERNAME":         "tester",
		"SEMAPHORE_CACHE_PRIVATE_KEY_PATH": "/root/.ssh/semaphore_cache_key",
	},
}

func runTestForAllBackends(t *testing.T, test func(string, storage.Storage)) {
	for backend, envVars := range testBackend {
		for envVarName, envVarValue := range envVars {
			os.Setenv(envVarName, envVarValue)
		}

		storage, err := storage.InitStorage()
		assert.Nil(t, err)

		test(backend, storage)
	}
}
