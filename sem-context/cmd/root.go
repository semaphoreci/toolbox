package cmd

import (
	"fmt"
	"os"

	"github.com/semaphoreci/toolbox/sem-context/pkg/flags"
	"github.com/semaphoreci/toolbox/sem-context/pkg/store"
	"github.com/spf13/cobra"
)

var IgnoreFailure bool

var Store store.Store

var RootCmd = &cobra.Command{
	Use:   "sem-context",
	Short: "Share variables between your Semaphore jobs",
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	Store = &store.ArtifactStore{}
	RootCmd.PersistentFlags().BoolVar(&flags.IgnoreFailure, "ignore-failure", false, "Ignore if failure occurs, and always return 0.")
}
