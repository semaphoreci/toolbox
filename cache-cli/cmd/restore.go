package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/files"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/metrics"
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

	metricsManager, err := metrics.InitMetricsManager("local")
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
				downloadAndUnpack(storage, metricsManager, entry.Keys)
			}
		}
	} else {
		keys := strings.Split(args[0], ",")
		downloadAndUnpack(storage, metricsManager, keys)
	}
}

func downloadAndUnpack(storage storage.Storage, metricsManager metrics.MetricsManager, keys []string) {
	for _, rawKey := range keys {
		key := NormalizeKey(rawKey)
		if ok, _ := storage.HasKey(key); ok {
			fmt.Printf("HIT: '%s', using key '%s'.\n", key, key)
			downloadAndUnpackKey(storage, metricsManager, key)
			break
		}

		availableKeys, err := storage.List()
		utils.Check(err)

		matchingKey := findMatchingKey(availableKeys, key)
		if matchingKey != "" {
			fmt.Printf("HIT: '%s', using key '%s'.\n", key, matchingKey)
			downloadAndUnpackKey(storage, metricsManager, matchingKey)
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

func downloadAndUnpackKey(storage storage.Storage, metricsManager metrics.MetricsManager, key string) {
	downloadStart := time.Now()
	fmt.Printf("Downloading key '%s'...\n", key)
	compressed, err := storage.Restore(key)
	utils.Check(err)

	downloadDuration := time.Since(downloadStart)
	info, _ := os.Stat(compressed.Name())

	fmt.Printf("Download complete. Duration: %v. Size: %v bytes.\n", downloadDuration.String(), files.HumanReadableSize(info.Size()))
	publishMetrics(metricsManager, info, downloadDuration)

	unpackStart := time.Now()
	fmt.Printf("Unpacking '%s'...\n", compressed.Name())
	restorationPath, err := files.Unpack(metricsManager, compressed.Name())
	utils.Check(err)

	unpackDuration := time.Since(unpackStart)
	fmt.Printf("Unpack complete. Duration: %v.\n", unpackDuration)
	fmt.Printf("Restored: %s.\n", restorationPath)
	os.Remove(compressed.Name())
}

func publishMetrics(metricsManager metrics.MetricsManager, fileInfo fs.FileInfo, downloadDuration time.Duration) error {
	metricsToPublish := []metrics.Metric{
		{Name: "cache_download_size", Value: fmt.Sprintf("%d", fileInfo.Size())},
		{Name: "cache_download_time", Value: downloadDuration.String()},
	}

	username := os.Getenv("SEMAPHORE_CACHE_USERNAME")
	if username != "" {
		metricsToPublish = append(metricsToPublish, metrics.Metric{Name: "cache_user", Value: username})
	}

	cacheServerIP := getCacheServerIP()
	if cacheServerIP != "" {
		metricsToPublish = append(metricsToPublish, metrics.Metric{Name: "cache_server", Value: cacheServerIP})
	}

	return metricsManager.PublishBatch(metricsToPublish)
}

func getCacheServerIP() string {
	cacheURL := os.Getenv("SEMAPHORE_CACHE_URL")
	if cacheURL != "" {
		ipAndPort := strings.Split(cacheURL, ":")
		if len(ipAndPort) != 2 {
			return ""
		}

		return ipAndPort[0]
	}

	return ""
}

func init() {
	RootCmd.AddCommand(restoreCmd)
}
