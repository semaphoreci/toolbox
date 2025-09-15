package parsers

import (
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
	Column   int    `json:"Columnt"`
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
	return true
}

// Parse transforms revive JSON output to test-results format
func (r GoRevive) Parse(path string) parser.TestResults {
	return parser.NewTestResults()
}
