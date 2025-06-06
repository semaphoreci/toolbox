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

	"github.com/semaphoreci/toolbox/test-results/pkg/cli"
	"github.com/semaphoreci/toolbox/test-results/pkg/logger"
	"github.com/spf13/cobra"
)

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish <xml-file-path>...",
	Short: "parses xml file to well defined json schema and publishes results to artifacts storage",
	Long: `Parses xml file to well defined json schema and publishes results to artifacts storage

	It traverses through directory structure specified by <xml-file-path>, compiles
	every .xml file and publishes it as one artifact.
	`,
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

		paths, err := cli.LoadFiles(inputs, ".xml")
		if err != nil {
			return err
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

		_, err = cli.PushArtifacts("job", fileName, path.Join("test-results", "junit.json"), cmd)
		if err != nil {
			return err
		}

		if err = pushSummaries(result.TestResults, "job", path.Join("test-results", "summary.json"), cmd); err != nil {
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

		_, err = cli.PushArtifacts("workflow", fileName, path.Join("test-results", pipelineID, jobID+".json"), cmd)
		if err != nil {
			return err
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

				_, err = cli.PushArtifacts("job", rawFilePath, outPath, cmd)
				if err != nil {
					return err
				}
			}
		}

		return nil
	},
}

func init() {

	desc := `Skips uploading raw XML files`
	publishCmd.Flags().BoolP("no-raw", "", false, desc)
	publishCmd.Flags().BoolP("force", "f", false, "force artifact push, passes -f flag to artifact CLI")
	publishCmd.Flags().Int32P("trim-output-to", "s", 0, "trim stdout to N characters, defaults to 0(unlimited)")
	publishCmd.Flags().BoolP("omit-output-for-passed", "o", false, "omit stdout if test passed, defaults to false")

	rootCmd.AddCommand(publishCmd)
}
