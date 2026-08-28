package parser

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestWriteJSONLFoldsAttempts(t *testing.T) {
	results := NewTestResults()
	results.Name = "golang suite"
	results.Framework = "golang"
	results.EnsureID()

	suite := NewSuite()
	suite.Name = "example.com/pkg"
	suite.EnsureID(results)

	failed := NewTest()
	failed.Name = "TestFlaky"
	failed.Classname = "example.com/pkg"
	failed.State = StateFailed
	failed.Failure = &Failure{Type: "assert", Message: "boom"}
	failed.EnsureID(suite)

	passed := NewTest()
	passed.Name = "TestFlaky"
	passed.Classname = "example.com/pkg"
	passed.State = StatePassed
	passed.EnsureID(suite)

	other := NewTest()
	other.Name = "TestOther"
	other.Classname = "example.com/pkg"
	other.State = StateDisabled
	other.EnsureID(suite)

	suite.Tests = []Test{failed, passed, other}
	results.Suites = append(results.Suites, suite)

	payload, err := WriteJSONL(&results, "test reporter")
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(string(payload)), "\n")
	if len(lines) != 4 {
		t.Fatalf("expected header + 2 tests + trailer, got %d lines", len(lines))
	}

	var header map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &header); err != nil {
		t.Fatal(err)
	}
	if header["kind"] != "report" || header["schema"] != "semaphore/test-report" || header["framework"] != "golang" {
		t.Fatalf("bad header: %v", header)
	}

	var flaky map[string]any
	if err := json.Unmarshal([]byte(lines[1]), &flaky); err != nil {
		t.Fatal(err)
	}
	attempts := flaky["attempts"].([]any)
	if len(attempts) != 2 {
		t.Fatalf("expected folded attempts, got %d", len(attempts))
	}
	first := attempts[0].(map[string]any)
	if first["state"] != "failed" || first["kind"] != "report_repeat" || first["failure"] == nil {
		t.Fatalf("bad first attempt: %v", first)
	}
	if flaky["state"] != "passed" {
		t.Fatalf("rollup should be final attempt state, got %v", flaky["state"])
	}

	var disabled map[string]any
	if err := json.Unmarshal([]byte(lines[2]), &disabled); err != nil {
		t.Fatal(err)
	}
	if disabled["state"] != "skipped" {
		t.Fatalf("disabled should fold into skipped, got %v", disabled["state"])
	}

	var trailer map[string]any
	if err := json.Unmarshal([]byte(lines[3]), &trailer); err != nil {
		t.Fatal(err)
	}
	if trailer["kind"] != "summary" || trailer["tests"].(float64) != 2 || trailer["complete"] != true {
		t.Fatalf("bad trailer: %v", trailer)
	}
}
