package cmd

import (
	"fmt"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/files"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/utils"
	"github.com/spf13/cobra"
)

var usageCmd = &cobra.Command{
	Use:   "usage",
	Short: "Get a summary of cache usage.",
	Long:  ``,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		RunUsage(cmd, args)
	},
}

func RunUsage(cmd *cobra.Command, args []string) {
	storage, err := storage.InitStorage()
	utils.Check(err)

	summary, err := storage.Usage()
	utils.Check(err)

	if summary.Free == -1 {
		fmt.Println("FREE SPACE: (unlimited)")
	} else {
		fmt.Printf("FREE SPACE: %s\n", files.HumanReadableSize(summary.Free))
	}

	fmt.Printf("USED SPACE: %s\n", files.HumanReadableSize(summary.Used))
}

func init() {
	RootCmd.AddCommand(usageCmd)
}
