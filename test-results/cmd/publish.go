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
	"path"
	"path/filepath"
	"strings"

	"github.com/semaphoreci/toolbox/test-results/pkg/cli"
	"github.com/semaphoreci/toolbox/test-results/pkg/logger"
	"github.com/semaphoreci/toolbox/test-results/pkg/parser"
	"github.com/semaphoreci/toolbox/test-results/pkg/parsers"
	"github.com/spf13/cobra"
)

func formatPublishDescription() string {
	var description strings.Builder
	description.WriteString(`Parses test result files to well defined json schema and publishes results to artifacts storage

It traverses through directory structure specified by <file-path>, compiles
test result files (XML, JSON, etc.) and publishes them as one artifact.

Available parsers:
`)
	
	for _, parser := range parsers.GetAvailableParsers() {
		description.WriteString(fmt.Sprintf("  %-15s - %s\n", parser.Name, parser.Description))
	}
	
	description.WriteString(`
Use --parser flag to specify a parser, or "auto" for automatic detection.`)
	
	return description.String()
}

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish <file-path>...",
	Short: "parses test result files to well defined json schema and publishes results to artifacts storage",
	Long:  formatPublishDescription(),
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputs := args
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

		pushStats := &cli.ArtifactStats{}

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

			_, err = cli.WriteToFile(jsonData, tmpFile, !skipCompression)
			if err != nil {
				return err
			}

			err = tmpFile.Close()
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

		fileName, err := cli.WriteToTmpFile(jsonData, !skipCompression)
		if err != nil {
			return err
		}

		defer os.Remove(fileName)

		_, stats, err := cli.PushArtifacts("job", fileName, path.Join("test-results", "junit.json"), cmd)
		if err != nil {
			return err
		}
		if stats != nil {
			pushStats.Operations++
			pushStats.FileCount += stats.FileCount
			pushStats.TotalSize += stats.TotalSize
		}

		if err = pushSummaryWithStats(result.TestResults, "job", path.Join("test-results", "summary.json"), cmd, pushStats); err != nil {
			return err
		}

		pipelineID, found := os.LookupEnv("SEMAPHORE_PIPELINE_ID")
		if !found {
			logger.Error("SEMAPHORE_PIPELINE_ID env is missing")
			return err
		}

		jobID, found := os.LookupEnv("SEMAPHORE_JOB_ID")
		if !found {
			logger.Error("SEMAPHORE_JOB_ID env is missing")
			return err
		}

		_, stats, err = cli.PushArtifacts("workflow", fileName, path.Join("test-results", pipelineID, jobID+".json"), cmd)
		if err != nil {
			return err
		}
		if stats != nil {
			pushStats.Operations++
			pushStats.FileCount += stats.FileCount
			pushStats.TotalSize += stats.TotalSize
		}

		noRaw, err := cmd.Flags().GetBool("no-raw")
		if err != nil {
			logger.Error("Reading flag error: %v", err)
			return err
		}

		if !noRaw {
			singlePath := true
			if len(paths) > 1 {
				singlePath = false
			}

			for idx, rawFilePath := range paths {
				outPath := path.Join("test-results", "junit.xml")
				if !singlePath {
					outPath = path.Join("test-results", fmt.Sprintf("junit-%d.xml", idx))
				}

				_, stats, err = cli.PushArtifacts("job", rawFilePath, outPath, cmd)
				if err != nil {
					return err
				}
				if stats != nil {
					pushStats.Operations++
					pushStats.FileCount += stats.FileCount
					pushStats.TotalSize += stats.TotalSize
				}
			}
		}

		emptyPullStats := &cli.ArtifactStats{}
		cli.DisplayTransferSummary(emptyPullStats, pushStats)

		return nil
	},
}

func pushSummaryWithStats(testResult []parser.TestResults, level, path string, cmd *cobra.Command, pushStats *cli.ArtifactStats) error {
	skipCompression, err := cmd.Flags().GetBool("no-compress")
	if err != nil {
		return err
	}
	if len(testResult) == 0 {
		logger.Info("no test results to process")
		return nil
	}

	logger.Info("starting to generate summary")
	summaryReport := parser.Summary{}
	for _, results := range testResult {
		summary := results.Summary
		summaryReport.Merge(&summary)
	}

	jsonSummary, err := json.Marshal(summaryReport)
	if err != nil {
		return err
	}

	summaryFileName, err := cli.WriteToTmpFile(jsonSummary, !skipCompression)
	if err != nil {
		return err
	}
	defer os.Remove(summaryFileName)

	_, stats, err := cli.PushArtifacts(level, summaryFileName, path, cmd)
	if err != nil {
		return err
	}
	if stats != nil {
		pushStats.Operations++
		pushStats.FileCount += stats.FileCount
		pushStats.TotalSize += stats.TotalSize
	}
	return nil
}

func init() {

	desc := `Skips uploading raw XML files`
	publishCmd.Flags().BoolP("no-raw", "", false, desc)
	publishCmd.Flags().BoolP("force", "f", false, "force artifact push, passes -f flag to artifact CLI")
	publishCmd.Flags().Int32P("trim-output-to", "s", 0, "trim stdout to N characters, defaults to 0(unlimited)")
	publishCmd.Flags().BoolP("omit-output-for-passed", "o", false, "omit stdout if test passed, defaults to false")

	rootCmd.AddCommand(publishCmd)
}
