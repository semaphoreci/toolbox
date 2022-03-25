package cmd

import (
	"os"
	"runtime"
	"testing"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	assert "github.com/stretchr/testify/assert"
)

type TestBackend struct {
	envVars      map[string]string
	runInWindows bool
}

var testBackends = map[string]TestBackend{
	"s3": {
		runInWindows: true,
		envVars: map[string]string{
			"SEMAPHORE_PROJECT_NAME":    "cache-cli",
			"SEMAPHORE_CACHE_BACKEND":   "s3",
			"SEMAPHORE_CACHE_S3_URL":    "http://s3:9000",
			"SEMAPHORE_CACHE_S3_BUCKET": "semaphore-cache",
		},
	},
	"sftp": {
		runInWindows: false,
		envVars: map[string]string{
			"SEMAPHORE_CACHE_BACKEND":          "sftp",
			"SEMAPHORE_CACHE_URL":              "sftp-server:22",
			"SEMAPHORE_CACHE_USERNAME":         "tester",
			"SEMAPHORE_CACHE_PRIVATE_KEY_PATH": "/root/.ssh/semaphore_cache_key",
		},
	},
}

func runTestForAllBackends(t *testing.T, test func(string, storage.Storage)) {
	for backendType, testBackend := range testBackends {
		if runtime.GOOS == "windows" && !testBackend.runInWindows {
			continue
		}

		for envVarName, envVarValue := range testBackend.envVars {
			os.Setenv(envVarName, envVarValue)
		}

		storage, err := storage.InitStorage()
		assert.Nil(t, err)
		test(backendType, storage)
	}
}
