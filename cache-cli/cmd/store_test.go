package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/files"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/logging"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	log "github.com/sirupsen/logrus"
	assert "github.com/stretchr/testify/assert"
)

func Test__Store(t *testing.T) {
	ctx := context.TODO()
	storeCmd := NewStoreCommand()
	log.SetFormatter(new(logging.CustomFormatter))
	log.SetLevel(log.InfoLevel)
	log.SetOutput(openLogfileForTests(t))

	runTestForAllBackends(ctx, t, func(backend string, storage storage.Storage) {
		t.Run(fmt.Sprintf("%s wrong number of arguments", backend), func(t *testing.T) {
			RunStore(storeCmd, []string{"key", "value", "extra-bad-argument"})
			output := readOutputFromFile(t)

			assert.Contains(t, output, "Incorrect number of arguments!")
		})

		t.Run(fmt.Sprintf("%s using key and invalid path", backend), func(*testing.T) {
			RunStore(storeCmd, []string{"abc001", "/tmp/this-path-does-not-exist"})
			output := readOutputFromFile(t)

			assert.Contains(t, output, fmt.Sprintf("'%s' doesn't exist locally.", filepath.FromSlash("/tmp/this-path-does-not-exist")))
		})

		t.Run(fmt.Sprintf("%s using key and valid path", backend), func(*testing.T) {
			storage.Clear(ctx)
			tempDir, _ := ioutil.TempDir(os.TempDir(), "*")
			ioutil.TempFile(tempDir, "*")

			RunStore(storeCmd, []string{"abc002", tempDir})
			output := readOutputFromFile(t)

			assert.Contains(t, output, fmt.Sprintf("Uploading '%s' with cache key 'abc002'", tempDir))
			assert.Contains(t, output, "Upload complete")
		})

		t.Run(fmt.Sprintf("%s normalizes key", backend), func(*testing.T) {
			storage.Clear(ctx)
			tempDir, _ := ioutil.TempDir(os.TempDir(), "*")
			ioutil.TempFile(tempDir, "*")

			RunStore(storeCmd, []string{"abc/00/12", tempDir})
			output := readOutputFromFile(t)

			assert.Contains(t, output, "Key 'abc/00/12' is normalized to 'abc-00-12'")
			assert.Contains(t, output, fmt.Sprintf("Uploading '%s' with cache key 'abc-00-12'", tempDir))
			assert.Contains(t, output, "Upload complete")
		})

		t.Run(fmt.Sprintf("%s using duplicate key", backend), func(*testing.T) {
			storage.Clear(ctx)
			tempDir, _ := ioutil.TempDir(os.TempDir(), "*")
			ioutil.TempFile(tempDir, "*")

			// Storing key for the first time
			RunStore(storeCmd, []string{"abc003", tempDir})
			output := readOutputFromFile(t)
			assert.Contains(t, output, fmt.Sprintf("Uploading '%s' with cache key 'abc003'", tempDir))
			assert.Contains(t, output, "Upload complete")

			// Storing key for the second time
			RunStore(storeCmd, []string{"abc003", tempDir})
			output = readOutputFromFile(t)
			assert.Contains(t, output, "Key 'abc003' already exists")
		})
	})
}

