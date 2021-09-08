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
	Key          string
	Path         string
}

func Lookup() []LookupResult {
	cwd, _ := os.Getwd()
	results := []LookupResult{}
	for _, lockFile := range lockFiles {
		if _, err := os.Stat(fmt.Sprintf("%s/%s", cwd, lockFile)); err == nil {
			resultsForFile := resultForfile(lockFile)
			results = append(results, resultsForFile...)
		}
	}

	return results
}

func resultForfile(file string) []LookupResult {
	results := []LookupResult{}
	gitBranch := os.Getenv("SEMAPHORE_GIT_BRANCH")
	homedir, _ := os.UserHomeDir()

	switch file {
	case ".nvmrc":
		buildResult(&results, buildResultOpts{
			file:      file,
			path:      fmt.Sprintf("%s/.nvm", homedir),
			gitBranch: gitBranch,
			keyPrefix: "nvm",
		})
	case "Gemfile.lock":
		buildResult(&results, buildResultOpts{
			file:      file,
			path:      "vendor/bundle",
			gitBranch: gitBranch,
			keyPrefix: "gems",
		})
	case "package-lock.json":
		buildResult(&results, buildResultOpts{
			file:      file,
			path:      "node_modules",
			gitBranch: gitBranch,
			keyPrefix: "node-modules",
		})
	case "yarn.lock":
		buildResult(&results, buildResultOpts{
			file:      file,
			path:      fmt.Sprintf("%s/.cache/yarn", homedir),
			gitBranch: gitBranch,
			keyPrefix: "yarn-cache",
		})
		buildResult(&results, buildResultOpts{
			file:      file,
			path:      "node_modules",
			gitBranch: gitBranch,
			keyPrefix: "node-modules",
		})
	case "mix.lock":
		buildResult(&results, buildResultOpts{
			file:      file,
			path:      "deps",
			gitBranch: gitBranch,
			keyPrefix: "mix-deps",
		})
		buildResult(&results, buildResultOpts{
			file:      file,
			path:      "_build",
			gitBranch: gitBranch,
			keyPrefix: "mix-build",
		})
	case "requirements.txt":
		buildResult(&results, buildResultOpts{
			file:      file,
			path:      ".pip_cache",
			gitBranch: gitBranch,
			keyPrefix: "requirements",
		})
	case "composer.lock":
		buildResult(&results, buildResultOpts{
			file:      file,
			path:      "vendor",
			gitBranch: gitBranch,
			keyPrefix: "composer",
		})
	case "pom.xml":
		buildResult(&results, buildResultOpts{
			file:      file,
			path:      ".m2",
			gitBranch: gitBranch,
			keyPrefix: "maven",
		})
		buildResult(&results, buildResultOpts{
			file:      file,
			path:      "target",
			gitBranch: gitBranch,
			keyPrefix: "maven-target",
		})
	case "Podfile.lock":
		buildResult(&results, buildResultOpts{
			file:      file,
			path:      "Pods",
			gitBranch: gitBranch,
			keyPrefix: "pods",
		})
	case "go.sum":
		buildResult(&results, buildResultOpts{
			file:      file,
			path:      fmt.Sprintf("%s/go/pkg/mod", homedir),
			gitBranch: gitBranch,
			keyPrefix: "go",
		})
	}

	return results
}

type buildResultOpts struct {
	file      string
	path      string
	keyPrefix string
	gitBranch string
}

func buildResult(results *[]LookupResult, opts buildResultOpts) {
	checksum, err := generateChecksum(opts.file)
	if err != nil {
		fmt.Printf("Error generating checksum for %s: %v\n", opts.file, err)
	} else {
		key := fmt.Sprintf("%s-%s-%s", opts.keyPrefix, opts.gitBranch, checksum)
		*results = append(*results, LookupResult{
			DetectedFile: opts.file,
			Key:          normalizeKey(key),
			Path:         opts.path,
		})
	}
}

func normalizeKey(key string) string {
	return strings.ReplaceAll(key, "/", "-")
}
