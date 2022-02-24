package cmd

import (
	"fmt"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/utils"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [key]",
	Short: "Delete a key from the cache.",
	Long:  ``,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		RunDelete(cmd, args)
	},
}

func RunDelete(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Printf("Incorrect number of arguments!\n")
		_ = cmd.Help()
		return
	}

	storage, err := storage.InitStorage()
	utils.Check(err)

	rawKey := args[0]
	key := NormalizeKey(rawKey)

	if ok, _ := storage.HasKey(key); ok {
		err := storage.Delete(key)
		utils.Check(err)
		fmt.Printf("Key '%s' is deleted.\n", key)
	} else {
		fmt.Printf("Key '%s' doesn't exist in the cache store.\n", key)
	}
}

func init() {
	RootCmd.AddCommand(deleteCmd)
}
