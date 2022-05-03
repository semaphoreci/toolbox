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
	runTestForAllBackends(t, func(backend string, storage storage.Storage) {
		t.Run(fmt.Sprintf("%s wrong number of arguments", backend), func(t *testing.T) {
			capturer := utils.CreateOutputCapturer()
			RunRestore(restoreCmd, []string{"key", "extra-bad-argument"})
			output := capturer.Done()

			assert.Contains(t, output, "Incorrect number of arguments!")
		})

		t.Run(fmt.Sprintf("%s using single missing key", backend), func(*testing.T) {
			storage.Clear()

			capturer := utils.CreateOutputCapturer()
			RunRestore(restoreCmd, []string{"this-key-does-not-exist"})
			output := capturer.Done()

			assert.Contains(t, output, "MISS: 'this-key-does-not-exist'.")
		})

		t.Run(fmt.Sprintf("%s using single exact key", backend), func(*testing.T) {
			storage.Clear()

			tempDir, _ := ioutil.TempDir(os.TempDir(), "*")
			tempFile, _ := ioutil.TempFile(tempDir, "*")
			_ = tempFile.Close()

			compressAndStore(storage, "abc-001", tempDir)

			capturer := utils.CreateOutputCapturer()
			RunRestore(restoreCmd, []string{"abc-001"})
			output := capturer.Done()

			restoredPath := filepath.FromSlash(fmt.Sprintf("%s/", tempDir))
			assert.Contains(t, output, "HIT: 'abc-001', using key 'abc-001'.")
			assert.Contains(t, output, fmt.Sprintf("Restored: %s.", restoredPath))

			os.Remove(tempFile.Name())
			os.Remove(tempDir)
		})

		t.Run(fmt.Sprintf("%s normalizes key", backend), func(*testing.T) {
			storage.Clear()

			tempDir, _ := ioutil.TempDir(os.TempDir(), "*")
			tempFile, _ := ioutil.TempFile(tempDir, "*")
			_ = tempFile.Close()

			compressAndStore(storage, "abc/00/22", tempDir)

			capturer := utils.CreateOutputCapturer()
			RunRestore(restoreCmd, []string{"abc/00/22"})
			output := capturer.Done()

			restoredPath := filepath.FromSlash(fmt.Sprintf("%s/", tempDir))
			assert.Contains(t, output, "Key 'abc/00/22' is normalized to 'abc-00-22'")
			assert.Contains(t, output, "HIT: 'abc-00-22', using key 'abc-00-22'.")
			assert.Contains(t, output, fmt.Sprintf("Restored: %s.", restoredPath))

			os.Remove(tempFile.Name())
			os.Remove(tempDir)
		})

		t.Run(fmt.Sprintf("%s using single matching key", backend), func(*testing.T) {
			storage.Clear()

			tempDir, _ := ioutil.TempDir(os.TempDir(), "*")
			tempFile, _ := ioutil.TempFile(tempDir, "*")
			_ = tempFile.Close()

			compressAndStore(storage, "abc-001", tempDir)

			capturer := utils.CreateOutputCapturer()
			RunRestore(restoreCmd, []string{"abc"})
			output := capturer.Done()

			restoredPath := filepath.FromSlash(fmt.Sprintf("%s/", tempDir))
			assert.Contains(t, output, "HIT: 'abc', using key 'abc-001'.")
			assert.Contains(t, output, fmt.Sprintf("Restored: %s.", restoredPath))

			os.Remove(tempFile.Name())
			os.Remove(tempDir)
		})

		t.Run(fmt.Sprintf("%s only first matching key is used", backend), func(*testing.T) {
			storage.Clear()

			tempDir, _ := ioutil.TempDir(os.TempDir(), "*")
			tempFile, _ := ioutil.TempFile(tempDir, "*")
			_ = tempFile.Close()

			compressAndStore(storage, "abc-001", tempDir)
			compressAndStore(storage, "abc-002", tempDir)

			capturer := utils.CreateOutputCapturer()
			RunRestore(restoreCmd, []string{"abc-001,abc-002"})
			output := capturer.Done()

			restoredPath := filepath.FromSlash(fmt.Sprintf("%s/", tempDir))
			assert.Contains(t, output, "HIT: 'abc-001', using key 'abc-001'.")
			assert.Contains(t, output, fmt.Sprintf("Restored: %s.", restoredPath))
			assert.NotContains(t, output, "HIT: 'abc-002', using key 'abc-002'.")

			os.Remove(tempFile.Name())
			os.Remove(tempDir)
		})

		t.Run(fmt.Sprintf("%s using fallback key", backend), func(*testing.T) {
			storage.Clear()

			tempDir, _ := ioutil.TempDir(os.TempDir(), "*")
			tempFile, _ := ioutil.TempFile(tempDir, "*")
			_ = tempFile.Close()

			compressAndStore(storage, "abc", tempDir)

			capturer := utils.CreateOutputCapturer()
			RunRestore(restoreCmd, []string{"abc-001,abc"})
			output := capturer.Done()

			restoredPath := filepath.FromSlash(fmt.Sprintf("%s/", tempDir))
			assert.Contains(t, output, "MISS: 'abc-001'.")
			assert.Contains(t, output, "HIT: 'abc', using key 'abc'.")
			assert.Contains(t, output, fmt.Sprintf("Restored: %s.", restoredPath))

			os.Remove(tempFile.Name())
			os.Remove(tempDir)
		})

		t.Run(fmt.Sprintf("%s using regex key", backend), func(*testing.T) {
			storage.Clear()

			tempDir, _ := ioutil.TempDir(os.TempDir(), "*")
			tempFile, _ := ioutil.TempFile(tempDir, "*")
			_ = tempFile.Close()

			compressAndStore(storage, "abc", tempDir)

			capturer := utils.CreateOutputCapturer()
			RunRestore(restoreCmd, []string{"^abc"})
			output := capturer.Done()

			restoredPath := filepath.FromSlash(fmt.Sprintf("%s/", tempDir))
			assert.Contains(t, output, "HIT: '^abc', using key 'abc'.")
			assert.Contains(t, output, fmt.Sprintf("Restored: %s.", restoredPath))

			os.Remove(tempFile.Name())
			os.Remove(tempDir)
		})
	})
}