func Test__AutomaticStore(t *testing.T) {
	ctx := context.TODO()
	storeCmd := NewStoreCommand()
	_, file, _, _ := runtime.Caller(0)
	cmdPath := filepath.Dir(file)
	rootPath := filepath.Dir(cmdPath)

	log.SetFormatter(new(logging.CustomFormatter))
	log.SetLevel(log.InfoLevel)
	log.SetOutput(openLogfileForTests(t))

	runTestForAllBackends(ctx, t, func(backend string, storage storage.Storage) {
		t.Run(fmt.Sprintf("%s nothing found", backend), func(t *testing.T) {
			os.Chdir(cmdPath)

			RunStore(storeCmd, []string{})
			output := readOutputFromFile(t)

			assert.Contains(t, output, "Nothing to store in cache.")
		})

		t.Run(fmt.Sprintf("%s does not store if path does not exist", backend), func(t *testing.T) {
			os.Chdir(fmt.Sprintf("%s/test/autocache/gems", rootPath))

			RunStore(storeCmd, []string{})
			output := readOutputFromFile(t)

			assert.Contains(t, output, fmt.Sprintf("'%s' doesn't exist locally.", filepath.FromSlash("vendor/bundle")))
		})

		t.Run(fmt.Sprintf("%s detects and stores using SEMAPHORE_GIT_BRANCH", backend), func(t *testing.T) {
			storage.Clear(ctx)

			os.Chdir(fmt.Sprintf("%s/test/autocache/gems", rootPath))
			os.Setenv("SEMAPHORE_GIT_BRANCH", "master")
			os.Setenv("SEMAPHORE_GIT_PR_BRANCH", "")
			os.MkdirAll("vendor/bundle", os.ModePerm)

			checksum, _ := files.GenerateChecksum("Gemfile.lock")

			key := fmt.Sprintf("gems-master-%s", checksum)
			RunStore(storeCmd, []string{})
			output := readOutputFromFile(t)

			assert.Contains(t, output, "Detected Gemfile.lock")
			assert.Contains(t, output, fmt.Sprintf("Compressing %s", filepath.FromSlash("vendor/bundle")))
			assert.Contains(t, output, fmt.Sprintf("Uploading '%s' with cache key '%s'", filepath.FromSlash("vendor/bundle"), key))
			assert.Contains(t, output, "Upload complete")

			os.RemoveAll("vendor")
		})

		t.Run(fmt.Sprintf("%s detects and stores using SEMAPHORE_GIT_PR_BRANCH", backend), func(t *testing.T) {
			storage.Clear(ctx)

			os.Chdir(fmt.Sprintf("%s/test/autocache/gems", rootPath))
			os.Setenv("SEMAPHORE_GIT_BRANCH", "master")
			os.Setenv("SEMAPHORE_GIT_PR_BRANCH", "some-development-branch")
			os.MkdirAll("vendor/bundle", os.ModePerm)

			checksum, _ := files.GenerateChecksum("Gemfile.lock")

			key := fmt.Sprintf("gems-some-development-branch-%s", checksum)
			RunStore(storeCmd, []string{})
			output := readOutputFromFile(t)

			assert.Contains(t, output, "Detected Gemfile.lock")
			assert.Contains(t, output, fmt.Sprintf("Compressing %s", filepath.FromSlash("vendor/bundle")))
			assert.Contains(t, output, fmt.Sprintf("Uploading '%s' with cache key '%s'", filepath.FromSlash("vendor/bundle"), key))
			assert.Contains(t, output, "Upload complete")

			os.RemoveAll("vendor")
		})

		t.Run(fmt.Sprintf("%s does not store if key already exist", backend), func(t *testing.T) {
			storage.Clear(ctx)

			os.Chdir(fmt.Sprintf("%s/test/autocache/gems", rootPath))
			os.Setenv("SEMAPHORE_GIT_BRANCH", "master")
			os.Setenv("SEMAPHORE_GIT_PR_BRANCH", "")
			os.MkdirAll("vendor/bundle", os.ModePerm)

			checksum, _ := files.GenerateChecksum("Gemfile.lock")

			tempFile, _ := ioutil.TempFile(os.TempDir(), "*")
			key := fmt.Sprintf("gems-master-%s", checksum)
			err := storage.Store(ctx, key, tempFile.Name())
			assert.Nil(t, err)

			RunStore(storeCmd, []string{})
			output := readOutputFromFile(t)

			assert.Contains(t, output, fmt.Sprintf("Key '%s' already exists.", key))

			os.RemoveAll("vendor")
			os.Remove(tempFile.Name())
		})
	})
}
