package cmd

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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
			"SEMAPHORE_PROJECT_ID":      "cache-cli",
			"SEMAPHORE_CACHE_BACKEND":   "s3",
			"SEMAPHORE_CACHE_S3_URL":    os.Getenv("SEMAPHORE_CACHE_S3_URL"),
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

func runTestForSingleBackend(t *testing.T, testBackend string, test func(storage.Storage)) {
	backend := testBackends[testBackend]
	if runtime.GOOS == "windows" && !backend.runInWindows {
		return
	}

	for envVarName, envVarValue := range backend.envVars {
		os.Setenv(envVarName, envVarValue)
	}

	storage, err := storage.InitStorage()
	assert.Nil(t, err)
	test(storage)
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

func readOutputFromFile(t *testing.T) string {
	path := filepath.Join(os.TempDir(), "cache_log")

	defer os.Truncate(path, 0)

	output, err := ioutil.ReadFile(path)
	assert.NoError(t, err)

	return string(output)
}

func openLogfileForTests(t *testing.T) io.Writer {
	filePath := filepath.Join(os.TempDir(), "cache_log")
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	assert.NoError(t, err)
	return io.MultiWriter(f, os.Stdout)
}
