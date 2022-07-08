package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var putCmd = &cobra.Command{
	Use:   "put key=value",
	Short: "Stores a variable",
	Run:   RunPutCmd,
}

func RunPutCmd(cmd *cobra.Command, args []string) {
	fmt.Printf("putting stuff")
}

func init() {
	RootCmd.AddCommand(putCmd)
}
