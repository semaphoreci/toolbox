package cli_test

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/semaphoreci/toolbox/test-results/pkg/parser"
	"github.com/spf13/cobra"

	"github.com/semaphoreci/toolbox/test-results/pkg/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_LoadFiles(t *testing.T) {
	t.Run("with invalid path to file", func(t *testing.T) {
		filePath := generateFile(t)
		paths, err := cli.LoadFiles([]string{fmt.Sprintf("%s1", filePath)}, ".xml")

		assert.Len(t, paths, 0, "should return correct number of files")
		assert.NotNil(t, err, "should throw error")
		os.RemoveAll(filePath)
	})

	t.Run("with single file", func(t *testing.T) {
		filePath := generateFile(t)
		paths, err := cli.LoadFiles([]string{filePath}, ".xml")

		assert.Equal(t, filePath, paths[0], "should contain correct file path")
		assert.Len(t, paths, 1, "should return correct number of files")
		assert.Nil(t, err, "should not throw error")
		os.RemoveAll(filePath)
	})

	t.Run("with directory", func(t *testing.T) {
		dirPath := generateDir(t)
		assert.NotEqual(t, "", dirPath)

		paths, err := cli.LoadFiles([]string{dirPath}, ".xml")
		assert.Len(t, paths, 5, "should return correct number of files")
		assert.Nil(t, err, "should not throw error")

		os.RemoveAll(dirPath)
	})

	t.Run("with big directory", func(t *testing.T) {
		dirPath := generateDirWithFilesAndNestedDir(t, 2600, 3)
		assert.NotEmpty(t, dirPath)

		paths, err := cli.LoadFiles([]string{dirPath}, ".xml")
		assert.Len(t, paths, 2600, "should return correct number of files")
		assert.Nil(t, err, "should not throw error")

		os.RemoveAll(dirPath)
	})

}

func generateFile(t *testing.T) string {
	filePath, err := os.CreateTemp("", "file-*.xml")
	if err != nil {
		t.Errorf("Failed to create temporary file: %v", err)
	}

	return filePath.Name()
}

func generateDir(t *testing.T) string {
	return generateDirWithFilesAndNestedDir(t, 5, 3)
}

func generateDirWithFilesAndNestedDir(t *testing.T, fNumber, dirNumber int) string {
	dirPath, err := os.MkdirTemp("", "random-dir-*")
	assert.Nil(t, err)

	xmlNestedDir, err := os.MkdirTemp(dirPath, "xml-*")
	assert.Nil(t, err)

	jsonNestedDir, err := os.MkdirTemp(dirPath, "json-*")
	assert.Nil(t, err)

	for i := 0; i < fNumber; i++ {
		_, err = os.CreateTemp(xmlNestedDir, "file-*.xml")
		assert.Nil(t, err)
	}

	for i := 0; i < dirNumber; i++ {
		_, err := os.CreateTemp(jsonNestedDir, "file-*.json")
		assert.Nil(t, err)
	}

	return dirPath
}

func TestWriteToTmpFile(t *testing.T) {
	tr := parser.TestResults{
		ID:         "1234",
		Name:       "Test",
		Framework:  "JUnit",
		IsDisabled: false,
		Suites:     nil,
		Summary: parser.Summary{
			Total:    10,
			Passed:   5,
			Skipped:  0,
			Error:    0,
			Failed:   5,
			Disabled: 0,
			Duration: 360,
		},
		Status:        "OK",
		StatusMessage: "Test",
	}
	result := parser.Result{TestResults: []parser.TestResults{tr}}
	jsonData, _ := json.Marshal(&result)

	t.Run("Write to one tmp file", func(t *testing.T) {
		file, err := cli.WriteToTmpFile(jsonData, false)
		assert.NoError(t, err)
		os.Remove(file)
	})

	t.Run("Write to three thousand tmp files", func(t *testing.T) {
		fileNumber := 3000

		var wg sync.WaitGroup
		errChan := make(chan error, fileNumber)

		wg.Add(fileNumber)

		for i := 0; i < fileNumber; i++ {
			go func(i int) {
				defer wg.Done()
				file, err := cli.WriteToTmpFile(jsonData, false)
				defer os.Remove(file)
				if err != nil {
					errChan <- err
					return
				}
			}(i)
		}

		wg.Wait()
		close(errChan)

		for err := range errChan {
			require.NoError(t, err)
		}
	})
}

