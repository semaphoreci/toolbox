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

func Test__Restore(t *testing.T) {
	storage, err := storage.InitStorage()
	assert.Nil(t, err)

	t.Run("wrong number of arguments", func(t *testing.T) {
		capturer := utils.CreateOutputCapturer()
		RunRestore(restoreCmd, []string{"key", "extra-bad-argument"})
		output := capturer.Done()

		assert.Contains(t, output, "Wrong number of arguments")
	})

	t.Run("using single missing key", func(*testing.T) {
		storage.Clear()

		capturer := utils.CreateOutputCapturer()
		RunRestore(restoreCmd, []string{"this-key-does-not-exist"})
		output := capturer.Done()

		assert.Contains(t, output, "MISS: 'this-key-does-not-exist'.")
	})

	t.Run("using single exact key", func(*testing.T) {
		storage.Clear()

		tempDir, _ := ioutil.TempDir("/tmp", "*")
		tempFile, _ := ioutil.TempFile(tempDir, "*")

		compressAndStore(storage, "abc-001", tempDir)

		capturer := utils.CreateOutputCapturer()
		RunRestore(restoreCmd, []string{"abc-001"})
		output := capturer.Done()

		assert.Contains(t, output, "HIT: 'abc-001', using key 'abc-001'.")
		assert.Contains(t, output, fmt.Sprintf("Restored: %s/.", tempDir))

		os.Remove(tempFile.Name())
		os.Remove(tempDir)
	})

	t.Run("using single matching key", func(*testing.T) {
		storage.Clear()

		tempDir, _ := ioutil.TempDir("/tmp", "*")
		tempFile, _ := ioutil.TempFile(tempDir, "*")

		compressAndStore(storage, "abc-001", tempDir)

		capturer := utils.CreateOutputCapturer()
		RunRestore(restoreCmd, []string{"abc"})
		output := capturer.Done()

		assert.Contains(t, output, "HIT: 'abc', using key 'abc-001'.")
		assert.Contains(t, output, fmt.Sprintf("Restored: %s/.", tempDir))

		os.Remove(tempFile.Name())
		os.Remove(tempDir)
	})

	t.Run("only first matching key is used", func(*testing.T) {
		storage.Clear()

		tempDir, _ := ioutil.TempDir("/tmp", "*")
		tempFile, _ := ioutil.TempFile(tempDir, "*")

		compressAndStore(storage, "abc-001", tempDir)
		compressAndStore(storage, "abc-002", tempDir)

		capturer := utils.CreateOutputCapturer()
		RunRestore(restoreCmd, []string{"abc-001,abc-002"})
		output := capturer.Done()

		assert.Contains(t, output, "HIT: 'abc-001', using key 'abc-001'.")
		assert.Contains(t, output, fmt.Sprintf("Restored: %s/.", tempDir))
		assert.NotContains(t, output, "HIT: 'abc-002', using key 'abc-002'.")

		os.Remove(tempFile.Name())
		os.Remove(tempDir)
	})

	t.Run("using fallback key", func(*testing.T) {
		storage.Clear()

		tempDir, _ := ioutil.TempDir("/tmp", "*")
		tempFile, _ := ioutil.TempFile(tempDir, "*")

		compressAndStore(storage, "abc", tempDir)

		capturer := utils.CreateOutputCapturer()
		RunRestore(restoreCmd, []string{"abc-001,abc"})
		output := capturer.Done()

		assert.Contains(t, output, "MISS: 'abc-001'.")
		assert.Contains(t, output, "HIT: 'abc', using key 'abc'.")
		assert.Contains(t, output, fmt.Sprintf("Restored: %s/.", tempDir))

		os.Remove(tempFile.Name())
		os.Remove(tempDir)
	})
}

func Test__AutomaticRestore(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	cmdPath := filepath.Dir(file)
	rootPath := filepath.Dir(cmdPath)
	storage, _ := storage.InitStorage()

	t.Run("nothing found", func(t *testing.T) {
		storage.Clear()
		os.Chdir(cmdPath)

		capturer := utils.CreateOutputCapturer()
		RunRestore(restoreCmd, []string{})
		output := capturer.Done()

		assert.Contains(t, output, "Nothing to restore from cache")
	})

	t.Run("detects and restores", func(t *testing.T) {
		storage.Clear()

		os.Chdir(fmt.Sprintf("%s/test/autocache/gems", rootPath))
		os.Setenv("SEMAPHORE_GIT_BRANCH", "master")
		os.MkdirAll("vendor/bundle", os.ModePerm)

		// storing
		checksum, _ := files.GenerateChecksum("Gemfile.lock")
		key := fmt.Sprintf("gems-master-%s", checksum)
		compressedFile, _ := files.Compress(key, "vendor/bundle")
		storage.Store(key, compressedFile)

		// restoring
		capturer := utils.CreateOutputCapturer()
		RunRestore(restoreCmd, []string{})
		output := capturer.Done()

		assert.Contains(t, output, "Detected Gemfile.lock")
		assert.Contains(t, output, fmt.Sprintf("Downloading key '%s'", key))
		assert.Contains(t, output, "Restored: vendor/bundle")

		os.RemoveAll("vendor")
		os.Remove(compressedFile)
	})
}
