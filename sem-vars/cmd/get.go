package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get key",
	Short: "Get a variable",
	Run:   RunPutCmd,
}

func RunGetCmd(cmd *cobra.Command, args []string) {
	fmt.Printf("getting stuff")
}

func init() {
	RootCmd.AddCommand(getCmd)
}
