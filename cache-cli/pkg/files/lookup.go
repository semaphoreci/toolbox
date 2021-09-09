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
	Key  string
	Path string
}

func Lookup() []LookupResult {
	cwd, _ := os.Getwd()
	results := []LookupResult{}
	for _, lockFile := range lockFiles {
		if _, err := os.Stat(fmt.Sprintf("%s/%s", cwd, lockFile)); err == nil {
			resultForFile := resultForfile(lockFile)
			results = append(results, *resultForFile)
		}
	}

	return results
}

func resultForfile(file string) *LookupResult {
	gitBranch := os.Getenv("SEMAPHORE_GIT_BRANCH")
	homedir, _ := os.UserHomeDir()

	switch file {
	case ".nvmrc":
		return buildResult(file, gitBranch, []LookupResultEntry{
			{"nvm", fmt.Sprintf("%s/.nvm", homedir)},
		})
	case "Gemfile.lock":
		return buildResult(file, gitBranch, []LookupResultEntry{
			{"gems", "vendor/bundle"},
		})
	case "package-lock.json":
		return buildResult(file, gitBranch, []LookupResultEntry{
			{"node-modules", "node_modules"},
		})
	case "yarn.lock":
		return buildResult(file, gitBranch, []LookupResultEntry{
			{"yarn-cache", fmt.Sprintf("%s/.cache/yarn", homedir)},
			{"node-modules", "node_modules"},
		})
	case "mix.lock":
		return buildResult(file, gitBranch, []LookupResultEntry{
			{"mix-deps", "deps"},
			{"mix-build", "_build"},
		})
	case "requirements.txt":
		return buildResult(file, gitBranch, []LookupResultEntry{
			{"requirements", ".pip_cache"},
		})
	case "composer.lock":
		return buildResult(file, gitBranch, []LookupResultEntry{
			{"composer", "vendor"},
		})
	case "pom.xml":
		return buildResult(file, gitBranch, []LookupResultEntry{
			{"maven", ".m2"},
			{"maven-target", "target"},
		})
	case "Podfile.lock":
		return buildResult(file, gitBranch, []LookupResultEntry{
			{"pods", "Pods"},
		})
	case "go.sum":
		return buildResult(file, gitBranch, []LookupResultEntry{
			{"go", fmt.Sprintf("%s/go/pkg/mod", homedir)},
		})
	default:
		fmt.Printf("Missing branch for %s - this should never happen!\n", file)
		return nil
	}
}

func buildResult(file, gitBranch string, entries []LookupResultEntry) *LookupResult {
	checksum, err := generateChecksum(file)
	if err != nil {
		fmt.Printf("Error generating checksum for %s: %v\n", file, err)
		return nil
	} else {
		newEntries := []LookupResultEntry{}
		for _, entry := range entries {
			key := fmt.Sprintf("%s-%s-%s", entry.Key, gitBranch, checksum)
			newEntries = append(newEntries, LookupResultEntry{
				Key:  normalizeKey(key),
				Path: entry.Path,
			})
		}

		return &LookupResult{
			DetectedFile: file,
			Entries:      newEntries,
		}
	}
}

func normalizeKey(key string) string {
	return strings.ReplaceAll(key, "/", "-")
}
