package cmd

import (
	"github.com/semaphoreci/toolbox/cache-cli/pkg/files"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var usageCmd = &cobra.Command{
	Use:   "usage",
	Short: "Get a summary of cache usage.",
	Long:  ``,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		RunUsage(cmd, args)
	},
}

func RunUsage(cmd *cobra.Command, args []string) {
	storage, err := storage.InitStorage(cmd.Context())
	utils.Check(err)

	summary, err := storage.Usage(cmd.Context())
	utils.Check(err)

	if summary.Free == -1 {
		log.Info("FREE SPACE: (unlimited)")
	} else {
		log.Infof("FREE SPACE: %s", files.HumanReadableSize(summary.Free))
	}

	log.Infof("USED SPACE: %s", files.HumanReadableSize(summary.Used))
}

func init() {
	RootCmd.AddCommand(usageCmd)
}
