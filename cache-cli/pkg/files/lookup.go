package files

import (
	"fmt"
	"os"
	"path/filepath"
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

type LookupOptions struct {
	LookupDirectory string
	HomeDirectory   string
	GitBranch       string
	Restore         bool
}

type LookupResult struct {
	DetectedFile string
	Entries      []LookupResultEntry
}

type LookupResultEntry struct {
	Keys []string
	Path string
}

func Lookup(options LookupOptions) []LookupResult {
	lookupDirectory := options.LookupDirectory
	if lookupDirectory == "" {
		lookupDirectory, _ = os.Getwd()
	}

	results := []LookupResult{}
	for _, lockFile := range lockFiles {
		lockFilePath := fmt.Sprintf("%s/%s", lookupDirectory, lockFile)
		if _, err := os.Stat(lockFilePath); err == nil {
			resultForFile := resultForfile(lockFilePath, options)
			results = append(results, *resultForFile)
		}
	}

	return results
}

func resultForfile(filePath string, options LookupOptions) *LookupResult {
	homedir := options.HomeDirectory
	if homedir == "" {
		homedir, _ = os.UserHomeDir()
	}

	file := filepath.Base(filePath)

	switch file {
	case ".nvmrc":
		return buildResult(filePath, options, []buildResultRequest{
			{"nvm", fmt.Sprintf("%s/.nvm", homedir)},
		})
	case "Gemfile.lock":
		return buildResult(filePath, options, []buildResultRequest{
			{"gems", "vendor/bundle"},
		})
	case "package-lock.json":
		return buildResult(filePath, options, []buildResultRequest{
			{"node-modules", "node_modules"},
		})
	case "yarn.lock":
		return buildResult(filePath, options, []buildResultRequest{
			{"yarn-cache", fmt.Sprintf("%s/.cache/yarn", homedir)},
			{"node-modules", "node_modules"},
		})
	case "mix.lock":
		return buildResult(filePath, options, []buildResultRequest{
			{"mix-deps", "deps"},
			{"mix-build", "_build"},
		})
	case "requirements.txt":
		return buildResult(filePath, options, []buildResultRequest{
			{"requirements", ".pip_cache"},
		})
	case "composer.lock":
		return buildResult(filePath, options, []buildResultRequest{
			{"composer", "vendor"},
		})
	case "pom.xml":
		return buildResult(filePath, options, []buildResultRequest{
			{"maven", ".m2"},
			{"maven-target", "target"},
		})
	case "Podfile.lock":
		return buildResult(filePath, options, []buildResultRequest{
			{"pods", "Pods"},
		})
	case "go.sum":
		return buildResult(filePath, options, []buildResultRequest{
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

func buildResult(filePath string, options LookupOptions, entries []buildResultRequest) *LookupResult {
	gitBranch := options.GitBranch
	if gitBranch == "" {
		gitBranch = "master"
	}

	checksum, err := GenerateChecksum(filePath)
	if err != nil {
		fmt.Printf("Error generating checksum for %s: %v\n", filePath, err)
		return nil
	}

	newEntries := []LookupResultEntry{}
	for _, entry := range entries {
		if options.Restore {
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
		DetectedFile: filepath.Base(filePath),
		Entries:      newEntries,
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
