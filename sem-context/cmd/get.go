package cmd

import (
	"fmt"

	"github.com/semaphoreci/toolbox/sem-context/pkg/flags"
	"github.com/semaphoreci/toolbox/sem-context/pkg/store"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a variable",
	Run:   RunGetCmd,
}

func RunGetCmd(cmd *cobra.Command, args []string) {
	key := args[0]
	fmt.Println(store.Get(key))
}

func init() {
	getCmd.Flags().StringVar(&flags.Fallback, "fallback", "", "Default value to be returned if key does not exist.")
	RootCmd.AddCommand(getCmd)
}
