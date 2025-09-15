package parsers

import (
	"testing"
)

func TestGoStaticcheck(t *testing.T) {
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
			Name:       "GoStaticcheck",
			Parser:     NewGoStaticcheck(),
			InputFile:  FixturePath("priv/parsers/go_staticcheck/in.json"),
			GoldenFile: FixturePath("priv/parsers/go_staticcheck/out.json"),
		}
		RunGoldenTest(t, test)
	})

	t.Run("Parser Identification", func(t *testing.T) {
		p := NewGoStaticcheck()

		// Should identify ExUnit test files
		AssertParserApplicable(t, p, FixturePath("priv/parsers/go_staticcheck/in.json"), true)

		// Should reject other parser files
		AssertParserApplicable(t, p, FixturePath("priv/parsers/go_revive/in.json"), false)
		AssertParserApplicable(t, p, FixturePath("priv/parsers/junit_generic/in.xml"), false)
	})

	t.Run("Extensions", func(t *testing.T) {
		p := NewGoStaticcheck()
		AssertParserSupportsExtension(t, p, ".json", true)
		AssertParserSupportsExtension(t, p, ".xml", false)
	})
}
