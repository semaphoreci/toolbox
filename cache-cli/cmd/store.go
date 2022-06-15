package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/files"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/utils"
	"github.com/spf13/cobra"
)

func NewStoreCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "store [key path]",
		Short: "Store keys in the cache.",
		Long:  ``,
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			RunStore(cmd, args)
		},
	}

	description := fmt.Sprintf(`
		If storage does not have enough space,
		keys will be sorted in descending order and will be cleaned up starting from the last one.
		Possible values are: %v.
	`, strings.Join(storage.ValidSortByKeys, ","))

	cmd.Flags().StringP("cleanup-by", "c", storage.SortByStoreTime, description)
	return cmd
}

func RunStore(cmd *cobra.Command, args []string) {
	if len(args) != 0 && len(args) != 2 {
		fmt.Printf("Incorrect number of arguments!\n")
		_ = cmd.Help()
		return
	}

	cleanupBy, err := cmd.Flags().GetString("cleanup-by")
	utils.Check(err)

	storage, err := storage.InitStorageWithConfig(storage.StorageConfig{SortKeysBy: cleanupBy})
	utils.Check(err)

	if len(args) == 0 {
		lookupResults := files.Lookup(files.LookupOptions{
			GitBranch: FindGitBranch(),
			Restore:   false,
		})

		if len(lookupResults) == 0 {
			fmt.Printf("Nothing to store in cache.\n")
			return
		}

		for _, lookupResult := range lookupResults {
			fmt.Printf("Detected %s.\n", lookupResult.DetectedFile)
			for _, entry := range lookupResult.Entries {
				fmt.Printf("Using default cache path '%s'.\n", entry.Path)
				key := entry.Keys[0]
				compressAndStore(storage, key, entry.Path)
			}
		}
	} else {
		path := filepath.FromSlash(args[1])
		compressAndStore(storage, args[0], path)
	}
}

func compressAndStore(storage storage.Storage, rawKey, path string) {
	key := NormalizeKey(rawKey)
	if _, err := os.Stat(path); err == nil {
		if ok, _ := storage.HasKey(key); ok {
			fmt.Printf("Key '%s' already exists.\n", key)
			return
		}

		compressedFilePath, compressedFileSize, err := compress(key, path)
		if err != nil {
			fmt.Printf("Error compressing %s: %v\n", path, err)
			return
		}

		maxSpace := storage.Config().MaxSpace
		if compressedFileSize > maxSpace {
			fmt.Printf("Archive exceeds allocated %s for cache.\n", files.HumanReadableSize(maxSpace))
			return
		}

		uploadStart := time.Now()
		fmt.Printf("Uploading '%s' with cache key '%s'...\n", path, key)
		err = storage.Store(key, compressedFilePath)
		utils.Check(err)

		uploadDuration := time.Since(uploadStart)
		fmt.Printf("Upload complete. Duration: %v.\n", uploadDuration)

		err = os.Remove(compressedFilePath)
		if err != nil {
			fmt.Printf("Error removing %s: %v", compressedFilePath, err)
		}
	} else {
		fmt.Printf("'%s' doesn't exist locally.\n", path)
	}
}

func compress(key, path string) (string, int64, error) {
	compressingStart := time.Now()
	fmt.Printf("Compressing %s...\n", path)
	compressed, err := files.Compress(key, path)
	utils.Check(err)

	compressionDuration := time.Since(compressingStart)
	info, err := os.Stat(compressed)
	if err != nil {
		_ = os.Remove(compressed)
		return "", -1, err
	}

	fmt.Printf("Compression complete. Duration: %v. Size: %v bytes.\n", compressionDuration.String(), files.HumanReadableSize(info.Size()))
	return compressed, info.Size(), nil
}

func NormalizeKey(key string) string {
	normalizedKey := strings.ReplaceAll(key, "/", "-")
	if normalizedKey != key {
		fmt.Printf("Key '%s' is normalized to '%s'.\n", key, normalizedKey)
	}

	return normalizedKey
}

func FindGitBranch() string {
	gitPrBranch := os.Getenv("SEMAPHORE_GIT_PR_BRANCH")
	if gitPrBranch != "" {
		return gitPrBranch
	}

	return os.Getenv("SEMAPHORE_GIT_BRANCH")
}

func init() {
	RootCmd.AddCommand(NewStoreCommand())
}
