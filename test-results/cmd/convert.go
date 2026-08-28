package cmd

import (
	"fmt"
	"os"

	"github.com/semaphoreci/toolbox/test-results/pkg/cli"
	"github.com/semaphoreci/toolbox/test-results/pkg/parser"
	"github.com/spf13/cobra"
)

var convertCmd = &cobra.Command{
	Use:   "convert <file-path>[:parser] <jsonl-file>",
	Short: "parses a test result file and writes it as a semaphore/test-report JSON Lines stream",
	Long: `Parses a test result file (JUnit XML, go test -json, …) with the detected
or specified parser and writes the result as a semaphore/test-report JSON Lines
stream — the same format the reporter libraries emit natively.

Examples:
  test-results convert report.xml report.semaphore.jsonl
  test-results convert events.jsonl:go-test-json report.semaphore.jsonl`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		err := cli.SetLogLevel(cmd)
		if err != nil {
			return err
		}

		pair := cli.ParseFileArgs(args[:1])[0]
		foundParser, err := cli.FindParserForFile(pair, cmd)
		if err != nil {
			return err
		}

		result, err := cli.Parse(foundParser, pair.Path, cmd)
		if err != nil {
			return err
		}

		if len(result.TestResults) == 0 {
			return fmt.Errorf("no test results parsed from %s", pair.Path)
		}

		payload, err := parser.WriteJSONL(&result.TestResults[0], "test-results convert")
		if err != nil {
			return err
		}

		if err := os.WriteFile(args[1], payload, 0o600); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Converted %s to %s\n", pair.Path, args[1])
		return nil
	},
}

func init() {
	convertCmd.Flags().BoolP("omit-output-for-passed", "o", false, "omit stdout if test passed, defaults to false")
	rootCmd.AddCommand(convertCmd)
}
