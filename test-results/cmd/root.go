package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// Flags

var cfgFile string

var versionString = "0.9.0"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "test-results",
	Version: versionString,
	Short:   fmt.Sprintf("Semaphore 2.0 Test results CLI v%s", versionString),
	Long:    fmt.Sprintf("Semaphore 2.0 Test results CLI v%s", versionString),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.test-results.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolP("trace", "", false, "trace output")
	rootCmd.PersistentFlags().StringP("name", "N", "", "name of the suite")
	rootCmd.PersistentFlags().StringP("suite-prefix", "S", "", "prefix for each suite")
	rootCmd.PersistentFlags().StringP("parser", "p", "auto", "override parser to be used")
	rootCmd.PersistentFlags().Bool("no-compress", false, "skip gzip compression for the output")
	rootCmd.PersistentFlags().IntP("trim-output-to", "s", 1000, "trim stdout/stderr to last N characters, defaults to 1000 (use 0 or --no-trim-output to disable)")
	rootCmd.PersistentFlags().Bool("no-trim-output", false, "disable output trimming entirely")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".test-results" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".test-results")
	}

	viper.AutomaticEnv() // read in environment variables that match

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
