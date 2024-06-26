package cmd

import (
	"os"

	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/utils"
	log "github.com/sirupsen/logrus"
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
		log.Error("Incorrect number of arguments!")
		_ = cmd.Help()
		return true
	}

	storage, err := storage.InitStorage(cmd.Context())
	utils.Check(err)

	rawKey := args[0]
	key := NormalizeKey(rawKey)
	exists, err := storage.HasKey(cmd.Context(), key)
	utils.Check(err)

	if exists {
		log.Infof("Key '%s' exists in the cache store.", key)
		return true
	}

	log.Infof("Key '%s' doesn't exist in the cache store.", key)
	return false
}

func init() {
	RootCmd.AddCommand(hasKeyCmd)
}
