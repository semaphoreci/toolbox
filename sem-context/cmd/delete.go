package cmd

import (
	"github.com/semaphoreci/toolbox/sem-context/pkg/store"
	"github.com/semaphoreci/toolbox/sem-context/pkg/utils"
	"github.com/semaphoreci/toolbox/sem-context/pkg/validators"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [key]",
	Short: "Delete a variable",
	Run:   RunDeleteCmd,
}

func RunDeleteCmd(cmd *cobra.Command, args []string) {
	utils.CheckError(validators.ValidateGetAndDeleteArguments(args))
	key := args[0]
	err := store.Delete(key)
	utils.CheckError(err)
	//TODO some feedback log
}

func init() {
	RootCmd.AddCommand(deleteCmd)
}
