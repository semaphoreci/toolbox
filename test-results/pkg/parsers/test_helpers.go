package parsers

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/semaphoreci/toolbox/test-results/pkg/parser"
)

var updateGolden = flag.Bool("update-golden", false, "Update golden test files")

// GoldenTest represents a test case with golden files
type GoldenTest struct {
	Name       string
	Parser     parser.Parser
	InputFile  string // Path to input file (e.g., "priv/parsers/junit_golang/in.xml")
	GoldenFile string // Path to golden output file (e.g., "priv/parsers/junit_golang/out.json")
}

// RunGoldenTest runs a parser test against golden files
func RunGoldenTest(t *testing.T, test GoldenTest) {
	t.Helper()

	// Parse the input file
	result := test.Parser.Parse(test.InputFile)

	actualJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	if *updateGolden {
		// Update mode: write the actual output as the new golden file
		err := os.MkdirAll(filepath.Dir(test.GoldenFile), 0755) // #nosec
		if err != nil {
			t.Fatalf("Failed to create golden file directory: %v", err)
		}

		err = os.WriteFile(test.GoldenFile, actualJSON, 0644) // #nosec
		if err != nil {
			t.Fatalf("Failed to update golden file: %v", err)
		}

		t.Logf("Updated golden file: %s", test.GoldenFile)
		return
	}

	// Read the golden file
	expectedJSON, err := os.ReadFile(test.GoldenFile) // #nosec
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("Golden file does not exist: %s\nRun with -update-golden flag to create it", test.GoldenFile)
		}
		t.Fatalf("Failed to read golden file: %v", err)
	}

	// Compare JSON strings
	if string(actualJSON) != string(expectedJSON) {
		// Parse both for better diff output
		var expected, actual interface{}
		err = json.Unmarshal(expectedJSON, &expected)
		if err != nil {
			t.Fatalf("Failed to unmarshal expected json: %v", err)
		}
		err = json.Unmarshal(actualJSON, &actual)
		if err != nil {
			t.Fatalf("Failed to unmarshal actual json: %v", err)
		}

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("Golden file mismatch (-want +got):\n%s", diff)
			t.Logf("Run with -update-golden flag to update the golden file")
		}
	}
}

// AssertParserApplicable checks that a parser correctly identifies applicable files
func AssertParserApplicable(t *testing.T, p parser.Parser, path string, expected bool) {
	t.Helper()

	actual := p.IsApplicable(path)
	if actual != expected {
		t.Errorf("Parser %s: IsApplicable(%q) = %v, want %v",
			p.GetName(), path, actual, expected)
	}
}

// AssertParserSupportsExtension checks that a parser supports a given extension
func AssertParserSupportsExtension(t *testing.T, p parser.Parser, ext string, expected bool) {
	t.Helper()

	extensions := p.GetSupportedExtensions()
	found := false
	for _, e := range extensions {
		if e == ext {
			found = true
			break
		}
	}

	if found != expected {
		t.Errorf("Parser %s: extension %q support = %v, want %v (supported: %v)",
			p.GetName(), ext, found, expected, extensions)
	}
}

// CreateTempFile creates a temporary file with the given content for testing
func CreateTempFile(t *testing.T, pattern string, content string) string {
	t.Helper()

	tmpfile, err := os.CreateTemp("", pattern)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	t.Cleanup(func() {
		err = os.Remove(tmpfile.Name())
		if err != nil {
			t.Fatalf("Failed to remove tmp file: %v", err)
		}
	})

	return tmpfile.Name()
}

// AssertTestSummary checks only the summary statistics of test results
func AssertTestSummary(t *testing.T, got, want parser.Summary) {
	t.Helper()

	if got.Total != want.Total {
		t.Errorf("Total tests: got %d, want %d", got.Total, want.Total)
	}
	if got.Passed != want.Passed {
		t.Errorf("Passed tests: got %d, want %d", got.Passed, want.Passed)
	}
	if got.Failed != want.Failed {
		t.Errorf("Failed tests: got %d, want %d", got.Failed, want.Failed)
	}
	if got.Error != want.Error {
		t.Errorf("Error tests: got %d, want %d", got.Error, want.Error)
	}
	if got.Skipped != want.Skipped {
		t.Errorf("Skipped tests: got %d, want %d", got.Skipped, want.Skipped)
	}
}

// LoadFixture loads a test fixture file
func LoadFixture(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path) // #nosec
	if err != nil {
		t.Fatalf("Failed to load fixture %s: %v", path, err)
	}

	return string(content)
}

// GetProjectRoot returns the project root directory for finding test fixtures
func GetProjectRoot() string {
	// Try to find the project root by looking for go.mod
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			return ""
		}
		dir = parent
	}
}

// FixturePath returns the full path to a fixture file
func FixturePath(relativePath string) string {
	root := GetProjectRoot()
	if root == "" {
		return relativePath
	}
	return filepath.Join(root, relativePath)
}
