package parsers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/semaphoreci/toolbox/test-results/pkg/parser"
)

// goTestJSONSample is the wrapper's aggregation table, verbatim: one test run
// twice, one panicking test, one skipped test and a package level failure.
const goTestJSONSample = `{"Time":"2026-08-25T10:00:00.000Z","Action":"run","Package":"example.com/pkg","Test":"TestA"}
{"Time":"2026-08-25T10:00:00.100Z","Action":"output","Package":"example.com/pkg","Test":"TestA","Output":"=== RUN   TestA\n"}
{"Time":"2026-08-25T10:00:00.200Z","Action":"output","Package":"example.com/pkg","Test":"TestA","Output":"    a_test.go:12: expected 1 got 2\n"}
{"Time":"2026-08-25T10:00:00.300Z","Action":"fail","Package":"example.com/pkg","Test":"TestA","Elapsed":0.3}
{"Time":"2026-08-25T10:00:00.400Z","Action":"run","Package":"example.com/pkg","Test":"TestA"}
{"Time":"2026-08-25T10:00:00.700Z","Action":"pass","Package":"example.com/pkg","Test":"TestA","Elapsed":0.25}
{"Time":"2026-08-25T10:00:00.800Z","Action":"run","Package":"example.com/pkg","Test":"TestB"}
{"Time":"2026-08-25T10:00:00.900Z","Action":"output","Package":"example.com/pkg","Test":"TestB","Output":"panic: boom [recovered]\n"}
{"Time":"2026-08-25T10:00:01.000Z","Action":"fail","Package":"example.com/pkg","Test":"TestB","Elapsed":0.1}
{"Time":"2026-08-25T10:00:01.100Z","Action":"run","Package":"example.com/pkg","Test":"TestC"}
{"Time":"2026-08-25T10:00:01.200Z","Action":"skip","Package":"example.com/pkg","Test":"TestC","Elapsed":0}
{"Time":"2026-08-25T10:00:01.300Z","Action":"fail","Package":"example.com/pkg","Elapsed":1.3}
`

const goTestJSONPackageFailureStream = `{"Time":"2026-08-25T10:00:00.000Z","Action":"start","Package":"example.com/broken"}
{"Time":"2026-08-25T10:00:00.100Z","Action":"output","Package":"example.com/broken","Output":"# example.com/broken\n"}
{"Time":"2026-08-25T10:00:00.200Z","Action":"output","Package":"example.com/broken","Output":"./broken.go:5:2: undefined: missing\n"}
{"Time":"2026-08-25T10:00:00.300Z","Action":"fail","Package":"example.com/broken","Elapsed":0.3}
`

func goTestJSONParse(t *testing.T, content string) parser.TestResults {
	t.Helper()

	return NewGoTestJSON().Parse(CreateTempFile(t, "go-test-json-*.jsonl", content))
}

func goTestJSONFind(t *testing.T, results parser.TestResults, name string) []parser.Test {
	t.Helper()

	found := []parser.Test{}
	for _, suite := range results.Suites {
		for _, test := range suite.Tests {
			if test.Name == name {
				found = append(found, test)
			}
		}
	}

	if len(found) == 0 {
		t.Fatalf("test %s not found in %+v", name, results.Suites)
	}

	return found
}

