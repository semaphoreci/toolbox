package cmd

import (
	"strings"

	"github.com/semaphoreci/toolbox/sem-vars/pkg/store"
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
	store.Put(key, value)
}

func init() {
	RootCmd.AddCommand(putCmd)
}
