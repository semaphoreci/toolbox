package cmd

import (
	"os"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/utils"
	"github.com/spf13/cobra"
)

var isNotEmptyCmd = &cobra.Command{
	Use:   "is_not_empty",
	Short: "Check if the cache is not empty.",
	Long:  ``,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if RunIsNotEmpty(cmd, args) {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	},
}

func RunIsNotEmpty(cmd *cobra.Command, args []string) bool {
	storage, err := storage.InitStorage()
	utils.Check(err)

	isNotEmpty, err := storage.IsNotEmpty()
	utils.Check(err)

	return isNotEmpty
}

func init() {
	RootCmd.AddCommand(isNotEmptyCmd)
}
