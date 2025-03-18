package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/archive"
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

	archiver := archive.NewArchiver(metricsManager)

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
				downloadAndUnpack(storage, archiver, metricsManager, entry.Keys)
			}
		}
	} else {
		keys := strings.Split(args[0], ",")
		downloadAndUnpack(storage, archiver, metricsManager, keys)
	}
}

func downloadAndUnpack(storage storage.Storage, archiver archive.Archiver, metricsManager metrics.MetricsManager, keys []string) {
	for _, rawKey := range keys {
		key := NormalizeKey(rawKey)
		if ok, _ := storage.HasKey(key); ok {
			log.Infof("HIT: '%s', using key '%s'.", key, key)
			downloadAndUnpackKey(storage, archiver, metricsManager, key)
			break
		}

		availableKeys, err := storage.List()
		utils.Check(err)

		matchingKey := findMatchingKey(availableKeys, key)
		if matchingKey != "" {
			log.Infof("HIT: '%s', using key '%s'.", key, matchingKey)
			downloadAndUnpackKey(storage, archiver, metricsManager, matchingKey)
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

func downloadAndUnpackKey(storage storage.Storage, archiver archive.Archiver, metricsManager metrics.MetricsManager, key string) {
	downloadStart := time.Now()
	log.Infof("Downloading key '%s'...", key)
	compressed, err := downloadKey(storage, key)
	utils.Check(err)

	downloadDuration := time.Since(downloadStart)
	info, _ := os.Stat(compressed.Name())

	log.Infof("Download complete. Duration: %v. Size: %v bytes.", downloadDuration.String(), files.HumanReadableSize(info.Size()))
	publishMetrics(metricsManager, info, downloadDuration)

	unpackStart := time.Now()
	log.Infof("Unpacking '%s'...", compressed.Name())
	restorationPath, err := archiver.Decompress(compressed.Name())
	utils.Check(err)

	unpackDuration := time.Since(unpackStart)
	log.Infof("Unpack complete. Duration: %v.", unpackDuration)
	log.Infof("Restored: %s.", restorationPath)

	err = os.Remove(compressed.Name())
	if err != nil {
		log.Errorf("Error removing %s: %v", compressed.Name(), err)
	}
}

func downloadKey(storage storage.Storage, key string) (*os.File, error) {
	backend := os.Getenv("SEMAPHORE_CACHE_BACKEND")

	// If this is not an sftp backend, then we are not in a cloud environment,
	// and in there, there's no CDN variation, so just use the storage.
	if backend != "sftp" {
		return storage.Restore(key)
	}

	// Here, we are using sftp, so we know we are in a cloud job.
	// But, not all cloud jobs should use this, so we only use it
	// if the SEMAPHORE_CACHE_CDN_* variables are defined
	cdnURL := os.Getenv("SEMAPHORE_CACHE_CDN_URL")
	cdnKey := os.Getenv("SEMAPHORE_CACHE_CDN_KEY")
	cdnSecret := os.Getenv("SEMAPHORE_CACHE_CDN_SECRET")
	if cdnURL == "" || cdnKey == "" || cdnSecret == "" {
		return storage.Restore(key)
	}

	log.Infof("Restoring using HTTP URL %s...", cdnURL)
	return files.DownloadFromHTTP(cdnURL, cdnKey, cdnSecret, key)
}

func publishMetrics(metricsManager metrics.MetricsManager, fileInfo fs.FileInfo, downloadDuration time.Duration) {
	metricsToPublish := []metrics.Metric{
		{Name: metrics.CacheDownloadSize, Value: fmt.Sprintf("%d", fileInfo.Size())},
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