func TestGoTestJSON(t *testing.T) {
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
			Name:       "GoTestJSON",
			Parser:     NewGoTestJSON(),
			InputFile:  FixturePath("priv/parsers/go_test_json/in.jsonl"),
			GoldenFile: FixturePath("priv/parsers/go_test_json/out.json"),
		}
		RunGoldenTest(t, test)
	})

	t.Run("Parser Identification", func(t *testing.T) {
		p := NewGoTestJSON()

		AssertParserApplicable(t, p, FixturePath("priv/parsers/go_test_json/in.jsonl"), true)

		AssertParserApplicable(t, p, FixturePath("priv/parsers/semaphore_jsonl/in.jsonl"), false)
		AssertParserApplicable(t, p, FixturePath("priv/parsers/junit_rspec/in.xml"), false)
		AssertParserApplicable(t, p, FixturePath("priv/parsers/go_staticcheck/in.json"), false)
		AssertParserApplicable(t, p, FixturePath("priv/parsers/go_revive/in.json"), false)

		AssertParserApplicable(t, p, CreateTempFile(t, "empty-*.jsonl", ""), false)
		AssertParserApplicable(t, p, CreateTempFile(t, "blank-*.jsonl", "\n\n"), false)
		AssertParserApplicable(t, p, CreateTempFile(t, "not-json-*.jsonl", "hello\n"), false)
		AssertParserApplicable(t, p, CreateTempFile(t, "no-action-*.jsonl", `{"Package":"example.com/pkg"}`+"\n"), false)
		AssertParserApplicable(t, p, CreateTempFile(t, "bad-action-*.jsonl", `{"Action":"bench","Package":"example.com/pkg"}`+"\n"), false)
		AssertParserApplicable(t, p, CreateTempFile(t, "lonely-action-*.jsonl", `{"Action":"run","Test":"TestA"}`+"\n"), false)
		AssertParserApplicable(t, p, CreateTempFile(t, "leading-blank-*.jsonl", "\n"+`{"Action":"start","Package":"example.com/pkg"}`+"\n"), true)
		AssertParserApplicable(t, p, CreateTempFile(t, "time-only-*.jsonl", `{"Time":"2026-08-25T10:00:00.000Z","Action":"run","Test":"TestA"}`+"\n"), true)
	})

	// The two JSON Lines parsers share both extensions, so their sniffing has
	// to stay disjoint: neither may claim the other's stream.
	t.Run("Disjoint From Semaphore JSONL", func(t *testing.T) {
		goTestJSON := FixturePath("priv/parsers/go_test_json/in.jsonl")
		semaphoreJSONL := FixturePath("priv/parsers/semaphore_jsonl/in.jsonl")

		AssertParserApplicable(t, NewGoTestJSON(), goTestJSON, true)
		AssertParserApplicable(t, NewSemaphoreJSONL(), goTestJSON, false)

		AssertParserApplicable(t, NewSemaphoreJSONL(), semaphoreJSONL, true)
		AssertParserApplicable(t, NewGoTestJSON(), semaphoreJSONL, false)
	})

	t.Run("Extensions", func(t *testing.T) {
		p := NewGoTestJSON()
		AssertParserSupportsExtension(t, p, ".jsonl", true)
		AssertParserSupportsExtension(t, p, ".json", true)
		AssertParserSupportsExtension(t, p, ".xml", false)
	})

	t.Run("Registered For Directory Walking", func(t *testing.T) {
		byName, err := FindParser("go-test-json", "")
		if err != nil {
			t.Fatalf("FindParser by name failed: %v", err)
		}

		if byName.GetName() != "go-test-json" {
			t.Errorf("FindParser by name picked %s, want go-test-json", byName.GetName())
		}

		for _, ext := range []string{".jsonl", ".json"} {
			path := CreateTempFile(t, "auto-*"+ext, goTestJSONSample)

			p, err := FindParser("auto", path)
			if err != nil {
				t.Fatalf("FindParser(%s) failed: %v", ext, err)
			}

			if p.GetName() != "go-test-json" {
				t.Errorf("FindParser(%s) picked %s, want go-test-json", ext, p.GetName())
			}
		}
	})

	t.Run("Attempt Aggregation", func(t *testing.T) {
		results := goTestJSONParse(t, goTestJSONSample)
		attempts := goTestJSONFind(t, results, "TestA")

		if len(attempts) != 2 {
			t.Fatalf("Attempts: got %d, want 2", len(attempts))
		}

		if attempts[0].State != parser.StateFailed {
			t.Errorf("First attempt state: got %s, want %s", attempts[0].State, parser.StateFailed)
		}

		if attempts[1].State != parser.StatePassed {
			t.Errorf("Final attempt state: got %s, want %s", attempts[1].State, parser.StatePassed)
		}

		if attempts[0].ID != attempts[1].ID {
			t.Errorf("Attempts of one test must share an ID: %s != %s", attempts[0].ID, attempts[1].ID)
		}

		if attempts[0].Duration != 300*1000*1000 {
			t.Errorf("First attempt duration: got %s, want 300ms", attempts[0].Duration)
		}
	})

	t.Run("Failure Site Stays Out Of Identity", func(t *testing.T) {
		results := goTestJSONParse(t, goTestJSONSample)
		failed := goTestJSONFind(t, results, "TestA")[0]

		if failed.Failure == nil {
			t.Fatal("Failure: got nil, want a failure")
		}

		if failed.Failure.Message != "expected 1 got 2" {
			t.Errorf("Failure message: got %q, want %q", failed.Failure.Message, "expected 1 got 2")
		}

		if !strings.HasPrefix(failed.Failure.Body, "a_test.go:12\n") {
			t.Errorf("Failure body: got %q, want the call site on the first line", failed.Failure.Body)
		}

		if failed.File != "" {
			t.Errorf("Test file: got %q, want empty (the call site must not move test identity)", failed.File)
		}

		if failed.Package != "" {
			t.Errorf("Test package: got %q, want empty", failed.Package)
		}
	})

	t.Run("Panic Taxonomy", func(t *testing.T) {
		results := goTestJSONParse(t, goTestJSONSample)
		panicked := goTestJSONFind(t, results, "TestB")[0]

		if panicked.State != parser.StateError {
			t.Errorf("State: got %s, want %s", panicked.State, parser.StateError)
		}

		if panicked.Failure != nil {
			t.Errorf("Failure: got %+v, want nil (errors carry the detail)", panicked.Failure)
		}

		if panicked.Error == nil {
			t.Fatal("Error: got nil, want a panic error")
		}

		if panicked.Error.Type != "panic" {
			t.Errorf("Error type: got %q, want panic", panicked.Error.Type)
		}

		if panicked.Error.Message != "panic: boom [recovered]" {
			t.Errorf("Error message: got %q, want the panic line", panicked.Error.Message)
		}
	})

	t.Run("Skip State", func(t *testing.T) {
		results := goTestJSONParse(t, goTestJSONSample)
		skipped := goTestJSONFind(t, results, "TestC")[0]

		if skipped.State != parser.StateSkipped {
			t.Errorf("State: got %s, want %s", skipped.State, parser.StateSkipped)
		}
	})

	t.Run("Unterminated Attempt Errors", func(t *testing.T) {
		stream := `{"Time":"2026-08-25T10:00:00.000Z","Action":"run","Package":"example.com/pkg","Test":"TestGone"}
{"Time":"2026-08-25T10:00:00.100Z","Action":"output","Package":"example.com/pkg","Test":"TestGone","Output":"still going\n"}
`
		results := goTestJSONParse(t, stream)
		gone := goTestJSONFind(t, results, "TestGone")[0]

		if gone.State != parser.StateError {
			t.Errorf("State: got %s, want %s", gone.State, parser.StateError)
		}
	})

	t.Run("Package Failure", func(t *testing.T) {
		results := goTestJSONParse(t, goTestJSONPackageFailureStream)

		AssertTestSummary(t, results.Summary, parser.Summary{Total: 1, Error: 1})

		synthetic := goTestJSONFind(t, results, "PackageFailure")[0]

		if synthetic.Classname != "example.com/broken" {
			t.Errorf("Classname: got %q, want example.com/broken", synthetic.Classname)
		}

		if synthetic.State != parser.StateError {
			t.Errorf("State: got %s, want %s", synthetic.State, parser.StateError)
		}

		if synthetic.Error == nil || synthetic.Error.Type != "package_failure" {
			t.Fatalf("Error: got %+v, want a package_failure", synthetic.Error)
		}

		if synthetic.Error.Message != "# example.com/broken" {
			t.Errorf("Error message: got %q, want the compiler header", synthetic.Error.Message)
		}

		if !strings.Contains(synthetic.Error.Body, "undefined: missing") {
			t.Errorf("Error body: got %q, want the compiler output", synthetic.Error.Body)
		}
	})

	t.Run("Package Failure Only Without Tests", func(t *testing.T) {
		results := goTestJSONParse(t, goTestJSONSample)

		for _, suite := range results.Suites {
			for _, test := range suite.Tests {
				if test.Name == "PackageFailure" {
					t.Errorf("PackageFailure: emitted for a package that reported tests")
				}
			}
		}
	})

	t.Run("Summary", func(t *testing.T) {
		tests := []struct {
			name    string
			content string
			summary parser.Summary
		}{
			{
				name:    "empty stream",
				content: "",
				summary: parser.Summary{},
			},
			{
				name:    "no events at all",
				content: "not json\n\nstill not json\n",
				summary: parser.Summary{},
			},
			{
				name:    "no envelope needed",
				content: goTestJSONSample,
				summary: parser.Summary{Total: 4, Passed: 1, Failed: 1, Error: 1, Skipped: 1},
			},
			{
				name:    "package failure",
				content: goTestJSONPackageFailureStream,
				summary: parser.Summary{Total: 1, Error: 1},
			},
			{
				name:    "package with no test files",
				content: `{"Time":"2026-08-25T10:00:00.000Z","Action":"start","Package":"example.com/empty"}` + "\n" + `{"Time":"2026-08-25T10:00:00.100Z","Action":"output","Package":"example.com/empty","Output":"?   \texample.com/empty\t[no test files]\n"}` + "\n" + `{"Time":"2026-08-25T10:00:00.200Z","Action":"skip","Package":"example.com/empty","Elapsed":0}` + "\n",
				summary: parser.Summary{},
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				results := goTestJSONParse(t, test.content)

				if results.Status != parser.StatusSuccess {
					t.Errorf("Status: got %s, want %s", results.Status, parser.StatusSuccess)
				}

				if results.StatusMessage != "" {
					t.Errorf("StatusMessage: got %q, want empty", results.StatusMessage)
				}

				AssertTestSummary(t, results.Summary, test.summary)
			})
		}
	})

	t.Run("Oversized Lines", func(t *testing.T) {
		t.Run("under the cap", func(t *testing.T) {
			output := strings.Repeat("x", 256*1024)
			stream := `{"Action":"run","Package":"example.com/pkg","Test":"TestBig"}` + "\n" +
				fmt.Sprintf(`{"Action":"output","Package":"example.com/pkg","Test":"TestBig","Output":%q}`, output) + "\n" +
				`{"Action":"fail","Package":"example.com/pkg","Test":"TestBig","Elapsed":0.1}` + "\n"

			results := goTestJSONParse(t, stream)
			AssertTestSummary(t, results.Summary, parser.Summary{Total: 1, Failed: 1})

			if results.StatusMessage != "" {
				t.Errorf("StatusMessage: got %q, want empty", results.StatusMessage)
			}

			if len(goTestJSONFind(t, results, "TestBig")[0].Failure.Body) != len(output) {
				t.Error("Failure body was cut")
			}
		})

		t.Run("over the cap", func(t *testing.T) {
			output := strings.Repeat("x", goTestJSONMaxLine+1)
			stream := `{"Action":"run","Package":"example.com/pkg","Test":"TestHuge"}` + "\n" +
				fmt.Sprintf(`{"Action":"output","Package":"example.com/pkg","Test":"TestHuge","Output":%q}`, output) + "\n" +
				`{"Action":"fail","Package":"example.com/pkg","Test":"TestHuge","Elapsed":0.1}` + "\n"

			results := goTestJSONParse(t, stream)

			if results.Status != parser.StatusSuccess {
				t.Errorf("Status: got %s, want %s", results.Status, parser.StatusSuccess)
			}

			AssertTestSummary(t, results.Summary, parser.Summary{Total: 1, Failed: 1})

			if !strings.Contains(results.StatusMessage, "skipped 1 lines over") {
				t.Errorf("StatusMessage: got %q, want a skipped lines note", results.StatusMessage)
			}
		})
	})
}

