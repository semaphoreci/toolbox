package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/archive"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/files"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/logging"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/metrics"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	log "github.com/sirupsen/logrus"
	assert "github.com/stretchr/testify/assert"
)

func Test__Restore(t *testing.T) {
	log.SetFormatter(new(logging.CustomFormatter))
	log.SetLevel(log.InfoLevel)
	log.SetOutput(openLogfileForTests(t))

	runTestForAllBackends(t, func(backend string, storage storage.Storage) {
		t.Run(fmt.Sprintf("%s wrong number of arguments", backend), func(t *testing.T) {
			RunRestore(restoreCmd, []string{"key", "extra-bad-argument"})
			output := readOutputFromFile(t)

			assert.Contains(t, output, "Incorrect number of arguments!")
		})

		t.Run(fmt.Sprintf("%s using single missing key", backend), func(*testing.T) {
			storage.Clear()

			RunRestore(restoreCmd, []string{"this-key-does-not-exist"})
			output := readOutputFromFile(t)

			assert.Contains(t, output, "MISS: 'this-key-does-not-exist'.")
		})

		t.Run(fmt.Sprintf("%s using single exact key", backend), func(*testing.T) {
			storage.Clear()

			tempDir, _ := ioutil.TempDir(os.TempDir(), "*")
			tempFile, _ := ioutil.TempFile(tempDir, "*")
			_ = tempFile.Close()

			archiver := archive.NewShellOutArchiver(metrics.NewNoOpMetricsManager())
			compressAndStore(storage, archiver, "abc-001", tempDir)
			RunRestore(restoreCmd, []string{"abc-001"})
			output := readOutputFromFile(t)

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

			archiver := archive.NewShellOutArchiver(metrics.NewNoOpMetricsManager())
			compressAndStore(storage, archiver, "abc/00/22", tempDir)
			RunRestore(restoreCmd, []string{"abc/00/22"})
			output := readOutputFromFile(t)

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

			archiver := archive.NewShellOutArchiver(metrics.NewNoOpMetricsManager())
			compressAndStore(storage, archiver, "abc-001", tempDir)
			RunRestore(restoreCmd, []string{"abc"})
			output := readOutputFromFile(t)

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

			archiver := archive.NewShellOutArchiver(metrics.NewNoOpMetricsManager())
			compressAndStore(storage, archiver, "abc-001", tempDir)
			compressAndStore(storage, archiver, "abc-002", tempDir)
			RunRestore(restoreCmd, []string{"abc-001,abc-002"})
			output := readOutputFromFile(t)

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

			archiver := archive.NewShellOutArchiver(metrics.NewNoOpMetricsManager())
			compressAndStore(storage, archiver, "abc", tempDir)
			RunRestore(restoreCmd, []string{"abc-001,abc"})
			output := readOutputFromFile(t)

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

			archiver := archive.NewShellOutArchiver(metrics.NewNoOpMetricsManager())
			compressAndStore(storage, archiver, "abc", tempDir)
			RunRestore(restoreCmd, []string{"^abc"})
			output := readOutputFromFile(t)

			restoredPath := filepath.FromSlash(fmt.Sprintf("%s/", tempDir))
			assert.Contains(t, output, "HIT: '^abc', using key 'abc'.")
			assert.Contains(t, output, fmt.Sprintf("Restored: %s.", restoredPath))

			os.Remove(tempFile.Name())
			os.Remove(tempDir)
		})
	})

	runTestForSingleBackend(t, "sftp", func(storage storage.Storage) {
		t.Run("restoring using HTTP works", func(t *testing.T) {
			storage.Clear()

			tempDir, _ := ioutil.TempDir(os.TempDir(), "*")
			tempFile, _ := ioutil.TempFile(tempDir, "*")
			_ = tempFile.Close()

			archiver := archive.NewShellOutArchiver(metrics.NewNoOpMetricsManager())
			compressAndStore(storage, archiver, "abc", tempDir)

			// set the environment variables to download using HTTP instead before restoring
			os.Setenv("SEMAPHORE_CACHE_CDN_URL", "http://sftp-server:80")
			os.Setenv("SEMAPHORE_CACHE_CDN_KEY", "test")
			os.Setenv("SEMAPHORE_CACHE_CDN_SECRET", "test")
			defer func() {
				os.Unsetenv("SEMAPHORE_CACHE_CDN_URL")
				os.Unsetenv("SEMAPHORE_CACHE_CDN_KEY")
				os.Unsetenv("SEMAPHORE_CACHE_CDN_SECRET")
			}()

			RunRestore(restoreCmd, []string{"^abc"})
			output := readOutputFromFile(t)

			restoredPath := filepath.FromSlash(fmt.Sprintf("%s/", tempDir))
			assert.Contains(t, output, "HIT: '^abc', using key 'abc'.")
			assert.Contains(t, output, fmt.Sprintf("Restored: %s.", restoredPath))

			os.Remove(tempFile.Name())
			os.Remove(tempDir)
		})

		t.Run("restoring defaults to use SFTP if not all variables are available", func(t *testing.T) {
			storage.Clear()

			tempDir, _ := ioutil.TempDir(os.TempDir(), "*")
			tempFile, _ := ioutil.TempFile(tempDir, "*")
			_ = tempFile.Close()

			archiver := archive.NewShellOutArchiver(metrics.NewNoOpMetricsManager())
			compressAndStore(storage, archiver, "abc", tempDir)

			// Set just the URL, but not the user/pass
			// This means SFTP will still be used.
			os.Setenv("SEMAPHORE_CACHE_CDN_URL", "http://sftp-server:80")
			defer func() {
				os.Unsetenv("SEMAPHORE_CACHE_CDN_URL")
			}()

			RunRestore(restoreCmd, []string{"^abc"})
			output := readOutputFromFile(t)

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

	log.SetFormatter(new(logging.CustomFormatter))
	log.SetLevel(log.InfoLevel)
	log.SetOutput(openLogfileForTests(t))

	runTestForAllBackends(t, func(backend string, storage storage.Storage) {
		t.Run(fmt.Sprintf("%s nothing found", backend), func(t *testing.T) {
			storage.Clear()
			os.Chdir(cmdPath)

			RunRestore(restoreCmd, []string{})
			output := readOutputFromFile(t)

			assert.Contains(t, output, "Nothing to restore from cache")
		})

		t.Run(fmt.Sprintf("%s detects and restores using SEMAPHORE_GIT_BRANCH", backend), func(t *testing.T) {
			storage.Clear()

			os.Chdir(fmt.Sprintf("%s/test/autocache/gems", rootPath))
			os.Setenv("SEMAPHORE_GIT_BRANCH", "master")
			os.Setenv("SEMAPHORE_GIT_PR_BRANCH", "")
			os.MkdirAll("vendor/bundle", os.ModePerm)

			// storing
			checksum, _ := files.GenerateChecksum("Gemfile.lock")
			key := fmt.Sprintf("gems-master-%s", checksum)
			compressedFile := filepath.Join(os.TempDir(), fmt.Sprintf("%s-%d", key, time.Now().Nanosecond()))
			archiver := archive.NewShellOutArchiver(metrics.NewNoOpMetricsManager())
			archiver.Compress(compressedFile, "vendor/bundle")
			storage.Store(key, compressedFile)

			// restoring
			RunRestore(restoreCmd, []string{})
			output := readOutputFromFile(t)

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
			os.Setenv("SEMAPHORE_GIT_PR_BRANCH", "some-development-branch")
			os.MkdirAll("vendor/bundle", os.ModePerm)

			// storing
			checksum, _ := files.GenerateChecksum("Gemfile.lock")
			key := fmt.Sprintf("gems-some-development-branch-%s", checksum)
			archiver := archive.NewShellOutArchiver(metrics.NewNoOpMetricsManager())
			compressedFile := filepath.Join(os.TempDir(), fmt.Sprintf("%s-%d", key, time.Now().Nanosecond()))
			archiver.Compress(compressedFile, "vendor/bundle")
			storage.Store(key, compressedFile)

			// restoring
			RunRestore(restoreCmd, []string{})
			output := readOutputFromFile(t)

			assert.Contains(t, output, "Detected Gemfile.lock")
			assert.Contains(t, output, fmt.Sprintf("Downloading key '%s'", key))
			assert.Contains(t, output, fmt.Sprintf("Restored: %s", filepath.FromSlash("vendor/bundle")))

			os.RemoveAll("vendor")
			os.Remove(compressedFile)
		})
	})
}
