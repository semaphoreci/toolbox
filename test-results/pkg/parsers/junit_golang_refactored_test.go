package parsers

import (
	"testing"

	"github.com/semaphoreci/toolbox/test-results/pkg/parser"
)

// Example of a refactored test using the new test helpers
func TestJUnitGoLangRefactored(t *testing.T) {
	p := NewJUnitGoLang()

	t.Run("Parser Metadata", func(t *testing.T) {
		if got := p.GetName(); got != "golang" {
			t.Errorf("GetName() = %q, want %q", got, "golang")
		}

		if got := p.GetDescription(); got == "" {
			t.Error("GetDescription() should not be empty")
		}
	})

	t.Run("Supported Extensions", func(t *testing.T) {
		AssertParserSupportsExtension(t, p, ".xml", true)
		AssertParserSupportsExtension(t, p, ".json", false)
	})

	t.Run("Parser Applicability", func(t *testing.T) {
		tests := []struct {
			name     string
			xml      string
			expected bool
		}{
			{
				name: "Accepts XML with go.version property",
				xml: `<?xml version="1.0"?>
					<testsuite name="test">
						<properties>
							<property name="go.version" value="go1.16"></property>
						</properties>
					</testsuite>`,
				expected: true,
			},
			{
				name: "Accepts testsuites with go.version",
				xml: `<?xml version="1.0"?>
					<testsuites>
						<testsuite name="test">
							<properties>
								<property name="go.version" value="go1.16"></property>
							</properties>
						</testsuite>
					</testsuites>`,
				expected: true,
			},
			{
				name: "Rejects XML without go.version",
				xml: `<?xml version="1.0"?>
					<testsuite name="test">
						<properties>
							<property name="other" value="value"></property>
						</properties>
					</testsuite>`,
				expected: false,
			},
			{
				name: "Rejects non-JUnit XML",
				xml: `<?xml version="1.0"?>
					<root>
						<item>test</item>
					</root>`,
				expected: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tmpFile := CreateTempFile(t, "*.xml", tt.xml)
				AssertParserApplicable(t, p, tmpFile, tt.expected)
			})
		}
	})

	t.Run("Parse Results", func(t *testing.T) {
		tests := []struct {
			name        string
			xml         string
			wantSummary parser.Summary
			wantStatus  parser.Status
			wantError   bool
		}{
			{
				name: "Empty file",
				xml:  "",
				wantSummary: parser.Summary{
					Total: 0,
				},
				wantStatus: parser.StatusError,
				wantError:  true,
			},
			{
				name: "Single test suite",
				xml: `<?xml version="1.0"?>
					<testsuite name="TestSuite" tests="2" failures="1" errors="0">
						<properties>
							<property name="go.version" value="go1.16"></property>
						</properties>
						<testcase name="TestPass" time="0.001"></testcase>
						<testcase name="TestFail" time="0.002">
							<failure message="assertion failed">Expected true, got false</failure>
						</testcase>
					</testsuite>`,
				wantSummary: parser.Summary{
					Total:  2,
					Passed: 1,
					Failed: 1,
					Error:  0,
				},
				wantStatus: parser.StatusSuccess,
				wantError:  false,
			},
			{
				name: "Multiple test suites",
				xml: `<?xml version="1.0"?>
					<testsuites name="AllTests" tests="4" failures="1" errors="1">
						<testsuite name="Suite1" tests="2" failures="1" errors="0">
							<properties>
								<property name="go.version" value="go1.16"></property>
							</properties>
							<testcase name="Test1" time="0.001"></testcase>
							<testcase name="Test2" time="0.002">
								<failure>Failed</failure>
							</testcase>
						</testsuite>
						<testsuite name="Suite2" tests="2" failures="0" errors="1">
							<properties>
								<property name="go.version" value="go1.16"></property>
							</properties>
							<testcase name="Test3" time="0.001"></testcase>
							<testcase name="Test4" time="0.002">
								<error>Error occurred</error>
							</testcase>
						</testsuite>
					</testsuites>`,
				wantSummary: parser.Summary{
					Total:  4,
					Passed: 2,
					Failed: 1,
					Error:  1,
				},
				wantStatus: parser.StatusSuccess,
				wantError:  false,
			},
			{
				name: "Invalid root element",
				xml: `<?xml version="1.0"?>
					<invalid>
						<test>data</test>
					</invalid>`,
				wantSummary: parser.Summary{
					Total: 0,
				},
				wantStatus: parser.StatusError,
				wantError:  true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tmpFile := CreateTempFile(t, "*.xml", tt.xml)
				result := p.Parse(tmpFile)

				// Check summary statistics
				AssertTestSummary(t, result.Summary, tt.wantSummary)

				// Check status
				if result.Status != tt.wantStatus {
					t.Errorf("Status = %q, want %q", result.Status, tt.wantStatus)
				}

				// Check if error message is present when expected
				if tt.wantError && result.StatusMessage == "" {
					t.Error("Expected error message but got none")
				}
				if !tt.wantError && result.StatusMessage != "" {
					t.Errorf("Unexpected error message: %s", result.StatusMessage)
				}
			})
		}
	})
}

// TestJUnitGoLangWithGoldenFile demonstrates using golden file testing
func TestJUnitGoLangWithGoldenFile(t *testing.T) {
	// Set up environment for consistent output
	t.Setenv("SEMAPHORE_PIPELINE_ID", "ppl-id")
	t.Setenv("SEMAPHORE_WORKFLOW_ID", "wf-id")
	t.Setenv("SEMAPHORE_JOB_NAME", "job-name")
	t.Setenv("SEMAPHORE_JOB_ID", "job-id")
	t.Setenv("SEMAPHORE_PROJECT_ID", "project-id")
	t.Setenv("SEMAPHORE_AGENT_MACHINE_TYPE", "agent-machine-type")
	t.Setenv("SEMAPHORE_AGENT_MACHINE_OS_IMAGE", "agent-machine-os-image")
	t.Setenv("SEMAPHORE_JOB_CREATION_TIME", "job-creation-time")
	t.Setenv("SEMAPHORE_GIT_REF_TYPE", "git-ref-type")

	test := GoldenTest{
		Name:       "JUnitGoLang",
		Parser:     NewJUnitGoLang(),
		InputFile:  FixturePath("priv/parsers/junit_golang/in.xml"),
		GoldenFile: FixturePath("priv/parsers/junit_golang/out.json"),
	}

	RunGoldenTest(t, test)
}
