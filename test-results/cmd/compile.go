package cmd

/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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

Available parsers:
`)
	
	for _, parser := range parsers.GetAvailableParsers() {
		description.WriteString(fmt.Sprintf("  %-15s - %s\n", parser.Name, parser.Description))
	}
	
	description.WriteString(`
Use --parser flag to specify a parser, or "auto" for automatic detection.`)
	
	return description.String()
}

// compileCmd represents the compile command
var compileCmd = &cobra.Command{
	Use:   "compile <file-path>... <json-file>]",
	Short: "parses test result files to well defined json schema",
	Long: formatCompileDescription(),
	Args: cobra.MinimumNArgs(2),
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

		// Get supported extensions from all parsers
		supportedExts := parsers.GetSupportedExtensions()
		extMap := make(map[string]bool)
		for _, ext := range supportedExts {
			extMap[ext] = true
		}
		
		// Load all files with supported extensions
		paths := []string{}
		for _, input := range inputs {
			file, err := os.Stat(input)
			if err != nil {
				return err
			}
			
			if file.IsDir() {
				// Walk directory and get all files with supported extensions
				err := filepath.WalkDir(input, func(path string, d os.DirEntry, err error) error {
					if d.Type().IsRegular() {
						ext := filepath.Ext(d.Name())
						if extMap[ext] {
							paths = append(paths, path)
						}
					}
					return nil
				})
				if err != nil {
					return err
				}
			} else {
				// Single file - always include it (parser will validate)
				paths = append(paths, input)
			}
		}

		dirPath, err := os.MkdirTemp("", "test-results-*")

		if err != nil {
			return err
		}

		defer os.RemoveAll(dirPath)

		for _, path := range paths {
			parser, err := cli.FindParser(path, cmd)
			if err != nil {
				return err
			}

			testResults, err := cli.Parse(parser, path, cmd)
			if err != nil {
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

		return nil
	},
}

func init() {
	compileCmd.Flags().Int32P("trim-output-to", "s", 0, "trim stdout to N characters, defaults to 0(unlimited)")
	compileCmd.Flags().BoolP("omit-output-for-passed", "o", false, "omit stdout if test passed, defaults to false")
	rootCmd.AddCommand(compileCmd)
}
