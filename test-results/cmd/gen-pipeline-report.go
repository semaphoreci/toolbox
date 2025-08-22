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
	"os"
	"path"

	"github.com/semaphoreci/toolbox/test-results/pkg/cli"
	"github.com/semaphoreci/toolbox/test-results/pkg/logger"
	"github.com/semaphoreci/toolbox/test-results/pkg/parser"
	"github.com/spf13/cobra"
)

// genPipelineReportCmd represents the publish command
var genPipelineReportCmd = &cobra.Command{
	Use:   "gen-pipeline-report [<path>...]",
	Short: "fetches workflow level JUnit reports and combines them together",
	Long: `fetches workflow level junit reports and combines them

	When <path>s are provided it recursively traverses through path structure and
	combines all .json files into one JSON schema file.
	`,
	Args: cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		err := cli.SetLogLevel(cmd)
		if err != nil {
			return err
		}
		skipCompression, err := cmd.Flags().GetBool("no-compress")
		if err != nil {
			return err
		}

		pullStats := &cli.ArtifactStats{}
		pushStats := &cli.ArtifactStats{}

		var dir string

		pipelineID, found := os.LookupEnv("SEMAPHORE_PIPELINE_ID")
		if !found {
			logger.Error("SEMAPHORE_PIPELINE_ID env is missing")
			return err
		}

		if len(args) == 0 {
			dir, err = os.MkdirTemp("", "test-results")
			if err != nil {
				logger.Error("Creating temporary directory failed %v", err)
				return err
			}
			defer os.Remove(dir)

			var stats *cli.ArtifactStats
			dir, stats, err = cli.PullArtifacts("workflow", path.Join("test-results", pipelineID), dir, cmd)
			if err != nil {
				return err
			}
			if stats != nil {
				pullStats.Operations++
				pullStats.FileCount += stats.FileCount
				pullStats.TotalSize += stats.TotalSize
			}
		} else {
			dir = args[0]
		}

		result, err := cli.MergeFiles(dir, cmd)
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

		_, stats, err := cli.PushArtifacts("workflow", fileName, path.Join("test-results", pipelineID+".json"), cmd)
		if err != nil {
			return err
		}
		if stats != nil {
			pushStats.Operations++
			pushStats.FileCount += stats.FileCount
			pushStats.TotalSize += stats.TotalSize
		}

		err = pushSummariesWithStats(result.TestResults, "workflow", path.Join("test-results", pipelineID+"-summary.json"), cmd, pushStats)
		if err != nil {
			return err
		}

		displayTransferSummary(pullStats, pushStats)

		return nil
	},
}

func pushSummariesWithStats(testResult []parser.TestResults, level, path string, cmd *cobra.Command, pushStats *cli.ArtifactStats) error {
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

func displayTransferSummary(pullStats *cli.ArtifactStats, pushStats *cli.ArtifactStats) {
	totalOps := pullStats.Operations + pushStats.Operations
	if totalOps > 0 {
		logger.Info("")
		logger.Info("========================================")
		logger.Info("test-results: Artifact Transfer Summary")
		logger.Info("========================================")
		
		if pullStats.Operations > 0 {
			if pullStats.FileCount > 0 || pullStats.TotalSize > 0 {
				logger.Info("Pull operations: %d (%d files, %s)", pullStats.Operations, pullStats.FileCount, cli.FormatBytes(pullStats.TotalSize))
			} else {
				logger.Info("Pull operations: %d", pullStats.Operations)
			}
		}
		
		if pushStats.Operations > 0 {
			if pushStats.FileCount > 0 || pushStats.TotalSize > 0 {
				logger.Info("Push operations: %d (%d files, %s)", pushStats.Operations, pushStats.FileCount, cli.FormatBytes(pushStats.TotalSize))
			} else {
				logger.Info("Push operations: %d", pushStats.Operations)
			}
		}
		
		totalFiles := pullStats.FileCount + pushStats.FileCount
		totalSize := pullStats.TotalSize + pushStats.TotalSize
		if totalFiles > 0 || totalSize > 0 {
			logger.Info("Total: %d operations (%d files, %s)", totalOps, totalFiles, cli.FormatBytes(totalSize))
		} else {
			logger.Info("Total: %d operations", totalOps)
		}
		
		logger.Info("========================================")
	}
}

func init() {
	genPipelineReportCmd.Flags().BoolP("force", "f", false, "force artifact push, passes -f flag to artifact CLI")
	rootCmd.AddCommand(genPipelineReportCmd)
}
