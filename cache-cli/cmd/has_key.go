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
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if !RunHasKey(cmd, args) {
			os.Exit(1)
		}
	},
}

func RunHasKey(cmd *cobra.Command, args []string) bool {
	if len(args) != 1 {
		fmt.Printf("Incorrect number of arguments!\n")
		_ = cmd.Help()
		return true
	}

	storage, err := storage.InitStorage()
	utils.Check(err)

	rawKey := args[0]
	key := NormalizeKey(rawKey)
	exists, err := storage.HasKey(key)
	utils.Check(err)

	if exists {
		fmt.Printf("Key '%s' exists in the cache store.\n", key)
		return true
	}

	fmt.Printf("Key '%s' doesn't exist in the cache store.\n", key)
	return false
}

func init() {
	RootCmd.AddCommand(hasKeyCmd)
}
