package cli_test

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/semaphoreci/toolbox/test-results/pkg/parser"

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
