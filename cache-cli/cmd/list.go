package cmd

import (
	"fmt"
	"time"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/files"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/utils"
	"github.com/spf13/cobra"
)

func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all keys in the cache.",
		Long:  ``,
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			RunList(cmd, args)
		},
	}

	cmd.Flags().Bool("sort-by-access-time", false, "Sort keys by access time instead of creation time.")
	return cmd
}

func RunList(cmd *cobra.Command, args []string) {
	sortByAccessTime, err := cmd.Flags().GetBool("sort-by-access-time")
	utils.Check(err)

	storage, err := storage.InitStorageWithConfig(storage.StorageConfig{SortKeysByAccessTime: sortByAccessTime})
	utils.Check(err)

	keys, err := storage.List()
	utils.Check(err)

	if len(keys) == 0 {
		fmt.Println("Cache is empty.")
	} else {
		fmt.Printf("%-60s %-12s %-22s %-22s\n", "NAME", "SIZE", "STORED AT", "ACCESSED AT")
		for _, key := range keys {
			fmt.Printf(
				"%-60s %-12s %-22s %-22s\n",
				key.Name,
				files.HumanReadableSize(key.Size),
				key.StoredAt.Format(time.RFC822),
				key.LastAccessedAt.Format(time.RFC822),
			)
		}
	}
}

func init() {
	RootCmd.AddCommand(NewListCommand())
}
