package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/semaphoreci/toolbox/test-results/pkg/cli"
	"github.com/semaphoreci/toolbox/test-results/pkg/logger"
	"github.com/semaphoreci/toolbox/test-results/pkg/parsers"
	"github.com/spf13/cobra"
)

func formatCompileDescription() string {
	var description strings.Builder
	description.WriteString(`Parses test result files to well defined json schema

It traverses through directory structure specified by <file-path> and compiles
test result files (XML, JSON, etc.) based on the detected or specified parser.

You can specify parsers for individual files using the syntax:
  file.xml:parser-name

Examples:
  test-results compile results.xml output.json
  test-results compile results.xml:golang lint.json:go:staticcheck output.json
  test-results compile --ignore-missing test1.xml test2.xml output.json

Available parsers:
`)

	for _, parser := range parsers.GetAvailableParsers() {
		description.WriteString(fmt.Sprintf("  %-15s - %s\n", parser.Name, parser.Description))
	}

	description.WriteString(`
Use --parser flag to specify a default parser for all files, or "auto" for automatic detection.
Use --ignore-missing to skip files that don't exist instead of failing.`)

	return description.String()
}

// compileCmd represents the compile command
var compileCmd = &cobra.Command{
	Use:   "compile <file-path>... <json-file>",
	Short: "parses test result files to well defined json schema",
	Long:  formatCompileDescription(),
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

		ignoreMissing, err := cmd.Flags().GetBool("ignore-missing")
		if err != nil {
			return err
		}

		fileParsers := cli.ParseFileArgs(inputs)

		supportedExts := parsers.GetSupportedExtensions()
		extMap := make(map[string]bool)
		for _, ext := range supportedExts {
			extMap[ext] = true
		}

		var allPairs []cli.FileParserPair
		for _, pair := range fileParsers {
			file, err := os.Stat(pair.Path)
			if err != nil {
				if os.IsNotExist(err) && ignoreMissing {
					logger.Warn("File not found, skipping: %s", pair.Path)
					continue
				}
				return fmt.Errorf("failed to stat %s: %v", pair.Path, err)
			}

			if file.IsDir() {
				err := filepath.WalkDir(pair.Path, func(path string, d os.DirEntry, err error) error {
					if err != nil {
						return err
					}
					if d.Type().IsRegular() {
						ext := filepath.Ext(d.Name())
						if extMap[ext] {
							allPairs = append(allPairs, cli.FileParserPair{Path: path, Parser: ""})
						}
					}
					return nil
				})
				if err != nil {
					return err
				}
			} else {
				allPairs = append(allPairs, pair)
			}
		}

		if len(allPairs) == 0 {
			logger.Warn("No files to process")
			result := cli.EmptyResult()
			jsonData, err := json.Marshal(result)
			if err != nil {
				return err
			}
			_, err = cli.WriteToFilePath(jsonData, output, !skipCompression)
			return err
		}

		dirPath, err := os.MkdirTemp("", "test-results-*")
		if err != nil {
			return err
		}
		defer os.RemoveAll(dirPath)

		for _, pair := range allPairs {
			parser, err := cli.FindParserForFile(pair, cmd)
			if err != nil {
				if ignoreMissing {
					logger.Warn("Could not find parser for %s, skipping: %v", pair.Path, err)
					continue
				}
				return err
			}

			testResults, err := cli.Parse(parser, pair.Path, cmd)
			if err != nil {
				if ignoreMissing {
					logger.Warn("Failed to parse %s, skipping: %v", pair.Path, err)
					continue
				}
				return err
			}

			jsonData, err := cli.Marshal(testResults)
			if err != nil {
				return err
			}

			tmpFile, err := os.CreateTemp(dirPath, "result-*.json")
			if err != nil {
				return err
			}

			_, err = cli.WriteToFilePath(jsonData, tmpFile.Name(), !skipCompression)
			if err != nil {
				return err
			}
		}

		result, err := cli.MergeFiles(dirPath, cmd)
		if err != nil {
			return err
		}

		jsonData, err := json.Marshal(result)
		if err != nil {
			logger.Error("Marshaling results failed with: %v", err)
			return err
		}

		_, err = cli.WriteToFilePath(jsonData, output, !skipCompression)
		if err != nil {
			return err
		}

		logger.Info("Compiled test results saved to %s", output)
		return nil
	},
}

func init() {
	compileCmd.Flags().Int32P("trim-output-to", "s", 0, "trim stdout to N characters, defaults to 0(unlimited)")
	compileCmd.Flags().BoolP("omit-output-for-passed", "o", false, "omit stdout if test passed, defaults to false")
	compileCmd.Flags().Bool("ignore-missing", false, "ignore missing files instead of failing")
	rootCmd.AddCommand(compileCmd)
}