func Test__AutomaticRestore(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	cmdPath := filepath.Dir(file)
	rootPath := filepath.Dir(cmdPath)

	runTestForAllBackends(t, func(backend string, storage storage.Storage) {
		t.Run(fmt.Sprintf("%s nothing found", backend), func(t *testing.T) {
			storage.Clear()
			os.Chdir(cmdPath)

			capturer := utils.CreateOutputCapturer()
			RunRestore(restoreCmd, []string{})
			output := capturer.Done()

			assert.Contains(t, output, "Nothing to restore from cache")
		})

		t.Run(fmt.Sprintf("%s detects and restores using SEMAPHORE_GIT_BRANCH", backend), func(t *testing.T) {
			storage.Clear()

			os.Chdir(fmt.Sprintf("%s/test/autocache/gems", rootPath))
			os.Setenv("SEMAPHORE_GIT_BRANCH", "master")
			os.Setenv("SEMAPHORE_GIT_REF_TYPE", "branch")
			os.Setenv("SEMAPHORE_GIT_PR_BRANCH", "")
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
			assert.Contains(t, output, fmt.Sprintf("Restored: %s", filepath.FromSlash("vendor/bundle")))

			os.RemoveAll("vendor")
			os.Remove(compressedFile)
		})

		t.Run(fmt.Sprintf("%s detects and restores using SEMAPHORE_GIT_PR_BRANCH", backend), func(t *testing.T) {
			storage.Clear()

			os.Chdir(fmt.Sprintf("%s/test/autocache/gems", rootPath))
			os.Setenv("SEMAPHORE_GIT_BRANCH", "master")
			os.Setenv("SEMAPHORE_GIT_REF_TYPE", "pull-request")
			os.Setenv("SEMAPHORE_GIT_PR_BRANCH", "some-development-branch")
			os.MkdirAll("vendor/bundle", os.ModePerm)

			// storing
			checksum, _ := files.GenerateChecksum("Gemfile.lock")
			key := fmt.Sprintf("gems-some-development-branch-%s", checksum)
			compressedFile, _ := files.Compress(key, "vendor/bundle")
			storage.Store(key, compressedFile)

			// restoring
			capturer := utils.CreateOutputCapturer()
			RunRestore(restoreCmd, []string{})
			output := capturer.Done()

			assert.Contains(t, output, "Detected Gemfile.lock")
			assert.Contains(t, output, fmt.Sprintf("Downloading key '%s'", key))
			assert.Contains(t, output, fmt.Sprintf("Restored: %s", filepath.FromSlash("vendor/bundle")))

			os.RemoveAll("vendor")
			os.Remove(compressedFile)
		})
	})
}
