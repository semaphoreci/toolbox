package cmd

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

		generateMCPSummary, err := cmd.Flags().GetBool("generate-mcp-summary")
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

		if generateMCPSummary {
			if err = pushMCPSummariesWithStats(result, "workflow", path.Join("test-results", pipelineID+"-mcp-summary.json"), cmd, pushStats); err != nil {
				return err
			}
		}

		cli.DisplayTransferSummary(pullStats, pushStats)

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

func pushMCPSummariesWithStats(result *parser.Result, level, path string, cmd *cobra.Command, pushStats *cli.ArtifactStats) error {
	if len(result.TestResults) == 0 {
		logger.Info("no test results to process for MCP summary")
		return nil
	}

	logger.Info("starting to generate MCP summary")
	mcpResult := result.FilterFailedTests()

	jsonData, err := json.Marshal(mcpResult)
	if err != nil {
		return err
	}

	// Write without compression
	mcpFileName, err := cli.WriteToTmpFile(jsonData, false)
	if err != nil {
		return err
	}
	defer os.Remove(mcpFileName)

	_, stats, err := cli.PushArtifacts(level, mcpFileName, path, cmd)
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
	genPipelineReportCmd.Flags().BoolP("force", "f", false, "force artifact push, passes -f flag to artifact CLI")
	genPipelineReportCmd.Flags().Bool("generate-mcp-summary", false, "generate and push a summary with only failed tests (no compression)")
	rootCmd.AddCommand(genPipelineReportCmd)
}
