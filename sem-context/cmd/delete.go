package cmd

import (
	"github.com/semaphoreci/toolbox/sem-context/pkg/store"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [key]",
	Short: "Delete a variable",
	Run:   RunDeleteCmd,
}

func RunDeleteCmd(cmd *cobra.Command, args []string) {
	key := args[0]
	store.Delete(key)
}

func init() {
	RootCmd.AddCommand(deleteCmd)
}
