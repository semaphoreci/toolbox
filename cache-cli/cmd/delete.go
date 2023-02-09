package cmd

import (
	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/utils"
	log "github.com/sirupsen/logrus"
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
		log.Errorf("Incorrect number of arguments!")
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
		log.Infof("Key '%s' is deleted.", key)
	} else {
		log.Infof("Key '%s' doesn't exist in the cache store.", key)
	}
}

func init() {
	RootCmd.AddCommand(deleteCmd)
}
