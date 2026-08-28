package parsers

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/semaphoreci/toolbox/test-results/pkg/parser"
)

const semaphoreJSONLHeader = `{"kind":"report","schema":"semaphore/test-report","version":"2.0-draft.2","reporter":"semaphore-test-reports-rspec 0.1.0","framework":"rspec","hostname":"job-42a1","seed":48163,"started_at":"2026-08-28T09:12:41.676Z"}`

const semaphoreJSONLPassed = `{"kind":"test","identity":{"name":"Calculator::Adder sums arguments","classname":"spec.calculator.adder_spec","file":"spec/calculator/adder_spec.rb"},"state":"passed","duration_ms":0.436,"tags":{},"attempts":[{"index":0,"state":"passed","duration_ms":0.436}]}`

const semaphoreJSONLFailed = `{"kind":"test","identity":{"name":"Calculator::Adder subtracts arguments","classname":"spec.calculator.adder_spec","file":"spec/calculator/adder_spec.rb"},"state":"failed","duration_ms":12.473,"tags":{},"attempts":[{"index":0,"state":"failed","duration_ms":12.473,"failure":{"type":"RSpec::Expectations::ExpectationNotMetError","message":"expected: -1\n     got: 3","stack":"Failure/Error: expect(result).to eq(-1)"}}]}`

const semaphoreJSONLSkipped = `{"kind":"test","identity":{"name":"Calculator::Divider is pending","classname":"spec.calculator.divider_spec","file":"spec/calculator/divider_spec.rb"},"state":"skipped","tags":{},"attempts":[{"index":0,"state":"skipped","skip_message":"not implemented"}]}`

const semaphoreJSONLRetried = `{"kind":"test","identity":{"name":"Calculator::Adder is flaky","classname":"spec.calculator.adder_spec","file":"spec/calculator/adder_spec.rb"},"state":"passed","duration_ms":1.9,"tags":{},"attempts":[{"index":0,"state":"error","kind":"framework_retry","failure":{"type":"RuntimeError","message":"first attempt fails"}},{"index":1,"state":"passed","kind":"framework_retry","duration_ms":1.9}]}`

