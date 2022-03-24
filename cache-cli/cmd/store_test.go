package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/files"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/utils"
	assert "github.com/stretchr/testify/assert"
)

func Test__Store(t *testing.T) {
	runTestForAllBackends(t, func(backend string, storage storage.Storage) {
		t.Run(fmt.Sprintf("%s wrong number of arguments", backend), func(t *testing.T) {
			capturer := utils.CreateOutputCapturer()
			RunStore(storeCmd, []string{"key", "value", "extra-bad-argument"})
			output := capturer.Done()

			assert.Contains(t, output, "Incorrect number of arguments!")
		})

		t.Run(fmt.Sprintf("%s using key and invalid path", backend), func(*testing.T) {
			capturer := utils.CreateOutputCapturer()
			RunStore(storeCmd, []string{"abc001", "/tmp/this-path-does-not-exist"})
			output := capturer.Done()

			assert.Contains(t, output, "'/tmp/this-path-does-not-exist' doesn't exist locally.")
		})

		t.Run(fmt.Sprintf("%s using key and valid path", backend), func(*testing.T) {
			storage.Clear()
			tempDir, _ := ioutil.TempDir(os.TempDir(), "*")
			ioutil.TempFile(tempDir, "*")

			capturer := utils.CreateOutputCapturer()
			RunStore(storeCmd, []string{"abc002", tempDir})
			output := capturer.Done()

			assert.Contains(t, output, fmt.Sprintf("Uploading '%s' with cache key 'abc002'", tempDir))
			assert.Contains(t, output, "Upload complete")
		})

		t.Run(fmt.Sprintf("%s normalizes key", backend), func(*testing.T) {
			storage.Clear()
			tempDir, _ := ioutil.TempDir(os.TempDir(), "*")
			ioutil.TempFile(tempDir, "*")

			capturer := utils.CreateOutputCapturer()
			RunStore(storeCmd, []string{"abc/00/12", tempDir})
			output := capturer.Done()

			assert.Contains(t, output, "Key 'abc/00/12' is normalized to 'abc-00-12'")
			assert.Contains(t, output, fmt.Sprintf("Uploading '%s' with cache key 'abc-00-12'", tempDir))
			assert.Contains(t, output, "Upload complete")
		})

		t.Run(fmt.Sprintf("%s using duplicate key", backend), func(*testing.T) {
			storage.Clear()
			tempDir, _ := ioutil.TempDir(os.TempDir(), "*")
			ioutil.TempFile(tempDir, "*")

			// Storing key for the first time
			capturer := utils.CreateOutputCapturer()
			RunStore(storeCmd, []string{"abc003", tempDir})
			output := capturer.Done()
			assert.Contains(t, output, fmt.Sprintf("Uploading '%s' with cache key 'abc003'", tempDir))
			assert.Contains(t, output, "Upload complete")

			// Storing key for the second time
			capturer = utils.CreateOutputCapturer()
			RunStore(storeCmd, []string{"abc003", tempDir})
			output = capturer.Done()
			assert.Contains(t, output, "Key 'abc003' already exists")
		})
	})
}

func Test__AutomaticStore(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	cmdPath := filepath.Dir(file)
	rootPath := filepath.Dir(cmdPath)

	runTestForAllBackends(t, func(backend string, storage storage.Storage) {
		t.Run(fmt.Sprintf("%s nothing found", backend), func(t *testing.T) {
			os.Chdir(cmdPath)

			capturer := utils.CreateOutputCapturer()
			RunStore(storeCmd, []string{})
			output := capturer.Done()

			assert.Contains(t, output, "Nothing to store in cache.")
		})

		t.Run(fmt.Sprintf("%s does not store if path does not exist", backend), func(t *testing.T) {
			os.Chdir(fmt.Sprintf("%s/test/autocache/gems", rootPath))

			capturer := utils.CreateOutputCapturer()
			RunStore(storeCmd, []string{})
			output := capturer.Done()

			assert.Contains(t, output, "'vendor/bundle' doesn't exist locally.")
		})

		t.Run(fmt.Sprintf("%s detects and stores", backend), func(t *testing.T) {
			storage.Clear()

			os.Chdir(fmt.Sprintf("%s/test/autocache/gems", rootPath))
			os.Setenv("SEMAPHORE_GIT_BRANCH", "master")
			os.MkdirAll("vendor/bundle", os.ModePerm)

			checksum, _ := files.GenerateChecksum("Gemfile.lock")

			key := fmt.Sprintf("gems-master-%s", checksum)

			capturer := utils.CreateOutputCapturer()
			RunStore(storeCmd, []string{})
			output := capturer.Done()

			assert.Contains(t, output, "Detected Gemfile.lock")
			assert.Contains(t, output, "Compressing vendor/bundle")
			assert.Contains(t, output, fmt.Sprintf("Uploading 'vendor/bundle' with cache key '%s'", key))
			assert.Contains(t, output, "Upload complete")

			os.RemoveAll("vendor")
		})

		t.Run(fmt.Sprintf("%s does not store if key already exist", backend), func(t *testing.T) {
			storage.Clear()

			os.Chdir(fmt.Sprintf("%s/test/autocache/gems", rootPath))
			os.Setenv("SEMAPHORE_GIT_BRANCH", "master")
			os.MkdirAll("vendor/bundle", os.ModePerm)

			checksum, _ := files.GenerateChecksum("Gemfile.lock")

			tempFile, _ := ioutil.TempFile(os.TempDir(), "*")
			key := fmt.Sprintf("gems-master-%s", checksum)
			err := storage.Store(key, tempFile.Name())
			assert.Nil(t, err)

			capturer := utils.CreateOutputCapturer()
			RunStore(storeCmd, []string{})
			output := capturer.Done()

			assert.Contains(t, output, fmt.Sprintf("Key '%s' already exists.", key))

			os.RemoveAll("vendor")
			os.Remove(tempFile.Name())
		})
	})
}
