package parsers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/semaphoreci/toolbox/test-results/pkg/logger"
	"github.com/semaphoreci/toolbox/test-results/pkg/parser"
)

// StaticcheckIssue represents a single issue from staticcheck JSON output
type StaticcheckIssue struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Location struct {
		File   string `json:"file"`
		Line   int    `json:"line"`
		Column int    `json:"column"`
	} `json:"location"`
	End struct {
		File   string `json:"file"`
		Line   int    `json:"line"`
		Column int    `json:"column"`
	} `json:"end"`
	Message string `json:"message"`
}

// GoStaticcheck parser for staticcheck JSON output
type GoStaticcheck struct{}

// NewGoStaticcheck creates a new staticcheck parser
func NewGoStaticcheck() GoStaticcheck {
	return GoStaticcheck{}
}

// GetName returns the name of the parser
func (s GoStaticcheck) GetName() string {
	return "go:staticcheck"
}

// GetDescription returns the description of the parser
func (s GoStaticcheck) GetDescription() string {
	return "Staticcheck linter output (newline-delimited JSON)"
}

// GetSupportedExtensions ...
func (s GoStaticcheck) GetSupportedExtensions() []string {
	return []string{".json"}
}

// IsApplicable checks if the file contains staticcheck JSON output
func (s GoStaticcheck) IsApplicable(path string) bool {
	logger.Debug("Checking applicability of %s parser", s.GetName())

	data, err := LoadFile(path)
	if err != nil {
		logger.Debug("Failed to load file as JSON: %v", err)
		return false
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var issue StaticcheckIssue
		if err := json.Unmarshal([]byte(line), &issue); err != nil {
			continue
		}

		// Check if it has the expected staticcheck structure
		if issue.Code != "" && issue.Location.File != "" && issue.Message != "" {
			logger.Debug("Detected staticcheck JSON format")
			return true
		}
	}

	return false
}

// Parse transforms staticcheck JSON output to test-results format
func (s GoStaticcheck) Parse(path string) parser.TestResults {
	results := parser.NewTestResults()
	results.Name = "Go Staticcheck Analysis"
	results.Framework = s.GetName()
	results.EnsureID()

	data, err := LoadFile(path)
	if err != nil {
		logger.Error("Failed to load file: %v", err)
		results.Status = parser.StatusError
		results.StatusMessage = err.Error()
		return results
	}

	// Parse JSON lines (staticcheck outputs one JSON object per line)
	var issues []StaticcheckIssue
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if line == "" {
			continue
		}

		var issue StaticcheckIssue
		if err := json.Unmarshal([]byte(line), &issue); err != nil {
			logger.Debug("Failed to parse line %d: %v", lineNum, err)
			continue
		}
		issues = append(issues, issue)
	}

	if err := scanner.Err(); err != nil {
		logger.Error("Error reading JSON lines: %v", err)
		results.Status = parser.StatusError
		results.StatusMessage = err.Error()
		return results
	}

	// Group issues by file
	fileGroups := groupIssuesByFile(issues)

	for file, fileIssues := range fileGroups {
		suite := parser.NewSuite()
		suite.Name = filepath.Base(file)
		suite.Package = filepath.Dir(file)
		suite.EnsureID(results)

		for _, issue := range fileIssues {
			test := parser.NewTest()
			test.Name = fmt.Sprintf("%s (line %d:%d)", issue.Code, issue.Location.Line, issue.Location.Column)
			test.Classname = issue.Code
			test.File = issue.Location.File
			test.Duration = time.Millisecond // Minimal duration

			switch issue.Severity {
			case "error":
				test.State = parser.StateError
				test.Error = &parser.Error{
					Type:    issue.Code,
					Message: issue.Message,
					Body:    fmt.Sprintf("Location: %s:%d:%d\n%s", issue.Location.File, issue.Location.Line, issue.Location.Column, issue.Message),
				}
			default: // warning, info, etc.
				test.State = parser.StateFailed
				test.Failure = &parser.Failure{
					Type:    issue.Code,
					Message: issue.Message,
					Body:    fmt.Sprintf("Location: %s:%d:%d\n%s", issue.Location.File, issue.Location.Line, issue.Location.Column, issue.Message),
				}
			}

			test.EnsureID(suite)
			suite.Tests = append(suite.Tests, test)
		}

		suite.Aggregate()
		results.Suites = append(results.Suites, suite)
	}

	if len(issues) == 0 {
		suite := parser.NewSuite()
		suite.Name = "Staticcheck"
		suite.EnsureID(results)

		test := parser.NewTest()
		test.Name = "No issues found"
		test.State = parser.StatePassed
		test.Duration = time.Millisecond
		test.EnsureID(suite)
		suite.Tests = append(suite.Tests, test)

		suite.Aggregate()
		results.Suites = append(results.Suites, suite)
	}

	slices.SortFunc(results.Suites, func(suite1, suite2 parser.Suite) int {
		return strings.Compare(suite1.ID, suite2.ID)
	})

	results.Aggregate()
	results.Status = parser.StatusSuccess

	logger.Info("Parsed %d staticcheck issues from %d files", len(issues), len(fileGroups))

	return results
}

// groupIssuesByFile groups issues by their file path
func groupIssuesByFile(issues []StaticcheckIssue) map[string][]StaticcheckIssue {
	groups := make(map[string][]StaticcheckIssue)
	for _, issue := range issues {
		groups[issue.Location.File] = append(groups[issue.Location.File], issue)
	}
	return groups
}
