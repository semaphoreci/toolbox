package cmd

import (
	"fmt"
	"time"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/files"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/utils"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all keys in the cache.",
	Long:  ``,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		RunList(cmd, args)
	},
}

func RunList(cmd *cobra.Command, args []string) {
	storage, err := storage.InitStorage()
	utils.Check(err)

	keys, err := storage.List()
	utils.Check(err)

	if len(keys) == 0 {
		fmt.Println("Cache is empty.")
	} else {
		fmt.Printf("%-60s %-12s %-12s\n", "NAME", "SIZE", "STORED AT")
		for _, key := range keys {
			fmt.Printf("%-60s %-12s %-12s\n", key.Name, files.HumanReadableSize(key.Size), key.StoredAt.Format(time.RFC822))
		}
	}
}

func init() {
	RootCmd.AddCommand(listCmd)
}
