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
		lookupResults := files.Lookup()
		downloadAndDecompress(storage, onlyKeys(lookupResults))
	} else {
		keys := strings.Split(args[0], ",")
		downloadAndDecompress(storage, keys)
	}
}

func onlyKeys(lookupResults []files.LookupResult) []string {
	keys := []string{}
	for _, result := range lookupResults {
		keys = append(keys, result.Key)
	}

	return keys
}

func downloadAndDecompress(storage storage.Storage, keys []string) {
	for _, key := range keys {
		if ok, _ := storage.HasKey(key); !ok {
			fmt.Printf("Key %s does not exist.\n", key)
			continue
		}

		downloadStart := time.Now()
		fmt.Printf("Downloading %s...\n", key)
		compressed, err := storage.Restore(key)
		utils.Check(err)

		downloadDuration := time.Since(downloadStart)
		info, _ := os.Stat(compressed.Name())
		fmt.Printf("Download complete. Duration: %v. Size: %v bytes.\n", downloadDuration.String(), files.HumanReadableSize(info.Size()))

		decompressStart := time.Now()
		fmt.Printf("Decompressing '%s'...\n", compressed.Name())
		err = files.Decompress(compressed.Name())
		utils.Check(err)

		decompressDuration := time.Since(decompressStart)
		fmt.Printf("Decompression complete. Duration: %v.\n", decompressDuration)
		os.Remove(compressed.Name())
	}
}

func init() {
	RootCmd.AddCommand(restoreCmd)
}
