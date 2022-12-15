package cmd

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/files"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/metrics"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/utils"
	log "github.com/sirupsen/logrus"
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
		log.Error("Incorrect number of arguments!")
		_ = cmd.Help()
		return
	}

	storage, err := storage.InitStorage()
	utils.Check(err)

	metricsManager, err := metrics.InitMetricsManager(metrics.LocalBackend)
	utils.Check(err)

	if len(args) == 0 {
		lookupResults := files.Lookup(files.LookupOptions{
			GitBranch: FindGitBranch(),
			Restore:   true,
		})

		if len(lookupResults) == 0 {
			log.Info("Nothing to restore from cache.")
			return
		}

		for _, lookupResult := range lookupResults {
			log.Infof("Detected %s.", lookupResult.DetectedFile)
			for _, entry := range lookupResult.Entries {
				log.Infof("Fetching '%s' directory with cache keys '%s'...", entry.Path, strings.Join(entry.Keys, ","))
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
			log.Infof("HIT: '%s', using key '%s'.\n", key, key)
			downloadAndUnpackKey(storage, metricsManager, key)
			break
		}

		availableKeys, err := storage.List()
		utils.Check(err)

		matchingKey := findMatchingKey(availableKeys, key)
		if matchingKey != "" {
			log.Infof("HIT: '%s', using key '%s'.", key, matchingKey)
			downloadAndUnpackKey(storage, metricsManager, matchingKey)
			break
		} else {
			log.Infof("MISS: '%s'.", key)
		}
	}
}

func findMatchingKey(availableKeys []storage.CacheKey, match string) string {
	for _, availableKey := range availableKeys {
		isMatch, _ := regexp.MatchString(match, availableKey.Name)
		if isMatch {
			return availableKey.Name
		}
	}

	return ""
}

func downloadAndUnpackKey(storage storage.Storage, metricsManager metrics.MetricsManager, key string) {
	start := time.Now()

	reader, writer := io.Pipe()
	var wg sync.WaitGroup
	wg.Add(2)

	// Download
	go func() {
		written, err := storage.Restore(key, writer)
		log.Infof("Download complete. Size: %v bytes.", files.HumanReadableSize(written))
		wg.Done()
		utils.Check(err)
	}()

	// Decompress
	go func() {
		err := files.Unpack(metricsManager, reader)
		wg.Done()
		utils.Check(err)
	}()

	log.Infof("Downloading key '%s'...", key)
	wg.Wait()

	duration := time.Since(start)
	log.Infof("Download complete. Duration: %v", duration.String())

	// publishMetrics(metricsManager, written, duration)
	// TODO: do I need to close the pipe?
}

func publishMetrics(metricsManager metrics.MetricsManager, downloadSize int64, downloadDuration time.Duration) {
	metricsToPublish := []metrics.Metric{
		{Name: metrics.CacheDownloadSize, Value: fmt.Sprintf("%d", downloadSize)},
		{Name: metrics.CacheDownloadTime, Value: downloadDuration.String()},
	}

	username := os.Getenv("SEMAPHORE_CACHE_USERNAME")
	if username != "" {
		metricsToPublish = append(metricsToPublish, metrics.Metric{Name: metrics.CacheUser, Value: username})
	}

	cacheServerIP := getCacheServerIP()
	if cacheServerIP != "" {
		metricsToPublish = append(metricsToPublish, metrics.Metric{Name: metrics.CacheServer, Value: cacheServerIP})
	}

	metricsToPublish = append(metricsToPublish, metrics.Metric{Name: metrics.CacheTotalRate, Value: "1"})

	err := metricsManager.PublishBatch(metricsToPublish)
	if err != nil {
		log.Errorf("Error publishing metrics: %v", err)
	}
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
