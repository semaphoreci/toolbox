package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/files"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/utils"
	log "github.com/sirupsen/logrus"
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

	description := fmt.Sprintf(
		`Sort keys by a specific field. Possible values are: %v.`,
		strings.Join(storage.ValidSortByKeys, ","),
	)

	cmd.Flags().StringP("sort-by", "s", storage.SortByStoreTime, description)
	return cmd
}

func RunList(cmd *cobra.Command, args []string) {
	sortBy, err := cmd.Flags().GetString("sort-by")
	utils.Check(err)

	storage, err := storage.InitStorageWithConfig(storage.StorageConfig{SortKeysBy: sortBy})
	utils.Check(err)

	keys, err := storage.List()
	utils.Check(err)

	if len(keys) == 0 {
		log.Info("Cache is empty.")
	} else {
		log.Info(formatList(keys))
	}
}

func formatList(keys []storage.CacheKey) string {
	formatted := fmt.Sprintf("%-60s %-12s %-22s %-22s\n", "NAME", "SIZE", "STORED AT", "ACCESSED AT")
	for _, key := range keys {
		formatted += fmt.Sprintf(
			"%-60s %-12s %-22s %-22s\n",
			key.Name,
			files.HumanReadableSize(key.Size),
			key.StoredAt.Format(time.RFC822),
			key.LastAccessedAt.Format(time.RFC822),
		)
	}

	return formatted
}

func init() {
	RootCmd.AddCommand(NewListCommand())
}
