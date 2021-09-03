package cmd

import (
	"fmt"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/utils"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all keys in the cache.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		RunList(cmd, args)
	},
}

func RunList(cmd *cobra.Command, args []string) {
	storage, err := storage.InitStorage()
	utils.Check(err)

	listResult, err := storage.List()
	utils.Check(err)

	if len(listResult.Keys) == 0 {
		fmt.Println("Cache is empty.")
	} else {
		fmt.Println(listResult)
	}
}

func init() {
	RootCmd.AddCommand(listCmd)
}