func TestSemaphoreJSONL(t *testing.T) {
	// Set up environment variables for consistent test output
	t.Setenv("SEMAPHORE_PIPELINE_ID", "ppl-id")
	t.Setenv("SEMAPHORE_WORKFLOW_ID", "wf-id")
	t.Setenv("SEMAPHORE_JOB_NAME", "job-name")
	t.Setenv("SEMAPHORE_JOB_ID", "job-id")
	t.Setenv("SEMAPHORE_PROJECT_ID", "project-id")
	t.Setenv("SEMAPHORE_AGENT_MACHINE_TYPE", "agent-machine-type")
	t.Setenv("SEMAPHORE_AGENT_MACHINE_OS_IMAGE", "agent-machine-os-image")
	t.Setenv("SEMAPHORE_JOB_CREATION_TIME", "job-creation-time")
	t.Setenv("SEMAPHORE_GIT_REF_TYPE", "git-ref-type")

	t.Run("Golden File Test", func(t *testing.T) {
		test := GoldenTest{
			Name:       "SemaphoreJSONL",
			Parser:     NewSemaphoreJSONL(),
			InputFile:  FixturePath("priv/parsers/semaphore_jsonl/in.jsonl"),
			GoldenFile: FixturePath("priv/parsers/semaphore_jsonl/out.json"),
		}
		RunGoldenTest(t, test)
	})

	t.Run("Parser Identification", func(t *testing.T) {
		p := NewSemaphoreJSONL()

		AssertParserApplicable(t, p, FixturePath("priv/parsers/semaphore_jsonl/in.jsonl"), true)

		AssertParserApplicable(t, p, FixturePath("priv/parsers/junit_rspec/in.xml"), false)
		AssertParserApplicable(t, p, FixturePath("priv/parsers/go_staticcheck/in.json"), false)
		AssertParserApplicable(t, p, CreateTempFile(t, "not-a-report-*.jsonl", "{\"kind\":\"test\"}\n"), false)
		AssertParserApplicable(t, p, CreateTempFile(t, "empty-*.jsonl", ""), false)
	})

	t.Run("Extensions", func(t *testing.T) {
		p := NewSemaphoreJSONL()
		AssertParserSupportsExtension(t, p, ".jsonl", true)
		AssertParserSupportsExtension(t, p, ".json", true)
		AssertParserSupportsExtension(t, p, ".xml", false)
	})

	t.Run("Registered For Directory Walking", func(t *testing.T) {
		found := false
		for _, ext := range GetSupportedExtensions() {
			if ext == ".jsonl" {
				found = true
			}
		}

		if !found {
			t.Errorf("Extension .jsonl is not registered, got %v", GetSupportedExtensions())
		}

		path := CreateTempFile(t, "auto-*.jsonl", semaphoreJSONLHeader+"\n"+semaphoreJSONLPassed+"\n")
		p, err := FindParser("auto", path)
		if err != nil {
			t.Fatalf("FindParser failed: %v", err)
		}

		if p.GetName() != "semaphore-jsonl" {
			t.Errorf("FindParser picked %s, want semaphore-jsonl", p.GetName())
		}
	})

	t.Run("Records", func(t *testing.T) {
		tests := []struct {
			name    string
			content string
			summary parser.Summary
			message string
		}{
			{
				name:    "empty report",
				content: semaphoreJSONLHeader + "\n",
				summary: parser.Summary{},
				message: "truncated: missing summary trailer",
			},
			{
				name:    "zero tests with trailer",
				content: semaphoreJSONLHeader + "\n" + `{"kind":"summary","tests":0,"complete":true}` + "\n",
				summary: parser.Summary{},
				message: "",
			},
			{
				name:    "torn report",
				content: semaphoreJSONLHeader + "\n" + semaphoreJSONLPassed + "\n" + semaphoreJSONLFailed + "\n",
				summary: parser.Summary{Total: 2, Passed: 1, Failed: 1},
				message: "truncated: missing summary trailer",
			},
			{
				name:    "summary count mismatch",
				content: semaphoreJSONLHeader + "\n" + semaphoreJSONLPassed + "\n" + `{"kind":"summary","tests":5,"complete":true}` + "\n",
				summary: parser.Summary{Total: 1, Passed: 1},
				message: "truncated: summary reported 5 tests, parsed 1",
			},
			{
				name:    "unknown kinds skipped",
				content: semaphoreJSONLHeader + "\n" + `{"kind":"checkpoint","at":"2026-08-28T09:12:42Z"}` + "\n" + semaphoreJSONLPassed + "\n" + `{"kind":"summary","tests":1,"complete":true}` + "\n",
				summary: parser.Summary{Total: 1, Passed: 1},
				message: "",
			},
			{
				name:    "output records skipped",
				content: semaphoreJSONLHeader + "\n" + semaphoreJSONLFailed + "\n" + `{"kind":"output","test_ref":"t3","attempt":0,"channel":"stdout","seq":0,"body":"nope\n"}` + "\n" + `{"kind":"summary","tests":1,"complete":true}` + "\n",
				summary: parser.Summary{Total: 1, Failed: 1},
				message: "",
			},
			{
				name:    "unparsable lines skipped",
				content: semaphoreJSONLHeader + "\n" + "not json at all\n\n" + semaphoreJSONLSkipped + "\n" + `{"kind":"summary","tests":1,"complete":true}` + "\n",
				summary: parser.Summary{Total: 1, Skipped: 1},
				message: "",
			},
			{
				name:    "attempts become tests",
				content: semaphoreJSONLHeader + "\n" + semaphoreJSONLRetried + "\n" + `{"kind":"summary","tests":1,"complete":true}` + "\n",
				summary: parser.Summary{Total: 2, Passed: 1, Error: 1},
				message: "",
			},
			{
				name:    "concatenated reports",
				content: semaphoreJSONLHeader + "\n" + semaphoreJSONLPassed + "\n" + `{"kind":"summary","tests":1,"complete":true}` + "\n" + semaphoreJSONLHeader + "\n" + semaphoreJSONLFailed + "\n" + `{"kind":"summary","tests":1,"complete":true}` + "\n",
				summary: parser.Summary{Total: 2, Passed: 1, Failed: 1},
				message: "",
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				path := CreateTempFile(t, "records-*.jsonl", test.content)
				results := NewSemaphoreJSONL().Parse(path)

				if results.Status != parser.StatusSuccess {
					t.Errorf("Status: got %s, want %s", results.Status, parser.StatusSuccess)
				}

				if results.StatusMessage != test.message {
					t.Errorf("StatusMessage: got %q, want %q", results.StatusMessage, test.message)
				}

				AssertTestSummary(t, results.Summary, test.summary)
			})
		}
	})

	t.Run("Envelope", func(t *testing.T) {
		path := CreateTempFile(t, "envelope-*.jsonl", semaphoreJSONLHeader+"\n"+semaphoreJSONLPassed+"\n")
		results := NewSemaphoreJSONL().Parse(path)

		if results.Framework != "rspec" {
			t.Errorf("Framework: got %s, want rspec", results.Framework)
		}

		if results.Name != "Rspec Suite" {
			t.Errorf("Name: got %s, want Rspec Suite", results.Name)
		}

		if len(results.Suites) != 1 {
			t.Fatalf("Suites: got %d, want 1", len(results.Suites))
		}

		suite := results.Suites[0]
		if suite.Name != "spec/calculator/adder_spec.rb" {
			t.Errorf("Suite name: got %s, want spec/calculator/adder_spec.rb", suite.Name)
		}

		if suite.Hostname != "job-42a1" {
			t.Errorf("Suite hostname: got %s, want job-42a1", suite.Hostname)
		}

		if suite.Timestamp != "2026-08-28T09:12:41.676Z" {
			t.Errorf("Suite timestamp: got %s, want 2026-08-28T09:12:41.676Z", suite.Timestamp)
		}

		if suite.Properties["seed"] != "48163" {
			t.Errorf("Suite seed: got %s, want 48163", suite.Properties["seed"])
		}
	})

	t.Run("Attempts Ordering", func(t *testing.T) {
		path := CreateTempFile(t, "attempts-*.jsonl", semaphoreJSONLHeader+"\n"+semaphoreJSONLRetried+"\n")
		results := NewSemaphoreJSONL().Parse(path)

		if len(results.Suites) != 1 {
			t.Fatalf("Suites: got %d, want 1", len(results.Suites))
		}

		tests := results.Suites[0].Tests
		if len(tests) != 2 {
			t.Fatalf("Tests: got %d, want 2", len(tests))
		}

		if tests[0].State != parser.StateError {
			t.Errorf("First attempt state: got %s, want %s", tests[0].State, parser.StateError)
		}

		if tests[0].Error == nil || tests[0].Error.Type != "RuntimeError" {
			t.Errorf("First attempt error: got %v, want RuntimeError", tests[0].Error)
		}

		if tests[1].State != parser.StatePassed {
			t.Errorf("Final attempt state: got %s, want %s", tests[1].State, parser.StatePassed)
		}

		if tests[0].ID != tests[1].ID {
			t.Errorf("Attempts of one test must share an ID: %s != %s", tests[0].ID, tests[1].ID)
		}
	})

	t.Run("Oversized Lines", func(t *testing.T) {
		t.Run("under the cap", func(t *testing.T) {
			message := strings.Repeat("x", 256*1024)
			record := fmt.Sprintf(`{"kind":"test","identity":{"name":"big","file":"spec/big_spec.rb"},"state":"failed","tags":{},"attempts":[{"index":0,"state":"failed","failure":{"message":%q}}]}`, message)
			path := CreateTempFile(t, "big-*.jsonl", semaphoreJSONLHeader+"\n"+record+"\n"+`{"kind":"summary","tests":1,"complete":true}`+"\n")

			results := NewSemaphoreJSONL().Parse(path)
			AssertTestSummary(t, results.Summary, parser.Summary{Total: 1, Failed: 1})

			if results.StatusMessage != "" {
				t.Errorf("StatusMessage: got %q, want empty", results.StatusMessage)
			}

			if len(results.Suites[0].Tests[0].Failure.Message) != len(message) {
				t.Errorf("Failure message was cut: got %d bytes, want %d", len(results.Suites[0].Tests[0].Failure.Message), len(message))
			}
		})

		t.Run("over the cap", func(t *testing.T) {
			message := strings.Repeat("x", semaphoreJSONLMaxLine+1)
			record := fmt.Sprintf(`{"kind":"test","identity":{"name":"huge","file":"spec/huge_spec.rb"},"state":"failed","tags":{},"attempts":[{"index":0,"state":"failed","failure":{"message":%q}}]}`, message)
			path := CreateTempFile(t, "huge-*.jsonl", semaphoreJSONLHeader+"\n"+record+"\n"+semaphoreJSONLPassed+"\n"+`{"kind":"summary","tests":2,"complete":true}`+"\n")

			results := NewSemaphoreJSONL().Parse(path)

			if results.Status != parser.StatusSuccess {
				t.Errorf("Status: got %s, want %s", results.Status, parser.StatusSuccess)
			}

			AssertTestSummary(t, results.Summary, parser.Summary{Total: 1, Passed: 1})

			if !strings.Contains(results.StatusMessage, "skipped 1 lines over") {
				t.Errorf("StatusMessage: got %q, want a skipped lines note", results.StatusMessage)
			}

			if !strings.Contains(results.StatusMessage, "truncated: summary reported 2 tests, parsed 1") {
				t.Errorf("StatusMessage: got %q, want a count mismatch note", results.StatusMessage)
			}
		})
	})
}

