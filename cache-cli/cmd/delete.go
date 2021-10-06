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
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		RunDelete(cmd, args)
	},
}

func RunDelete(cmd *cobra.Command, args []string) {
	storage, err := storage.InitStorage()
	utils.Check(err)

	key := args[0]
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
