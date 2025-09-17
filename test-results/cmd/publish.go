package cmd

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

You can specify parsers for individual files using the syntax:
  file.xml:parser-name

Examples:
  test-results publish results.xml
  test-results publish results.xml:golang lint.json:go:staticcheck
  test-results publish --ignore-missing test1.xml test2.xml test3.xml

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

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish <file-path>...",
	Short: "parses test result files to well defined json schema and publishes results to artifacts storage",
	Long:  formatPublishDescription(),
	Args:  cobra.MinimumNArgs(1),
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
		var rawFilePaths []string

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
							rawFilePaths = append(rawFilePaths, path)
						}
					}
					return nil
				})
				if err != nil {
					return err
				}
			} else {
				allPairs = append(allPairs, pair)
				rawFilePaths = append(rawFilePaths, pair.Path)
			}
		}

		if len(allPairs) == 0 {
			logger.Warn("No files to process")
			return nil
		}

		dirPath, err := os.MkdirTemp("", "test-results-*")
		if err != nil {
			return err
		}
		defer os.RemoveAll(dirPath)

		pushStats := &cli.ArtifactStats{}

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
			if len(rawFilePaths) > 1 {
				singlePath = false
			}

			for idx, rawFilePath := range rawFilePaths {
				pathExt := path.Ext(rawFilePath)
				outPath := path.Join("test-results", fmt.Sprintf("junit%s", pathExt))
				if !singlePath {
					outPath = path.Join("test-results", fmt.Sprintf("junit-%d%s", idx, pathExt))
				}

				_, stats, err = cli.PushArtifacts("job", rawFilePath, outPath, cmd)
				if err != nil {
					if ignoreMissing && os.IsNotExist(err) {
						logger.Warn("Raw file no longer exists, skipping: %s", rawFilePath)
						continue
					}
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

	desc := `Skips uploading raw input files`
	publishCmd.Flags().BoolP("no-raw", "", false, desc)
	publishCmd.Flags().BoolP("force", "f", false, "force artifact push, passes -f flag to artifact CLI")
	publishCmd.Flags().BoolP("omit-output-for-passed", "o", false, "omit stdout if test passed, defaults to false")
	publishCmd.Flags().Bool("ignore-missing", false, "ignore missing files instead of failing")

	rootCmd.AddCommand(publishCmd)
}
