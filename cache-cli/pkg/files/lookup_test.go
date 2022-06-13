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
		results := Lookup(LookupOptions{Restore: false, LookupDirectory: testFilePath})
		assert.Empty(t, results)
	})

	t.Run("finds .nvmrc", func(t *testing.T) {
		checksum, err := GenerateChecksum(fmt.Sprintf("%s/test/autocache/nvm/.nvmrc", rootPath))
		assert.Nil(t, err)

		lookupDirectory := fmt.Sprintf("%s/test/autocache/nvm", rootPath)
		assertLookupResults(t, Lookup(LookupOptions{Restore: false, LookupDirectory: lookupDirectory}), []LookupResult{
			{
				DetectedFile: ".nvmrc",
				Entries: []LookupResultEntry{
					{
						Path: filepath.FromSlash(fmt.Sprintf("%s/.nvm", homedir)),
						Keys: []string{fmt.Sprintf("nvm-master-%s", checksum)},
					},
				},
			},
		})
	})

	t.Run("finds Gemfile.lock", func(t *testing.T) {
		checksum, err := GenerateChecksum(fmt.Sprintf("%s/test/autocache/gems/Gemfile.lock", rootPath))
		assert.Nil(t, err)

		lookupDirectory := fmt.Sprintf("%s/test/autocache/gems", rootPath)
		assertLookupResults(t, Lookup(LookupOptions{Restore: false, LookupDirectory: lookupDirectory}), []LookupResult{
			{
				DetectedFile: "Gemfile.lock",
				Entries: []LookupResultEntry{
					{
						Path: filepath.FromSlash("vendor/bundle"),
						Keys: []string{fmt.Sprintf("gems-master-%s", checksum)},
					},
				},
			},
		})
	})

	t.Run("finds package-lock.json", func(t *testing.T) {
		checksum, err := GenerateChecksum(fmt.Sprintf("%s/test/autocache/npm/package-lock.json", rootPath))
		assert.Nil(t, err)

		lookupDirectory := fmt.Sprintf("%s/test/autocache/npm", rootPath)
		assertLookupResults(t, Lookup(LookupOptions{Restore: false, LookupDirectory: lookupDirectory}), []LookupResult{
			{
				DetectedFile: "package-lock.json",
				Entries: []LookupResultEntry{
					{Path: "node_modules", Keys: []string{fmt.Sprintf("node-modules-master-%s", checksum)}},
				},
			},
		})
	})

	t.Run("finds requirements.txt", func(t *testing.T) {
		checksum, err := GenerateChecksum(fmt.Sprintf("%s/test/autocache/pip/requirements.txt", rootPath))
		assert.Nil(t, err)

		lookupDirectory := fmt.Sprintf("%s/test/autocache/pip", rootPath)
		assertLookupResults(t, Lookup(LookupOptions{Restore: false, LookupDirectory: lookupDirectory}), []LookupResult{
			{
				DetectedFile: "requirements.txt",
				Entries: []LookupResultEntry{
					{Path: ".pip_cache", Keys: []string{fmt.Sprintf("requirements-master-%s", checksum)}},
				},
			},
		})
	})

	t.Run("finds composer.lock", func(t *testing.T) {
		checksum, err := GenerateChecksum(fmt.Sprintf("%s/test/autocache/composer/composer.lock", rootPath))
		assert.Nil(t, err)

		lookupDirectory := fmt.Sprintf("%s/test/autocache/composer", rootPath)
		assertLookupResults(t, Lookup(LookupOptions{Restore: false, LookupDirectory: lookupDirectory}), []LookupResult{
			{
				DetectedFile: "composer.lock",
				Entries: []LookupResultEntry{
					{Path: "vendor", Keys: []string{fmt.Sprintf("composer-master-%s", checksum)}},
				},
			},
		})
	})

	t.Run("finds Podfile.lock", func(t *testing.T) {
		checksum, err := GenerateChecksum(fmt.Sprintf("%s/test/autocache/cocoapods/Podfile.lock", rootPath))
		assert.Nil(t, err)

		lookupDirectory := fmt.Sprintf("%s/test/autocache/cocoapods", rootPath)
		assertLookupResults(t, Lookup(LookupOptions{Restore: false, LookupDirectory: lookupDirectory}), []LookupResult{
			{
				DetectedFile: "Podfile.lock",
				Entries: []LookupResultEntry{
					{Path: "Pods", Keys: []string{fmt.Sprintf("pods-master-%s", checksum)}},
				},
			},
		})
	})

	t.Run("finds go.sum", func(t *testing.T) {
		checksum, err := GenerateChecksum(fmt.Sprintf("%s/test/autocache/go/go.sum", rootPath))
		assert.Nil(t, err)

		lookupDirectory := fmt.Sprintf("%s/test/autocache/go", rootPath)
		assertLookupResults(t, Lookup(LookupOptions{Restore: false, LookupDirectory: lookupDirectory}), []LookupResult{
			{
				DetectedFile: "go.sum",
				Entries: []LookupResultEntry{
					{
						Path: filepath.FromSlash(fmt.Sprintf("%s/go/pkg/mod", homedir)),
						Keys: []string{fmt.Sprintf("go-master-%s", checksum)},
					},
				},
			},
		})
	})

	t.Run("finds yarn.lock", func(t *testing.T) {
		checksum, err := GenerateChecksum(fmt.Sprintf("%s/test/autocache/yarn/yarn.lock", rootPath))
		assert.Nil(t, err)

		lookupDirectory := fmt.Sprintf("%s/test/autocache/yarn", rootPath)
		assertLookupResults(t, Lookup(LookupOptions{Restore: false, LookupDirectory: lookupDirectory}), []LookupResult{
			{
				DetectedFile: "yarn.lock",
				Entries: []LookupResultEntry{
					{
						Path: filepath.FromSlash(fmt.Sprintf("%s/.cache/yarn", homedir)),
						Keys: []string{fmt.Sprintf("yarn-cache-master-%s", checksum)},
					},
					{
						Path: "node_modules",
						Keys: []string{fmt.Sprintf("node-modules-master-%s", checksum)},
					},
				},
			},
		})
	})

	t.Run("finds mix.lock", func(t *testing.T) {
		checksum, err := GenerateChecksum(fmt.Sprintf("%s/test/autocache/elixir/mix.lock", rootPath))
		assert.Nil(t, err)

		lookupDirectory := fmt.Sprintf("%s/test/autocache/elixir", rootPath)
		assertLookupResults(t, Lookup(LookupOptions{Restore: false, LookupDirectory: lookupDirectory}), []LookupResult{
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
		checksum, err := GenerateChecksum(fmt.Sprintf("%s/test/autocache/maven/pom.xml", rootPath))
		assert.Nil(t, err)

		lookupDirectory := fmt.Sprintf("%s/test/autocache/maven", rootPath)
		assertLookupResults(t, Lookup(LookupOptions{Restore: false, LookupDirectory: lookupDirectory}), []LookupResult{
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
		requirementsChecksum, err := GenerateChecksum(fmt.Sprintf("%s/test/autocache/multiple-files/requirements.txt", rootPath))
		assert.Nil(t, err)

		packageLockChecksum, err := GenerateChecksum(fmt.Sprintf("%s/test/autocache/multiple-files/package-lock.json", rootPath))
		assert.Nil(t, err)

		lookupDirectory := fmt.Sprintf("%s/test/autocache/multiple-files", rootPath)
		assertLookupResults(t, Lookup(LookupOptions{Restore: false, LookupDirectory: lookupDirectory}), []LookupResult{
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

func Test__LookupRestore(t *testing.T) {
	homedir, _ := os.UserHomeDir()

	// TODO: find a better way to find the root path
	_, b, _, _ := runtime.Caller(0)
	testFilePath := filepath.Dir(b)
	pkgPath := filepath.Dir(testFilePath)
	rootPath := filepath.Dir(pkgPath)

	t.Run("finds nothing", func(t *testing.T) {
		results := Lookup(LookupOptions{Restore: true, LookupDirectory: testFilePath})
		assert.Empty(t, results)
	})

	t.Run("finds .nvmrc", func(t *testing.T) {
		checksum, err := GenerateChecksum(fmt.Sprintf("%s/test/autocache/nvm/.nvmrc", rootPath))
		assert.Nil(t, err)

		lookupDirectory := fmt.Sprintf("%s/test/autocache/nvm", rootPath)
		assertLookupResults(t, Lookup(LookupOptions{Restore: true, GitBranch: "some-branch", LookupDirectory: lookupDirectory}), []LookupResult{
			{
				DetectedFile: ".nvmrc",
				Entries: []LookupResultEntry{
					{
						Path: filepath.FromSlash(fmt.Sprintf("%s/.nvm", homedir)),
						Keys: []string{
							fmt.Sprintf("nvm-some-branch-%s", checksum),
							"nvm-some-branch",
							"nvm-master",
							"nvm-main",
						},
					},
				},
			},
		})
	})

	t.Run("finds Gemfile.lock", func(t *testing.T) {
		checksum, err := GenerateChecksum(fmt.Sprintf("%s/test/autocache/gems/Gemfile.lock", rootPath))
		assert.Nil(t, err)

		lookupDirectory := fmt.Sprintf("%s/test/autocache/gems", rootPath)
		assertLookupResults(t, Lookup(LookupOptions{Restore: true, GitBranch: "some-branch", LookupDirectory: lookupDirectory}), []LookupResult{
			{
				DetectedFile: "Gemfile.lock",
				Entries: []LookupResultEntry{
					{
						Path: filepath.FromSlash("vendor/bundle"),
						Keys: []string{
							fmt.Sprintf("gems-some-branch-%s", checksum),
							"gems-some-branch",
							"gems-master",
							"gems-main",
						},
					},
				},
			},
		})
	})

	t.Run("finds package-lock.json", func(t *testing.T) {
		checksum, err := GenerateChecksum(fmt.Sprintf("%s/test/autocache/npm/package-lock.json", rootPath))
		assert.Nil(t, err)

		lookupDirectory := fmt.Sprintf("%s/test/autocache/npm", rootPath)
		assertLookupResults(t, Lookup(LookupOptions{Restore: true, GitBranch: "some-branch", LookupDirectory: lookupDirectory}), []LookupResult{
			{
				DetectedFile: "package-lock.json",
				Entries: []LookupResultEntry{
					{Path: "node_modules", Keys: []string{
						fmt.Sprintf("node-modules-some-branch-%s", checksum),
						"node-modules-some-branch",
						"node-modules-master",
						"node-modules-main",
					}},
				},
			},
		})
	})

	t.Run("finds requirements.txt", func(t *testing.T) {
		checksum, err := GenerateChecksum(fmt.Sprintf("%s/test/autocache/pip/requirements.txt", rootPath))
		assert.Nil(t, err)

		lookupDirectory := fmt.Sprintf("%s/test/autocache/pip", rootPath)
		assertLookupResults(t, Lookup(LookupOptions{Restore: true, GitBranch: "some-branch", LookupDirectory: lookupDirectory}), []LookupResult{
			{
				DetectedFile: "requirements.txt",
				Entries: []LookupResultEntry{
					{Path: ".pip_cache", Keys: []string{
						fmt.Sprintf("requirements-some-branch-%s", checksum),
						"requirements-some-branch",
						"requirements-master",
						"requirements-main",
					}},
				},
			},
		})
	})

	t.Run("finds composer.lock", func(t *testing.T) {
		checksum, err := GenerateChecksum(fmt.Sprintf("%s/test/autocache/composer/composer.lock", rootPath))
		assert.Nil(t, err)

		lookupDirectory := fmt.Sprintf("%s/test/autocache/composer", rootPath)
		assertLookupResults(t, Lookup(LookupOptions{Restore: true, GitBranch: "some-branch", LookupDirectory: lookupDirectory}), []LookupResult{
			{
				DetectedFile: "composer.lock",
				Entries: []LookupResultEntry{
					{Path: "vendor", Keys: []string{
						fmt.Sprintf("composer-some-branch-%s", checksum),
						"composer-some-branch",
						"composer-master",
						"composer-main",
					}},
				},
			},
		})
	})

	t.Run("finds Podfile.lock", func(t *testing.T) {
		checksum, err := GenerateChecksum(fmt.Sprintf("%s/test/autocache/cocoapods/Podfile.lock", rootPath))
		assert.Nil(t, err)

		lookupDirectory := fmt.Sprintf("%s/test/autocache/cocoapods", rootPath)
		assertLookupResults(t, Lookup(LookupOptions{Restore: true, GitBranch: "some-branch", LookupDirectory: lookupDirectory}), []LookupResult{
			{
				DetectedFile: "Podfile.lock",
				Entries: []LookupResultEntry{
					{Path: "Pods", Keys: []string{
						fmt.Sprintf("pods-some-branch-%s", checksum),
						"pods-some-branch",
						"pods-master",
						"pods-main",
					}},
				},
			},
		})
	})

	t.Run("finds go.sum", func(t *testing.T) {
		checksum, err := GenerateChecksum(fmt.Sprintf("%s/test/autocache/go/go.sum", rootPath))
		assert.Nil(t, err)

		lookupDirectory := fmt.Sprintf("%s/test/autocache/go", rootPath)
		assertLookupResults(t, Lookup(LookupOptions{Restore: true, GitBranch: "some-branch", LookupDirectory: lookupDirectory}), []LookupResult{
			{
				DetectedFile: "go.sum",
				Entries: []LookupResultEntry{
					{
						Path: filepath.FromSlash(fmt.Sprintf("%s/go/pkg/mod", homedir)),
						Keys: []string{
							fmt.Sprintf("go-some-branch-%s", checksum),
							"go-some-branch",
							"go-master",
							"go-main",
						},
					},
				},
			},
		})
	})

	t.Run("finds yarn.lock", func(t *testing.T) {
		checksum, err := GenerateChecksum(fmt.Sprintf("%s/test/autocache/yarn/yarn.lock", rootPath))
		assert.Nil(t, err)

		lookupDirectory := fmt.Sprintf("%s/test/autocache/yarn", rootPath)
		assertLookupResults(t, Lookup(LookupOptions{Restore: true, GitBranch: "some-branch", LookupDirectory: lookupDirectory}), []LookupResult{
			{
				DetectedFile: "yarn.lock",
				Entries: []LookupResultEntry{
					{
						Path: filepath.FromSlash(fmt.Sprintf("%s/.cache/yarn", homedir)),
						Keys: []string{
							fmt.Sprintf("yarn-cache-some-branch-%s", checksum),
							"yarn-cache-some-branch",
							"yarn-cache-master",
							"yarn-cache-main",
						},
					},
					{
						Path: "node_modules",
						Keys: []string{
							fmt.Sprintf("node-modules-some-branch-%s", checksum),
							"node-modules-some-branch",
							"node-modules-master",
							"node-modules-main",
						},
					},
				},
			},
		})
	})

	t.Run("finds mix.lock", func(t *testing.T) {
		checksum, err := GenerateChecksum(fmt.Sprintf("%s/test/autocache/elixir/mix.lock", rootPath))
		assert.Nil(t, err)

		lookupDirectory := fmt.Sprintf("%s/test/autocache/elixir", rootPath)
		assertLookupResults(t, Lookup(LookupOptions{Restore: true, GitBranch: "some-branch", LookupDirectory: lookupDirectory}), []LookupResult{
			{
				DetectedFile: "mix.lock",
				Entries: []LookupResultEntry{
					{Path: "deps", Keys: []string{
						fmt.Sprintf("mix-deps-some-branch-%s", checksum),
						"mix-deps-some-branch",
						"mix-deps-master",
						"mix-deps-main",
					}},
					{Path: "_build", Keys: []string{
						fmt.Sprintf("mix-build-some-branch-%s", checksum),
						"mix-build-some-branch",
						"mix-build-master",
						"mix-build-main",
					}},
				},
			},
		})
	})

	t.Run("finds pom.xml", func(t *testing.T) {
		checksum, err := GenerateChecksum(fmt.Sprintf("%s/test/autocache/maven/pom.xml", rootPath))
		assert.Nil(t, err)

		lookupDirectory := fmt.Sprintf("%s/test/autocache/maven", rootPath)
		assertLookupResults(t, Lookup(LookupOptions{Restore: true, GitBranch: "some-branch", LookupDirectory: lookupDirectory}), []LookupResult{
			{
				DetectedFile: "pom.xml",
				Entries: []LookupResultEntry{
					{Path: ".m2", Keys: []string{
						fmt.Sprintf("maven-some-branch-%s", checksum),
						"maven-some-branch",
						"maven-master",
						"maven-main",
					}},
					{Path: "target", Keys: []string{
						fmt.Sprintf("maven-target-some-branch-%s", checksum),
						"maven-target-some-branch",
						"maven-target-master",
						"maven-target-main",
					}},
				},
			},
		})
	})

	t.Run("finds requirements.txt and package-lock.json", func(t *testing.T) {
		requirementsChecksum, err := GenerateChecksum(fmt.Sprintf("%s/test/autocache/multiple-files/requirements.txt", rootPath))
		assert.Nil(t, err)

		packageLockChecksum, err := GenerateChecksum(fmt.Sprintf("%s/test/autocache/multiple-files/package-lock.json", rootPath))
		assert.Nil(t, err)

		lookupDirectory := fmt.Sprintf("%s/test/autocache/multiple-files", rootPath)
		assertLookupResults(t, Lookup(LookupOptions{Restore: true, GitBranch: "some-branch", LookupDirectory: lookupDirectory}), []LookupResult{
			{
				DetectedFile: "package-lock.json",
				Entries: []LookupResultEntry{
					{Path: "node_modules", Keys: []string{
						fmt.Sprintf("node-modules-some-branch-%s", packageLockChecksum),
						"node-modules-some-branch",
						"node-modules-master",
						"node-modules-main",
					}},
				},
			},
			{
				DetectedFile: "requirements.txt",
				Entries: []LookupResultEntry{
					{Path: ".pip_cache", Keys: []string{
						fmt.Sprintf("requirements-some-branch-%s", requirementsChecksum),
						"requirements-some-branch",
						"requirements-master",
						"requirements-main",
					}},
				},
			},
		})
	})

	t.Run("no duplicated master suffixed key", func(t *testing.T) {
		checksum, err := GenerateChecksum(fmt.Sprintf("%s/test/autocache/npm/package-lock.json", rootPath))
		assert.Nil(t, err)

		lookupDirectory := fmt.Sprintf("%s/test/autocache/npm", rootPath)
		assertLookupResults(t, Lookup(LookupOptions{Restore: true, GitBranch: "master", LookupDirectory: lookupDirectory}), []LookupResult{
			{
				DetectedFile: "package-lock.json",
				Entries: []LookupResultEntry{
					{Path: "node_modules", Keys: []string{
						fmt.Sprintf("node-modules-master-%s", checksum),
						"node-modules-master",
						"node-modules-main",
					}},
				},
			},
		})
	})

	t.Run("no duplicated main suffixed key", func(t *testing.T) {
		checksum, err := GenerateChecksum(fmt.Sprintf("%s/test/autocache/npm/package-lock.json", rootPath))
		assert.Nil(t, err)

		lookupDirectory := fmt.Sprintf("%s/test/autocache/npm", rootPath)
		assertLookupResults(t, Lookup(LookupOptions{Restore: true, GitBranch: "main", LookupDirectory: lookupDirectory}), []LookupResult{
			{
				DetectedFile: "package-lock.json",
				Entries: []LookupResultEntry{
					{Path: "node_modules", Keys: []string{
						fmt.Sprintf("node-modules-main-%s", checksum),
						"node-modules-main",
						"node-modules-master",
					}},
				},
			},
		})
	})
}

func assertLookupResults(t *testing.T, actualResults []LookupResult, expectedResults []LookupResult) {
	if assert.Len(t, actualResults, len(expectedResults)) {
		for resultIndex, result := range actualResults {
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