// TestGoTestJSONIDContinuity pins the carrier swap: a run reported through
// gotestsum's JUnit XML and the same run reported as a go test -json stream
// must produce the same result, suite and test ids.
func TestGoTestJSONIDContinuity(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<testsuites>
  <testsuite tests="3" failures="1" errors="0" time="0.020000" name="example.com/calc">
    <properties>
      <property name="go.version" value="go1.21 darwin/arm64"></property>
    </properties>
    <testcase classname="example.com/calc" name="TestAdd" time="0.020000"></testcase>
    <testcase classname="example.com/calc" name="TestSub" time="0.000000">
      <failure message="Failed" type="">=== RUN   TestSub&#xA;    calc_test.go:17: expected 2 got 4&#xA;--- FAIL: TestSub (0.00s)&#xA;</failure>
    </testcase>
    <testcase classname="example.com/calc" name="TestSkipped" time="0.000000">
      <skipped message="not implemented"></skipped>
    </testcase>
  </testsuite>
</testsuites>`

	fromXML := NewJUnitGoLang().Parse(CreateTempFile(t, "continuity-*.xml", xml))
	fromJSON := NewGoTestJSON().Parse(FixturePath("priv/parsers/go_test_json/in.jsonl"))

	if fromXML.Framework != fromJSON.Framework {
		t.Errorf("Framework: xml %q, json %q", fromXML.Framework, fromJSON.Framework)
	}

	if fromXML.Name != "" || fromJSON.Name != "" {
		t.Errorf("Name: xml %q, json %q, want both empty", fromXML.Name, fromJSON.Name)
	}

	if fromXML.ID != fromJSON.ID {
		t.Errorf("TestResults ID: xml %s, json %s", fromXML.ID, fromJSON.ID)
	}

	if diff := semaphoreJSONLDiffIDs(semaphoreJSONLSuiteIDs(fromXML), semaphoreJSONLSuiteIDs(fromJSON)); diff != "" {
		t.Errorf("Suite IDs differ: %s", diff)
	}

	if fromJSON.Suites[0].Name != "example.com/calc" {
		t.Errorf("Suite name: got %s, want example.com/calc", fromJSON.Suites[0].Name)
	}

	xmlIDs := semaphoreJSONLTestIDs(fromXML)
	jsonIDs := semaphoreJSONLTestIDs(fromJSON)

	if len(xmlIDs) != 3 {
		t.Fatalf("XML test IDs: got %d, want 3", len(xmlIDs))
	}

	if diff := semaphoreJSONLDiffIDs(xmlIDs, jsonIDs); diff != "" {
		t.Errorf("Test IDs differ: %s", diff)
	}
}
