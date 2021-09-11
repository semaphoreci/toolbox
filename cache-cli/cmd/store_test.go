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
	storage, err := storage.InitStorage()
	assert.Nil(t, err)

	t.Run("wrong number of arguments", func(t *testing.T) {
		capturer := utils.CreateOutputCapturer()
		RunStore(storeCmd, []string{"key", "value", "extra-bad-argument"})
		output := capturer.Done()

		assert.Contains(t, output, "Wrong number of arguments")
	})

	t.Run("using key and invalid path", func(*testing.T) {
		capturer := utils.CreateOutputCapturer()
		RunStore(storeCmd, []string{"abc001", "/tmp/this-path-does-not-exist"})
		output := capturer.Done()

		assert.Contains(t, output, "Path /tmp/this-path-does-not-exist does not exist")
	})

	t.Run("using key and valid path", func(*testing.T) {
		storage.Clear()
		tempDir, _ := ioutil.TempDir("/tmp", "*")
		ioutil.TempFile(tempDir, "*")

		capturer := utils.CreateOutputCapturer()
		RunStore(storeCmd, []string{"abc002", tempDir})
		output := capturer.Done()

		assert.Contains(t, output, fmt.Sprintf("Uploading '%s' with cache key 'abc002'", tempDir))
		assert.Contains(t, output, "Upload complete")
	})

	t.Run("using duplicate key", func(*testing.T) {
		storage.Clear()
		tempDir, _ := ioutil.TempDir("/tmp", "*")
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
}

func Test__AutomaticStore(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	cmdPath := filepath.Dir(file)
	rootPath := filepath.Dir(cmdPath)
	storage, _ := storage.InitStorage()

	t.Run("nothing found", func(t *testing.T) {
		os.Chdir(cmdPath)

		capturer := utils.CreateOutputCapturer()
		RunStore(storeCmd, []string{})
		output := capturer.Done()

		assert.Contains(t, output, "Nothing to store in cache.")
	})

	t.Run("does not store if path does not exist", func(t *testing.T) {
		os.Chdir(fmt.Sprintf("%s/test/autocache/gems", rootPath))

		capturer := utils.CreateOutputCapturer()
		RunStore(storeCmd, []string{})
		output := capturer.Done()

		assert.Contains(t, output, "Path vendor/bundle does not exist")
	})

	t.Run("detects and stores", func(t *testing.T) {
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

	t.Run("does not store if key already exist", func(t *testing.T) {
		storage.Clear()

		os.Chdir(fmt.Sprintf("%s/test/autocache/gems", rootPath))
		os.Setenv("SEMAPHORE_GIT_BRANCH", "master")
		os.MkdirAll("vendor/bundle", os.ModePerm)

		checksum, _ := files.GenerateChecksum("Gemfile.lock")

		tempFile, _ := ioutil.TempFile("/tmp", "*")
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
}
