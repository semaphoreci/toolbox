package files

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__LookupStore(t *testing.T) {
	homedir, _ := os.UserHomeDir()

	// TODO: find a better way to find the root path
	_, b, _, _ := runtime.Caller(0)
	testFilePath := filepath.Dir(b)
	pkgPath := filepath.Dir(testFilePath)
	rootPath := filepath.Dir(pkgPath)

	t.Run("finds nothing", func(t *testing.T) {
		results := LookupStore()
		assert.Empty(t, results)
	})

	t.Run("finds .nvmrc", func(t *testing.T) {
		err := os.Chdir(fmt.Sprintf("%s/test/autocache/nvm", rootPath))
		assert.Nil(t, err)

		checksum, err := generateChecksum(".nvmrc")
		assert.Nil(t, err)

		assertLookupStoreResult(t, ".nvmrc", []LookupResultEntry{
			{Path: fmt.Sprintf("%s/.nvm", homedir), Keys: []string{fmt.Sprintf("nvm-master-%s", checksum)}},
		})
	})

	t.Run("finds Gemfile.lock", func(t *testing.T) {
		err := os.Chdir(fmt.Sprintf("%s/test/autocache/gems", rootPath))
		assert.Nil(t, err)

		checksum, err := generateChecksum("Gemfile.lock")
		assert.Nil(t, err)

		assertLookupStoreResult(t, "Gemfile.lock", []LookupResultEntry{
			{Path: "vendor/bundle", Keys: []string{fmt.Sprintf("gems-master-%s", checksum)}},
		})
	})

	t.Run("finds package-lock.json", func(t *testing.T) {
		err := os.Chdir(fmt.Sprintf("%s/test/autocache/npm", rootPath))
		assert.Nil(t, err)

		checksum, err := generateChecksum("package-lock.json")
		assert.Nil(t, err)

		assertLookupStoreResult(t, "package-lock.json", []LookupResultEntry{
			{Path: "node_modules", Keys: []string{fmt.Sprintf("node-modules-master-%s", checksum)}},
		})
	})

	t.Run("finds requirements.txt", func(t *testing.T) {
		err := os.Chdir(fmt.Sprintf("%s/test/autocache/pip", rootPath))
		assert.Nil(t, err)

		checksum, err := generateChecksum("requirements.txt")
		assert.Nil(t, err)

		assertLookupStoreResult(t, "requirements.txt", []LookupResultEntry{
			{Path: ".pip_cache", Keys: []string{fmt.Sprintf("requirements-master-%s", checksum)}},
		})
	})

	t.Run("finds composer.lock", func(t *testing.T) {
		err := os.Chdir(fmt.Sprintf("%s/test/autocache/composer", rootPath))
		assert.Nil(t, err)

		checksum, err := generateChecksum("composer.lock")
		assert.Nil(t, err)

		assertLookupStoreResult(t, "composer.lock", []LookupResultEntry{
			{Path: "vendor", Keys: []string{fmt.Sprintf("composer-master-%s", checksum)}},
		})
	})

	t.Run("finds Podfile.lock", func(t *testing.T) {
		err := os.Chdir(fmt.Sprintf("%s/test/autocache/cocoapods", rootPath))
		assert.Nil(t, err)

		checksum, err := generateChecksum("Podfile.lock")
		assert.Nil(t, err)

		assertLookupStoreResult(t, "Podfile.lock", []LookupResultEntry{
			{Path: "Pods", Keys: []string{fmt.Sprintf("pods-master-%s", checksum)}},
		})
	})

	t.Run("finds go.sum", func(t *testing.T) {
		err := os.Chdir(fmt.Sprintf("%s/test/autocache/go", rootPath))
		assert.Nil(t, err)

		checksum, err := generateChecksum("go.sum")
		assert.Nil(t, err)

		assertLookupStoreResult(t, "go.sum", []LookupResultEntry{
			{Path: fmt.Sprintf("%s/go/pkg/mod", homedir), Keys: []string{fmt.Sprintf("go-master-%s", checksum)}},
		})
	})

	t.Run("finds yarn.lock", func(t *testing.T) {
		err := os.Chdir(fmt.Sprintf("%s/test/autocache/yarn", rootPath))
		assert.Nil(t, err)

		checksum, err := generateChecksum("yarn.lock")
		assert.Nil(t, err)

		assertLookupStoreResult(t, "yarn.lock", []LookupResultEntry{
			{Path: fmt.Sprintf("%s/.cache/yarn", homedir), Keys: []string{fmt.Sprintf("yarn-cache-master-%s", checksum)}},
			{Path: "node_modules", Keys: []string{fmt.Sprintf("node-modules-master-%s", checksum)}},
		})
	})

	t.Run("finds mix.lock", func(t *testing.T) {
		err := os.Chdir(fmt.Sprintf("%s/test/autocache/elixir", rootPath))
		assert.Nil(t, err)

		checksum, err := generateChecksum("mix.lock")
		assert.Nil(t, err)

		assertLookupStoreResult(t, "mix.lock", []LookupResultEntry{
			{Path: "deps", Keys: []string{fmt.Sprintf("mix-deps-master-%s", checksum)}},
			{Path: "_build", Keys: []string{fmt.Sprintf("mix-build-master-%s", checksum)}},
		})
	})

	t.Run("finds pom.xml", func(t *testing.T) {
		err := os.Chdir(fmt.Sprintf("%s/test/autocache/maven", rootPath))
		assert.Nil(t, err)

		checksum, err := generateChecksum("pom.xml")
		assert.Nil(t, err)

		assertLookupStoreResult(t, "pom.xml", []LookupResultEntry{
			{Path: ".m2", Keys: []string{fmt.Sprintf("maven-master-%s", checksum)}},
			{Path: "target", Keys: []string{fmt.Sprintf("maven-target-master-%s", checksum)}},
		})
	})
}

func assertLookupStoreResult(t *testing.T, detectedFile string, entries []LookupResultEntry) {
	results := LookupStore()
	if assert.Len(t, results, 1) {
		result := results[0]
		assert.Equal(t, result.DetectedFile, detectedFile)
		if assert.Len(t, result.Entries, len(entries)) {
			for index, entry := range result.Entries {
				expectedEntry := entries[index]
				assert.Equal(t, entry.Path, expectedEntry.Path)
				assert.Equal(t, entry.Keys, expectedEntry.Keys)
			}
		}
	}
}
