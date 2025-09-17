package parsers

import (
	"testing"
)

func TestJUnitGoLang(t *testing.T) {
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
			Name:       "JUnitGoLang",
			Parser:     NewJUnitGoLang(),
			InputFile:  FixturePath("priv/parsers/junit_golang/in.xml"),
			GoldenFile: FixturePath("priv/parsers/junit_golang/out.json"),
		}
		RunGoldenTest(t, test)
	})
	
	t.Run("Parser Identification", func(t *testing.T) {
		p := NewJUnitGoLang()
		
		// Should identify Go test files
		AssertParserApplicable(t, p, FixturePath("priv/parsers/junit_golang/in.xml"), true)
		
		// Should reject other parser files
		AssertParserApplicable(t, p, FixturePath("priv/parsers/junit_rspec/in.xml"), false)
		AssertParserApplicable(t, p, FixturePath("priv/parsers/junit_generic/in.xml"), false)
	})
	
	t.Run("Extensions", func(t *testing.T) {
		p := NewJUnitGoLang()
		AssertParserSupportsExtension(t, p, ".xml", true)
		AssertParserSupportsExtension(t, p, ".json", false)
	})
}