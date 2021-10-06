package cmd

import (
	"fmt"
	"os"

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
		fmt.Printf("Key '%s' exists in the cache store.\n", key)
	} else {
		fmt.Printf("Key '%s' doesn't exist in the cache store.\n", key)
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(hasKeyCmd)
}
