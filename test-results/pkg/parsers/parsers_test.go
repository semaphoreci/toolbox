package parsers

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/semaphoreci/toolbox/test-results/pkg/fileloader"
	"github.com/semaphoreci/toolbox/test-results/pkg/parser"
)

type parserTestCase struct {
	Name  string
	Input string
	want  parser.TestResults
}

var commonParserTestCases = map[string]string{
	"empty": ``,
	"basic": `
			<?xml version="1.0"?>
			<testsuite name="foo" id="1234">
				<testcase name="bar">
				</testcase>
				<testcase name="baz">
				</testcase>
				<testcase name="bar">
				</testcase>
			</testsuite>
		`,
	"multi-suite": `
		<?xml version="1.0"?>
		<testsuites name="ff">
			<testsuite name="foo" id="1234">
				<testcase name="bar">
				</testcase>
				<testcase name="baz">
				</testcase>
			</testsuite>
			<testsuite name="1234">
				<testcase name="bar">
				</testcase>
				<testcase name="baz">
				</testcase>
			</testsuite>
			<testsuite id="1234">
				<testcase name="bar">
				</testcase>
				<testcase name="baz">
				</testcase>
			</testsuite>
			<testsuite name="1235">
				<testcase name="bar" file="foo/bar:123">
				</testcase>
				<testcase name="baz" file="foo/baz">
				</testcase>
			</testsuite>
			<testsuite name="diff by classname">
				<testcase name="bar" file="foo/bar" classname="foo">
				</testcase>
				<testcase name="bar" file="foo/bar" classname="bar">
				</testcase>
			</testsuite>
		</testsuites>
		`,
	"invalid-root": `
			<?xml version="1.0"?>
			<nontestsuites name="em">
				<testsuite name="foo" id="1234">
					<testcase name="bar">
					</testcase>
					<testsuite name="zap" id="4321">
						<testcase name="baz">
						</testcase>
					</testsuite>
					<testsuite name="zup" id="54321">
						<testcase name="bar">
						</testcase>
					</testsuite>
				</testsuite>
			</nontestsuites>
			`,
}

func buildParserTestCases(inputs map[string]string, wants map[string]parser.TestResults) []parserTestCase {
	cases := []parserTestCase{}
	for key, input := range inputs {
		want, exists := wants[key]
		if !exists {
			// Handle missing expected results (either ignore or report an error)
			continue
		}
		cases = append(cases, parserTestCase{
			Name:  key,
			Input: input,
			want:  want,
		})
	}
	return cases
}

func runParserTests(t *testing.T, parser parser.Parser, testCases []parserTestCase) {
	t.Setenv("IP", "192.168.0.1")
	t.Setenv("SEMAPHORE_PIPELINE_ID", "ppl-id")
	t.Setenv("SEMAPHORE_WORKFLOW_ID", "wf-id")
	t.Setenv("SEMAPHORE_JOB_NAME", "job-name")
	t.Setenv("SEMAPHORE_JOB_ID", "job-id")
	t.Setenv("SEMAPHORE_PROJECT_ID", "project-id")
	t.Setenv("SEMAPHORE_AGENT_MACHINE_TYPE", "agent-machine-type")
	t.Setenv("SEMAPHORE_AGENT_MACHINE_OS_IMAGE", "agent-machine-os-image")
	t.Setenv("SEMAPHORE_JOB_CREATION_TIME", "job-creation-time")

	// For branch
	t.Setenv("SEMAPHORE_GIT_REF_TYPE", "git-ref-type")
	t.Setenv("SEMAPHORE_GIT_BRANCH", "git-branch")
	t.Setenv("SEMAPHORE_GIT_SHA", "git-sha")

	for _, tc := range testCases {
		xml := bytes.NewReader([]byte(tc.Input))
		path := fileloader.Ensure(xml)
		got := parser.Parse(path)

		if diff := cmp.Diff(tc.want, got); diff != "" {
			t.Errorf("%s parsing failed for \"%s\" case:\n%s", parser.GetName(), tc.Name, diff)
			t.Errorf("%#v\n\n", got)
		}
	}
}
