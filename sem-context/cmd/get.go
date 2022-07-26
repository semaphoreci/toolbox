package cmd

import (
	"fmt"

	"github.com/semaphoreci/toolbox/sem-context/pkg/flags"
	"github.com/semaphoreci/toolbox/sem-context/pkg/store"
	"github.com/semaphoreci/toolbox/sem-context/pkg/utils"
	"github.com/semaphoreci/toolbox/sem-context/pkg/validators"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a variable",
	Run:   RunGetCmd,
}

func RunGetCmd(cmd *cobra.Command, args []string) {
	utils.CheckError(validators.ValidateGetAndDeleteArguments(args))
	value, err := store.Get(args[0])
	if err != nil && err.(*utils.Error).ExitCode == 1 && flags.Fallback != "" {
		fmt.Println(flags.Fallback)
		return
	}
	utils.CheckError(err)
	fmt.Println(value)
}

func init() {
	getCmd.Flags().StringVar(&flags.Fallback, "fallback", "", "Default value to be returned if key does not exist.")
	RootCmd.AddCommand(getCmd)
}