// TestSemaphoreJSONLIDContinuity pins the v1 -> v2 transition: the same logical
// rspec suite must produce the same test IDs whether it arrives as JUnit XML or
// as a semaphore test report stream.
func TestSemaphoreJSONLIDContinuity(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<testsuite name="rspec" tests="3" skipped="1" failures="1" errors="0" time="0.012909" timestamp="2026-08-28T09:12:41+02:00" hostname="job-42a1">
  <properties>
    <property name="seed" value="48163" />
  </properties>
  <testcase classname="spec.calculator.adder_spec" name="Calculator::Adder sums arguments" file="./spec/calculator/adder_spec.rb" time="0.000436"></testcase>
  <testcase classname="spec.calculator.adder_spec" name="Calculator::Adder subtracts arguments" file="./spec/calculator/adder_spec.rb" time="0.012473">
    <failure message="expected: -1&#10;     got: 3" type="RSpec::Expectations::ExpectationNotMetError">Failure/Error: expect(result).to eq(-1)</failure>
  </testcase>
  <testcase classname="spec.calculator.divider_spec" name="Calculator::Divider is pending" file="./spec/calculator/divider_spec.rb" time="0.0">
    <skipped/>
  </testcase>
</testsuite>`

	jsonl := strings.Join([]string{
		semaphoreJSONLHeader,
		semaphoreJSONLPassed,
		semaphoreJSONLFailed,
		semaphoreJSONLSkipped,
		`{"kind":"summary","tests":3,"complete":true,"duration_ms":12.909}`,
	}, "\n") + "\n"

	fromXML := NewJUnitRSpec().Parse(CreateTempFile(t, "continuity-*.xml", xml))
	fromJSONL := NewSemaphoreJSONL().Parse(CreateTempFile(t, "continuity-*.jsonl", jsonl))

	if fromXML.ID != fromJSONL.ID {
		t.Errorf("TestResults ID: xml %s, jsonl %s", fromXML.ID, fromJSONL.ID)
	}

	if diff := semaphoreJSONLDiffIDs(semaphoreJSONLSuiteIDs(fromXML), semaphoreJSONLSuiteIDs(fromJSONL)); diff != "" {
		t.Errorf("Suite IDs differ: %s", diff)
	}

	xmlIDs := semaphoreJSONLTestIDs(fromXML)
	jsonlIDs := semaphoreJSONLTestIDs(fromJSONL)

	if len(xmlIDs) != 3 {
		t.Fatalf("XML test IDs: got %d, want 3", len(xmlIDs))
	}

	if diff := semaphoreJSONLDiffIDs(xmlIDs, jsonlIDs); diff != "" {
		t.Errorf("Test IDs differ: %s", diff)
	}
}

// TestSemaphoreJSONLIDContinuityWithoutFile covers runners that don't report a
// file: the suites are grouped differently, the test IDs still have to match.
func TestSemaphoreJSONLIDContinuityWithoutFile(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<testsuite name="rspec" tests="1" skipped="0" failures="0" errors="0" time="0.000436">
  <testcase classname="spec.calculator.adder_spec" name="Calculator::Adder sums arguments" time="0.000436"></testcase>
</testsuite>`

	jsonl := semaphoreJSONLHeader + "\n" +
		`{"kind":"test","identity":{"name":"Calculator::Adder sums arguments","classname":"spec.calculator.adder_spec"},"state":"passed","duration_ms":0.436,"tags":{},"attempts":[{"index":0,"state":"passed","duration_ms":0.436}]}` + "\n" +
		`{"kind":"summary","tests":1,"complete":true}` + "\n"

	fromXML := NewJUnitRSpec().Parse(CreateTempFile(t, "continuity-nofile-*.xml", xml))
	fromJSONL := NewSemaphoreJSONL().Parse(CreateTempFile(t, "continuity-nofile-*.jsonl", jsonl))

	if diff := semaphoreJSONLDiffIDs(semaphoreJSONLTestIDs(fromXML), semaphoreJSONLTestIDs(fromJSONL)); diff != "" {
		t.Errorf("Test IDs differ: %s", diff)
	}

	if fromXML.Suites[0].Name != "rspec" {
		t.Errorf("XML suite name: got %s, want rspec", fromXML.Suites[0].Name)
	}

	if fromJSONL.Suites[0].Name != "spec.calculator.adder_spec" {
		t.Errorf("JSONL suite name: got %s, want spec.calculator.adder_spec", fromJSONL.Suites[0].Name)
	}
}

func semaphoreJSONLSuiteIDs(results parser.TestResults) []string {
	ids := []string{}
	for _, suite := range results.Suites {
		ids = append(ids, suite.ID)
	}
	sort.Strings(ids)

	return ids
}

func semaphoreJSONLTestIDs(results parser.TestResults) []string {
	ids := []string{}
	for _, suite := range results.Suites {
		for _, test := range suite.Tests {
			ids = append(ids, test.ID)
		}
	}
	sort.Strings(ids)

	return ids
}

func semaphoreJSONLDiffIDs(want []string, got []string) string {
	if len(want) != len(got) {
		return fmt.Sprintf("want %v, got %v", want, got)
	}

	for i := range want {
		if want[i] != got[i] {
			return fmt.Sprintf("want %v, got %v", want, got)
		}
	}

	return ""
}