func TestApplyOutputTrimming(t *testing.T) {
	longText := func(n int) string {
		s := ""
		for i := 0; i < n; i++ {
			s += "x"
		}
		return s
	}

	createTestResult := func(text string) *parser.Result {
		return &parser.Result{
			TestResults: []parser.TestResults{
				{
					ID:   "test-1",
					Name: "Test",
					Suites: []parser.Suite{
						{
							Name:      "Suite1",
							SystemOut: text,
							SystemErr: text,
							Tests: []parser.Test{
								{
									Name:      "Test1",
									SystemOut: text,
									SystemErr: text,
									Failure: &parser.Failure{
										Message: text,
										Body:    text,
									},
								},
							},
						},
					},
				},
			},
		}
	}

	createCmd := func(trimTo int, noTrim bool) *cobra.Command {
		cmd := &cobra.Command{}
		cmd.Flags().Int("trim-output-to", trimTo, "")
		cmd.Flags().Bool("no-trim-output", noTrim, "")
		return cmd
	}

	t.Run("default trimming to 1000 characters", func(t *testing.T) {
		result := createTestResult(longText(2000))
		cmd := createCmd(1000, false)

		cli.ApplyOutputTrimming(result, cmd)

		suite := result.TestResults[0].Suites[0]
		assert.True(t, len(suite.SystemOut) <= 1000+len("...[truncated]...\n"))
		assert.True(t, len(suite.SystemErr) <= 1000+len("...[truncated]...\n"))
		assert.Contains(t, suite.SystemOut, "...[truncated]...")
	})

	t.Run("custom trim length", func(t *testing.T) {
		result := createTestResult(longText(5000))
		cmd := createCmd(3000, false)

		cli.ApplyOutputTrimming(result, cmd)

		suite := result.TestResults[0].Suites[0]
		assert.True(t, len(suite.SystemOut) <= 3000+len("...[truncated]...\n"))
		assert.Contains(t, suite.SystemOut, "...[truncated]...")
	})

	t.Run("no trimming when --no-trim-output is set", func(t *testing.T) {
		originalText := longText(5000)
		result := createTestResult(originalText)
		cmd := createCmd(1000, true)

		cli.ApplyOutputTrimming(result, cmd)

		suite := result.TestResults[0].Suites[0]
		assert.Equal(t, originalText, suite.SystemOut)
		assert.Equal(t, originalText, suite.SystemErr)
		assert.Equal(t, originalText, suite.Tests[0].SystemOut)
		assert.Equal(t, originalText, suite.Tests[0].Failure.Message)
	})

	t.Run("no trimming when --trim-output-to is 0", func(t *testing.T) {
		originalText := longText(5000)
		result := createTestResult(originalText)
		cmd := createCmd(0, false)

		cli.ApplyOutputTrimming(result, cmd)

		suite := result.TestResults[0].Suites[0]
		assert.Equal(t, originalText, suite.SystemOut)
		assert.Equal(t, originalText, suite.SystemErr)
	})

	t.Run("no trimming when --trim-output-to is negative", func(t *testing.T) {
		originalText := longText(5000)
		result := createTestResult(originalText)
		cmd := createCmd(-1, false)

		cli.ApplyOutputTrimming(result, cmd)

		suite := result.TestResults[0].Suites[0]
		assert.Equal(t, originalText, suite.SystemOut)
	})

	t.Run("text shorter than trim limit is not modified", func(t *testing.T) {
		originalText := longText(500)
		result := createTestResult(originalText)
		cmd := createCmd(1000, false)

		cli.ApplyOutputTrimming(result, cmd)

		suite := result.TestResults[0].Suites[0]
		assert.Equal(t, originalText, suite.SystemOut)
		assert.NotContains(t, suite.SystemOut, "...[truncated]...")
	})

	t.Run("nil result does not panic", func(t *testing.T) {
		cmd := createCmd(1000, false)
		assert.NotPanics(t, func() {
			cli.ApplyOutputTrimming(nil, cmd)
		})
	})

	t.Run("trims failure and error fields", func(t *testing.T) {
		result := &parser.Result{
			TestResults: []parser.TestResults{
				{
					ID: "test-1",
					Suites: []parser.Suite{
						{
							Tests: []parser.Test{
								{
									Failure: &parser.Failure{
										Message: longText(2000),
										Type:    longText(2000),
										Body:    longText(2000),
									},
									Error: &parser.Error{
										Message: longText(2000),
										Type:    longText(2000),
										Body:    longText(2000),
									},
								},
							},
						},
					},
				},
			},
		}
		cmd := createCmd(1000, false)

		cli.ApplyOutputTrimming(result, cmd)

		test := result.TestResults[0].Suites[0].Tests[0]
		assert.Contains(t, test.Failure.Message, "...[truncated]...")
		assert.Contains(t, test.Failure.Body, "...[truncated]...")
		assert.Contains(t, test.Error.Message, "...[truncated]...")
		assert.Contains(t, test.Error.Body, "...[truncated]...")
	})
}

func TestWriteToFilePath(t *testing.T) {
	tr := parser.TestResults{
		ID:         "1234",
		Name:       "Test",
		Framework:  "JUnit",
		IsDisabled: false,
		Suites:     nil,
		Summary: parser.Summary{
			Total:    10,
			Passed:   5,
			Skipped:  0,
			Error:    0,
			Failed:   5,
			Disabled: 0,
			Duration: 360,
		},
		Status:        "OK",
		StatusMessage: "Test",
	}
	result := parser.Result{TestResults: []parser.TestResults{tr}}
	jsonData, _ := json.Marshal(&result)

	t.Run("Write to one file", func(t *testing.T) {
		file, err := cli.WriteToFilePath(jsonData, "out", false)
		assert.NoError(t, err)
		os.Remove(file)
	})

	t.Run("Write to three thousand files", func(t *testing.T) {
		fileNumber := 3000
		dirPath, err := os.MkdirTemp("", "test-results-*")
		require.NoError(t, err)

		defer os.RemoveAll(dirPath)

		var wg sync.WaitGroup
		errChan := make(chan error, fileNumber)

		wg.Add(fileNumber)

		for i := 0; i < fileNumber; i++ {
			go func(i int) {
				defer wg.Done()
				tmpFile, err := os.CreateTemp(dirPath, "result-*.json")
				if err != nil {
					errChan <- err
					return
				}

				_, err = cli.WriteToFilePath(jsonData, tmpFile.Name(), false)
				if err != nil {
					errChan <- err
					return
				}
			}(i)
		}

		wg.Wait()
		close(errChan)

		for err := range errChan {
			require.NoError(t, err)
		}
	})
}
