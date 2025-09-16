package parsers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/semaphoreci/toolbox/test-results/pkg/logger"
	"github.com/semaphoreci/toolbox/test-results/pkg/parser"
)

// ReviveIssue represents a single issue from revive JSON output
type ReviveIssue struct {
	Severity string `json:"Severity"`
	Failure  string `json:"Failure"`
	RuleName string `json:"RuleName"`
	Category string `json:"Category"`
	Position struct {
		Start RevivePosition `json:"Start"`
		End   RevivePosition `json:"End"`
	} `json:"Position"`
	Confidence      int    `json:"Confidence"`
	ReplacementLine string `json:"ReplacementLine"`
}

type RevivePosition struct {
	Filename string `json:"Filename"`
	Offset   int    `json:"Offset"`
	Line     int    `json:"Line"`
	Column   int    `json:"Column"`
}

// GoRevive parser for revive JSON output
type GoRevive struct{}

// NewGoRevive creates a new revive parser
func NewGoRevive() GoRevive {
	return GoRevive{}
}

// GetName returns the name of the parser
func (r GoRevive) GetName() string {
	return "go:revive"
}

// GetDescription returns the description of the parser
func (r GoRevive) GetDescription() string {
	return "Go revive linter JSON output"
}

// GetSupportedExtensions ...
func (r GoRevive) GetSupportedExtensions() []string {
	return []string{".json"}
}

// IsApplicable checks if the file contains revive JSON output
func (r GoRevive) IsApplicable(path string) bool {
	if filepath.Ext(path) != ".json" {
		return false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	var issues []ReviveIssue
	if err := json.Unmarshal(data, &issues); err != nil {
		return false
	}

	if len(issues) == 0 {
		return false
	}

	// Check for revive-specific fields
	for _, issue := range issues {
		// Revive issues should have these required fields
		if issue.RuleName != "" && issue.Category != "" && issue.Position.Start.Filename != "" {
			return true
		}
	}

	return false
}

// Parse transforms revive JSON output to test-results format
func (r GoRevive) Parse(path string) parser.TestResults {
	results := parser.NewTestResults()
	results.Name = "Go Revive Analysis"
	results.Framework = r.GetName()
	results.EnsureID()

	data, err := os.ReadFile(path)
	if err != nil {
		logger.Error("Failed to read file: %v", err)
		results.Status = parser.StatusError
		results.StatusMessage = fmt.Sprintf("Failed to read file: %v", err)
		return results
	}

	var issues []ReviveIssue
	if err := json.Unmarshal(data, &issues); err != nil {
		logger.Error("Failed to parse JSON: %v", err)
		results.Status = parser.StatusError
		results.StatusMessage = fmt.Sprintf("Failed to parse JSON: %v", err)
		return results
	}

	// Group issues by file
	fileIssues := make(map[string][]ReviveIssue)
	for _, issue := range issues {
		filename := issue.Position.Start.Filename
		if filename == "" {
			filename = "unknown"
		}
		fileIssues[filename] = append(fileIssues[filename], issue)
	}

	for filename, fileIssueList := range fileIssues {
		suite := parser.NewSuite()
		suite.Name = filename
		suite.EnsureID(results)

		for _, issue := range fileIssueList {
			test := parser.NewTest()
			test.Name = fmt.Sprintf("%s:%d:%d - %s",
				filepath.Base(issue.Position.Start.Filename),
				issue.Position.Start.Line,
				issue.Position.Start.Column,
				issue.RuleName)
			test.File = fmt.Sprintf("%s:%d", issue.Position.Start.Filename, issue.Position.Start.Line)
			test.Classname = issue.Category

			test.State = parser.StateFailed
			failure := parser.NewFailure()
			failure.Message = issue.Failure
			failure.Type = fmt.Sprintf("%s (%s)", issue.RuleName, issue.Severity)
			failure.Body = fmt.Sprintf("Rule: %s\nCategory: %s\nSeverity: %s\nConfidence: %d\n\n%s",
				issue.RuleName,
				issue.Category,
				issue.Severity,
				issue.Confidence,
				issue.Failure)
			test.Failure = &failure

			test.Duration = time.Millisecond
			test.EnsureID(suite)
			suite.Tests = append(suite.Tests, test)
		}

		suite.Aggregate()
		results.Suites = append(results.Suites, suite)
	}

	// If no issues found, create a passing test
	if len(issues) == 0 {
		suite := parser.NewSuite()
		suite.Name = "Revive"
		suite.EnsureID(results)

		test := parser.NewTest()
		test.Name = "No issues found"
		test.State = parser.StatePassed
		test.Duration = 1000000
		test.EnsureID(suite)
		suite.Tests = append(suite.Tests, test)
		suite.Aggregate()
		results.Suites = append(results.Suites, suite)
	}

	results.Aggregate()
	results.Status = parser.StatusSuccess

	logger.Info("Parsed %d revive issues from %d files", len(issues), len(fileIssues))

	return results
}
