package cmd

import (
	"fmt"
	"os"

	"github.com/semaphoreci/toolbox/sem-context/pkg/flags"
	"github.com/spf13/cobra"
)

var IgnoreFailure bool

var RootCmd = &cobra.Command{
	Use:   "sem-context",
	Short: "Share variables between your Semaphore jobs",
}

func Execute() {
	RootCmd.PersistentFlags().BoolVar(&flags.IgnoreFailure, "ignore-failure", false, "Ignore if failure occures, and always return 0.")
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
