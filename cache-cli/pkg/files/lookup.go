package files

import (
	"fmt"
	"os"
	"strings"
)

var lockFiles = []string{
	".nvmrc",
	"Gemfile.lock",
	"package-lock.json",
	"yarn.lock",
	"mix.lock",
	"requirements.txt",
	"composer.lock",
	"pom.xml",
	"Podfile.lock",
	"go.sum",
}

type LookupResult struct {
	DetectedFile string
	Entries      []LookupResultEntry
}

type LookupResultEntry struct {
	Keys []string
	Path string
}

func LookupStore() []LookupResult {
	return Lookup(false)
}

func LookupRestore() []LookupResult {
	return Lookup(true)
}

func Lookup(restore bool) []LookupResult {
	cwd, _ := os.Getwd()
	results := []LookupResult{}
	for _, lockFile := range lockFiles {
		if _, err := os.Stat(fmt.Sprintf("%s/%s", cwd, lockFile)); err == nil {
			resultForFile := resultForfile(lockFile, restore)
			results = append(results, *resultForFile)
		}
	}

	return results
}

func resultForfile(file string, restore bool) *LookupResult {
	gitBranch := os.Getenv("SEMAPHORE_GIT_BRANCH")
	if gitBranch == "" {
		gitBranch = "master"
	}

	homedir, _ := os.UserHomeDir()

	switch file {
	case ".nvmrc":
		return buildResult(file, gitBranch, restore, []buildResultRequest{
			{"nvm", fmt.Sprintf("%s/.nvm", homedir)},
		})
	case "Gemfile.lock":
		return buildResult(file, gitBranch, restore, []buildResultRequest{
			{"gems", "vendor/bundle"},
		})
	case "package-lock.json":
		return buildResult(file, gitBranch, restore, []buildResultRequest{
			{"node-modules", "node_modules"},
		})
	case "yarn.lock":
		return buildResult(file, gitBranch, restore, []buildResultRequest{
			{"yarn-cache", fmt.Sprintf("%s/.cache/yarn", homedir)},
			{"node-modules", "node_modules"},
		})
	case "mix.lock":
		return buildResult(file, gitBranch, restore, []buildResultRequest{
			{"mix-deps", "deps"},
			{"mix-build", "_build"},
		})
	case "requirements.txt":
		return buildResult(file, gitBranch, restore, []buildResultRequest{
			{"requirements", ".pip_cache"},
		})
	case "composer.lock":
		return buildResult(file, gitBranch, restore, []buildResultRequest{
			{"composer", "vendor"},
		})
	case "pom.xml":
		return buildResult(file, gitBranch, restore, []buildResultRequest{
			{"maven", ".m2"},
			{"maven-target", "target"},
		})
	case "Podfile.lock":
		return buildResult(file, gitBranch, restore, []buildResultRequest{
			{"pods", "Pods"},
		})
	case "go.sum":
		return buildResult(file, gitBranch, restore, []buildResultRequest{
			{"go", fmt.Sprintf("%s/go/pkg/mod", homedir)},
		})
	default:
		fmt.Printf("Missing branch for %s - this should never happen!\n", file)
		return nil
	}
}

type buildResultRequest struct {
	KeyPrefix string
	Path      string
}

func buildResult(file, gitBranch string, restore bool, entries []buildResultRequest) *LookupResult {
	checksum, err := generateChecksum(file)
	if err != nil {
		fmt.Printf("Error generating checksum for %s: %v\n", file, err)
		return nil
	} else {
		newEntries := []LookupResultEntry{}
		for _, entry := range entries {
			if restore {
				newEntries = append(newEntries, LookupResultEntry{
					Path: entry.Path,
					Keys: keysForRestore(entry.KeyPrefix, gitBranch, checksum),
				})
			} else {
				key := fmt.Sprintf("%s-%s-%s", entry.KeyPrefix, gitBranch, checksum)
				newEntries = append(newEntries, LookupResultEntry{
					Keys: []string{normalizeKey(key)},
					Path: entry.Path,
				})
			}
		}

		return &LookupResult{
			DetectedFile: file,
			Entries:      newEntries,
		}
	}
}

func keysForRestore(keyPrefix, gitBranch, checksum string) []string {
	keys := []string{
		normalizeKey(fmt.Sprintf("%s-%s-%s", keyPrefix, gitBranch, checksum)),
		normalizeKey(fmt.Sprintf("%s-%s", keyPrefix, gitBranch)),
	}

	if gitBranch != "master" {
		keys = append(keys, normalizeKey(fmt.Sprintf("%s-master", keyPrefix)))
	}

	return keys
}

func normalizeKey(key string) string {
	return strings.ReplaceAll(key, "/", "-")
}
