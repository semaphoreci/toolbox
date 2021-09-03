package cmd

import (
	"fmt"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/utils"
	"github.com/spf13/cobra"
)

var hasKeyCmd = &cobra.Command{
	Use:   "has_key [key]",
	Short: "Check if a key is present in the cache.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		RunHasKey(cmd, args)
	},
}

func RunHasKey(cmd *cobra.Command, args []string) {
	storage, err := storage.InitStorage()
	utils.Check(err)

	key := args[0]
	exists, err := storage.HasKey(key)
	utils.Check(err)

	if exists {
		fmt.Printf("The key %s exists in the cache.", key)
	} else {
		fmt.Printf("The key %s does not exist in the cache.", key)
	}
}

func init() {
	RootCmd.AddCommand(hasKeyCmd)
}
