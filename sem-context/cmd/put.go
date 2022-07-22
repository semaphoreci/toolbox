package cmd

import (
	"strings"

	"github.com/semaphoreci/toolbox/sem-context/pkg/flags"
	"github.com/semaphoreci/toolbox/sem-context/pkg/store"
	"github.com/semaphoreci/toolbox/sem-context/pkg/utils"
	"github.com/semaphoreci/toolbox/sem-context/pkg/validators"
	"github.com/spf13/cobra"
)

var putCmd = &cobra.Command{
	Use:   "put key=value",
	Short: "Stores a variable",
	Run:   RunPutCmd,
}

func RunPutCmd(cmd *cobra.Command, args []string) {
	argument := args[0]
	key_value := strings.Split(argument, "=")
	key, value := key_value[0], key_value[1]
	utils.CheckError(validators.IsKeyValid(key), 3)
	utils.CheckError(validators.IsValueValid(key), 4)
	store.Put(key, value)
}

func init() {
	putCmd.Flags().BoolVarP(&flags.Force, "force", "f", false, "If same key already exists, overwrite it.")
	RootCmd.AddCommand(putCmd)
}
