package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/files"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/utils"
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore [keys]",
	Short: "Restore keys from the cache.",
	Long:  ``,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		RunRestore(cmd, args)
	},
}

func RunRestore(cmd *cobra.Command, args []string) {
	if len(args) > 1 {
		fmt.Printf("Wrong number of arguments %d for restore command\n", len(args))
		cmd.Help()
		return
	}

	storage, err := storage.InitStorage()
	utils.Check(err)

	if len(args) == 0 {
		lookupResults := files.Lookup(files.LookupOptions{
			GitBranch: os.Getenv("SEMAPHORE_GIT_BRANCH"),
			Restore:   true,
		})

		if len(lookupResults) == 0 {
			fmt.Printf("Nothing to restore from cache.")
			return
		}

		for _, lookupResult := range lookupResults {
			fmt.Printf("Detected %s.\n", lookupResult.DetectedFile)
			for _, entry := range lookupResult.Entries {
				fmt.Printf("Fetching '%s' directory with cache keys '%s'...\n", entry.Path, strings.Join(entry.Keys, ","))
				downloadAndUnpack(storage, entry.Keys)
			}
		}
	} else {
		keys := strings.Split(args[0], ",")
		downloadAndUnpack(storage, keys)
	}
}

func downloadAndUnpack(storage storage.Storage, keys []string) {
	for _, rawKey := range keys {
		key := NormalizeKey(rawKey)
		if ok, _ := storage.HasKey(key); ok {
			fmt.Printf("HIT: '%s', using key '%s'.\n", key, key)
			downloadAndUnpackKey(storage, key)
			break
		}

		availableKeys, err := storage.List()
		utils.Check(err)

		matchingKey := findMatchingKey(availableKeys, key)
		if matchingKey != "" {
			fmt.Printf("HIT: '%s', using key '%s'.\n", key, matchingKey)
			downloadAndUnpackKey(storage, matchingKey)
			break
		} else {
			fmt.Printf("MISS: '%s'.\n", key)
		}
	}
}

func findMatchingKey(availableKeys []storage.CacheKey, match string) string {
	for _, availableKey := range availableKeys {
		if strings.Contains(availableKey.Name, match) {
			return availableKey.Name
		}
	}

	return ""
}

func downloadAndUnpackKey(storage storage.Storage, key string) {
	downloadStart := time.Now()
	fmt.Printf("Downloading key '%s'...\n", key)
	compressed, err := storage.Restore(key)
	utils.Check(err)

	downloadDuration := time.Since(downloadStart)
	info, _ := os.Stat(compressed.Name())
	fmt.Printf("Download complete. Duration: %v. Size: %v bytes.\n", downloadDuration.String(), files.HumanReadableSize(info.Size()))

	unpackStart := time.Now()
	fmt.Printf("Unpacking '%s'...\n", compressed.Name())
	restorationPath, err := files.Unpack(compressed.Name())
	utils.Check(err)

	unpackDuration := time.Since(unpackStart)
	fmt.Printf("Unpack complete. Duration: %v.\n", unpackDuration)
	fmt.Printf("Restored: %s.\n", restorationPath)
	os.Remove(compressed.Name())
}

func init() {
	RootCmd.AddCommand(restoreCmd)
}
