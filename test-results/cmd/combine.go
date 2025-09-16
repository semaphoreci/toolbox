package cmd

import (
	"github.com/semaphoreci/toolbox/test-results/pkg/cli"
	"github.com/semaphoreci/toolbox/test-results/pkg/logger"
	"github.com/semaphoreci/toolbox/test-results/pkg/parser"
	"github.com/spf13/cobra"
)

// combineCmd represents the combine command
var combineCmd = &cobra.Command{
	Use:   "combine <json-file-path>... <json-file>]",
	Short: "combines multiples json summary files into one",
	Long:  `Combines multiples json summary files into one"`,
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputs := args[:len(args)-1]
		output := args[len(args)-1]

		err := cli.SetLogLevel(cmd)
		if err != nil {
			return err
		}

		skipCompression, err := cmd.Flags().GetBool("no-compress")
		if err != nil {
			return err
		}

		paths, err := cli.LoadFiles(inputs, ".json")
		if err != nil {
			return err
		}

		result := parser.NewResult()
		for _, path := range paths {
			inFile, err := cli.CheckFile(path)
			if err != nil {
				logger.Error(err.Error())
				return err
			}

			newResult, err := cli.Load(inFile)

			if err != nil {
				logger.Error(err.Error())
				return err
			}
			result.Combine(*newResult)
		}

		err = cli.DecorateResults(&result, cmd)
		if err != nil {
			logger.Error("Decorating results failed with error: %v", err)
			return err
		}

		jsonData, err := cli.Marshal(result)
		if err != nil {
			return err
		}

		_, err = cli.WriteToFilePath(jsonData, output, !skipCompression)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	combineCmd.Flags().Int32P("trim-output-to", "s", 0, "trim stdout to N characters, defaults to 0(unlimited)")
	combineCmd.Flags().BoolP("omit-output-for-passed", "o", false, "omit stdout if test passed, defaults to false")
	rootCmd.AddCommand(combineCmd)
}
