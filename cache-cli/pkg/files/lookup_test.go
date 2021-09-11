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

		assertLookupStoreResults(t, []LookupResult{
			{
				DetectedFile: ".nvmrc",
				Entries: []LookupResultEntry{
					{Path: fmt.Sprintf("%s/.nvm", homedir), Keys: []string{fmt.Sprintf("nvm-master-%s", checksum)}},
				},
			},
		})
	})

	t.Run("finds Gemfile.lock", func(t *testing.T) {
		err := os.Chdir(fmt.Sprintf("%s/test/autocache/gems", rootPath))
		assert.Nil(t, err)

		checksum, err := generateChecksum("Gemfile.lock")
		assert.Nil(t, err)

		assertLookupStoreResults(t, []LookupResult{
			{
				DetectedFile: "Gemfile.lock",
				Entries: []LookupResultEntry{
					{Path: "vendor/bundle", Keys: []string{fmt.Sprintf("gems-master-%s", checksum)}},
				},
			},
		})
	})

	t.Run("finds package-lock.json", func(t *testing.T) {
		err := os.Chdir(fmt.Sprintf("%s/test/autocache/npm", rootPath))
		assert.Nil(t, err)

		checksum, err := generateChecksum("package-lock.json")
		assert.Nil(t, err)

		assertLookupStoreResults(t, []LookupResult{
			{
				DetectedFile: "package-lock.json",
				Entries: []LookupResultEntry{
					{Path: "node_modules", Keys: []string{fmt.Sprintf("node-modules-master-%s", checksum)}},
				},
			},
		})
	})

	t.Run("finds requirements.txt", func(t *testing.T) {
		err := os.Chdir(fmt.Sprintf("%s/test/autocache/pip", rootPath))
		assert.Nil(t, err)

		checksum, err := generateChecksum("requirements.txt")
		assert.Nil(t, err)

		assertLookupStoreResults(t, []LookupResult{
			{
				DetectedFile: "requirements.txt",
				Entries: []LookupResultEntry{
					{Path: ".pip_cache", Keys: []string{fmt.Sprintf("requirements-master-%s", checksum)}},
				},
			},
		})
	})

	t.Run("finds composer.lock", func(t *testing.T) {
		err := os.Chdir(fmt.Sprintf("%s/test/autocache/composer", rootPath))
		assert.Nil(t, err)

		checksum, err := generateChecksum("composer.lock")
		assert.Nil(t, err)

		assertLookupStoreResults(t, []LookupResult{
			{
				DetectedFile: "composer.lock",
				Entries: []LookupResultEntry{
					{Path: "vendor", Keys: []string{fmt.Sprintf("composer-master-%s", checksum)}},
				},
			},
		})
	})

	t.Run("finds Podfile.lock", func(t *testing.T) {
		err := os.Chdir(fmt.Sprintf("%s/test/autocache/cocoapods", rootPath))
		assert.Nil(t, err)

		checksum, err := generateChecksum("Podfile.lock")
		assert.Nil(t, err)

		assertLookupStoreResults(t, []LookupResult{
			{
				DetectedFile: "Podfile.lock",
				Entries: []LookupResultEntry{
					{Path: "Pods", Keys: []string{fmt.Sprintf("pods-master-%s", checksum)}},
				},
			},
		})
	})

	t.Run("finds go.sum", func(t *testing.T) {
		err := os.Chdir(fmt.Sprintf("%s/test/autocache/go", rootPath))
		assert.Nil(t, err)

		checksum, err := generateChecksum("go.sum")
		assert.Nil(t, err)

		assertLookupStoreResults(t, []LookupResult{
			{
				DetectedFile: "go.sum",
				Entries: []LookupResultEntry{
					{Path: fmt.Sprintf("%s/go/pkg/mod", homedir), Keys: []string{fmt.Sprintf("go-master-%s", checksum)}},
				},
			},
		})
	})

	t.Run("finds yarn.lock", func(t *testing.T) {
		err := os.Chdir(fmt.Sprintf("%s/test/autocache/yarn", rootPath))
		assert.Nil(t, err)

		checksum, err := generateChecksum("yarn.lock")
		assert.Nil(t, err)

		assertLookupStoreResults(t, []LookupResult{
			{
				DetectedFile: "yarn.lock",
				Entries: []LookupResultEntry{
					{Path: fmt.Sprintf("%s/.cache/yarn", homedir), Keys: []string{fmt.Sprintf("yarn-cache-master-%s", checksum)}},
					{Path: "node_modules", Keys: []string{fmt.Sprintf("node-modules-master-%s", checksum)}},
				},
			},
		})
	})

	t.Run("finds mix.lock", func(t *testing.T) {
		err := os.Chdir(fmt.Sprintf("%s/test/autocache/elixir", rootPath))
		assert.Nil(t, err)

		checksum, err := generateChecksum("mix.lock")
		assert.Nil(t, err)

		assertLookupStoreResults(t, []LookupResult{
			{
				DetectedFile: "mix.lock",
				Entries: []LookupResultEntry{
					{Path: "deps", Keys: []string{fmt.Sprintf("mix-deps-master-%s", checksum)}},
					{Path: "_build", Keys: []string{fmt.Sprintf("mix-build-master-%s", checksum)}},
				},
			},
		})
	})

	t.Run("finds pom.xml", func(t *testing.T) {
		err := os.Chdir(fmt.Sprintf("%s/test/autocache/maven", rootPath))
		assert.Nil(t, err)

		checksum, err := generateChecksum("pom.xml")
		assert.Nil(t, err)

		assertLookupStoreResults(t, []LookupResult{
			{
				DetectedFile: "pom.xml",
				Entries: []LookupResultEntry{
					{Path: ".m2", Keys: []string{fmt.Sprintf("maven-master-%s", checksum)}},
					{Path: "target", Keys: []string{fmt.Sprintf("maven-target-master-%s", checksum)}},
				},
			},
		})
	})

	t.Run("finds requirements.txt and package-lock.json", func(t *testing.T) {
		err := os.Chdir(fmt.Sprintf("%s/test/autocache/multiple-files", rootPath))
		assert.Nil(t, err)

		requirementsChecksum, err := generateChecksum("requirements.txt")
		assert.Nil(t, err)

		packageLockChecksum, err := generateChecksum("package-lock.json")
		assert.Nil(t, err)

		assertLookupStoreResults(t, []LookupResult{
			{
				DetectedFile: "package-lock.json",
				Entries: []LookupResultEntry{
					{Path: "node_modules", Keys: []string{fmt.Sprintf("node-modules-master-%s", packageLockChecksum)}},
				},
			},
			{
				DetectedFile: "requirements.txt",
				Entries: []LookupResultEntry{
					{Path: ".pip_cache", Keys: []string{fmt.Sprintf("requirements-master-%s", requirementsChecksum)}},
				},
			},
		})
	})
}

func assertLookupStoreResults(t *testing.T, expectedResults []LookupResult) {
	results := LookupStore()
	if assert.Len(t, results, len(expectedResults)) {
		for resultIndex, result := range results {
			expectedResult := expectedResults[resultIndex]
			assert.Equal(t, result.DetectedFile, expectedResult.DetectedFile)
			if assert.Len(t, result.Entries, len(expectedResult.Entries)) {
				for index, entry := range result.Entries {
					expectedEntry := expectedResult.Entries[index]
					assert.Equal(t, entry.Path, expectedEntry.Path)
					assert.Equal(t, entry.Keys, expectedEntry.Keys)
				}
			}
		}
	}
}
