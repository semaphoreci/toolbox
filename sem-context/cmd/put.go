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
	utils.CheckError(validators.ValidatePutArguments(args))
	key_value := strings.Split(args[0], "=")
	key, value := key_value[0], key_value[1]
	err := store.Put(key, value)
	utils.CheckError(err)
	//TODO Maybe some feedback log
}

func init() {
	putCmd.Flags().BoolVarP(&flags.Force, "force", "f", false, "If same key already exists, overwrite it.")
	RootCmd.AddCommand(putCmd)
}
