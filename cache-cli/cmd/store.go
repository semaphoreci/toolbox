package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/files"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/utils"
	"github.com/spf13/cobra"
)

var storeCmd = &cobra.Command{
	Use:   "store [key path]",
	Short: "Store keys in the cache.",
	Long:  ``,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		RunStore(cmd, args)
	},
}

func RunStore(cmd *cobra.Command, args []string) {
	if len(args) != 0 && len(args) != 2 {
		fmt.Printf("Wrong number of arguments %d for store command\n", len(args))
		cmd.Help()
		return
	}

	storage, err := storage.InitStorage()
	utils.Check(err)

	if len(args) == 0 {
		lookupResults := files.Lookup(files.LookupOptions{
			GitBranch: os.Getenv("SEMAPHORE_GIT_BRANCH"),
			Restore:   false,
		})

		for _, lookupResult := range lookupResults {
			fmt.Printf("Detected %s.\n", lookupResult.DetectedFile)
			for _, entry := range lookupResult.Entries {
				key := entry.Keys[0]
				compressAndStore(storage, key, entry.Path)
			}
		}
	} else {
		compressAndStore(storage, args[0], args[1])
	}
}

func compressAndStore(storage storage.Storage, key, path string) {
	if _, err := os.Stat(path); err == nil {
		if ok, _ := storage.HasKey(key); ok {
			fmt.Printf("Key '%s' already exists.\n", key)
			return
		}

		compressingStart := time.Now()
		fmt.Printf("Compressing %s...\n", path)
		compressed, err := files.Compress(key, path)
		utils.Check(err)

		compressionDuration := time.Since(compressingStart)
		info, _ := os.Stat(compressed)
		fmt.Printf("Compression complete. Duration: %v. Size: %v bytes.\n", compressionDuration.String(), files.HumanReadableSize(info.Size()))

		uploadStart := time.Now()
		fmt.Printf("Uploading '%s' with cache key '%s'...\n", path, key)
		err = storage.Store(key, compressed)
		utils.Check(err)

		uploadDuration := time.Since(uploadStart)
		fmt.Printf("Upload complete. Duration: %v.\n", uploadDuration)
		os.Remove(compressed)
	} else {
		fmt.Printf("Path %s does not exist.\n", path)
	}
}

func init() {
	RootCmd.AddCommand(storeCmd)
}
