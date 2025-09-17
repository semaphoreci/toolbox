package parsers

import (
	"testing"
)

func TestGoRevive(t *testing.T) {
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
			Name:       "GoRevive",
			Parser:     NewGoRevive(),
			InputFile:  FixturePath("priv/parsers/go_revive/in.json"),
			GoldenFile: FixturePath("priv/parsers/go_revive/out.json"),
		}
		RunGoldenTest(t, test)
	})
	
	t.Run("Parser Identification", func(t *testing.T) {
		p := NewGoRevive()
		
		// Should identify revive JSON files
		AssertParserApplicable(t, p, FixturePath("priv/parsers/go_revive/in.json"), true)
		
		// Should reject other JSON files
		AssertParserApplicable(t, p, FixturePath("priv/parsers/go_staticcheck/in.json"), false)
		
		// Should reject XML files
		AssertParserApplicable(t, p, FixturePath("priv/parsers/junit_golang/in.xml"), false)
	})
	
	t.Run("Extensions", func(t *testing.T) {
		p := NewGoRevive()
		AssertParserSupportsExtension(t, p, ".json", true)
		AssertParserSupportsExtension(t, p, ".xml", false)
	})
	
	t.Run("Empty JSON Array", func(t *testing.T) {
		emptyJSON := "[]"
		tmpFile := CreateTempFile(t, "*.json", emptyJSON)
		
		p := NewGoRevive()
		result := p.Parse(tmpFile)
		
		// Should handle empty issue list gracefully
		if result.Status != "success" {
			t.Errorf("Expected success status for empty issues, got %s", result.Status)
		}
		
		// Should have one suite with a passing test
		if len(result.Suites) != 1 {
			t.Fatalf("Expected 1 suite, got %d", len(result.Suites))
		}
		
		if len(result.Suites[0].Tests) != 1 {
			t.Fatalf("Expected 1 test, got %d", len(result.Suites[0].Tests))
		}
		
		if result.Suites[0].Tests[0].State != "passed" {
			t.Errorf("Expected passing test for no issues, got %s", result.Suites[0].Tests[0].State)
		}
	})
	
	t.Run("Invalid JSON", func(t *testing.T) {
		invalidJSON := "not json"
		tmpFile := CreateTempFile(t, "*.json", invalidJSON)
		
		p := NewGoRevive()
		
		// Should not be applicable
		if p.IsApplicable(tmpFile) {
			t.Error("Should not identify invalid JSON as revive output")
		}
		
		// Parse should handle gracefully
		result := p.Parse(tmpFile)
		if result.Status != "error" {
			t.Errorf("Expected error status for invalid JSON, got %s", result.Status)
		}
	})
}