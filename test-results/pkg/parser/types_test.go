package parser

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/semaphoreci/toolbox/test-results/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func Test_Result_Combine(t *testing.T) {
	result := NewResult()

	resultToMerge := NewResult()

	testResult := NewTestResults()
	suite := newSuite("1", "foo")
	newTest(&suite, "1", "foo.1")
	newTest(&suite, "2", "foo.2")
	testResult.Suites = append(testResult.Suites, suite)
	suite = newSuite("1", "foo")
	newTest(&suite, "3", "foo.3")
	newTest(&suite, "4", "foo.4")
	testResult.Suites = append(testResult.Suites, suite)
	resultToMerge.TestResults = append(resultToMerge.TestResults, testResult)
	result.Combine(resultToMerge)

	resultToMerge = NewResult()
	suite = newSuite("1", "foo")
	newTest(&suite, "3", "foo.3")
	newTest(&suite, "4", "foo.4")
	testResult = NewTestResults()
	testResult.Suites = append(testResult.Suites, suite)
	resultToMerge.TestResults = append(resultToMerge.TestResults, testResult)
	result.Combine(resultToMerge)

	resultToMerge = NewResult()
	suite = newSuite("1", "foo")
	newTest(&suite, "5", "foo.51")
	newTest(&suite, "6", "foo.61")
	testResult = NewTestResults()
	testResult.Suites = append(testResult.Suites, suite)
	resultToMerge.TestResults = append(resultToMerge.TestResults, testResult)
	result.Combine(resultToMerge)

	for _, suite := range result.TestResults[0].Suites {
		logger.Info("%+v\n", suite)
	}

	assert.Equal(t, 1, len(result.TestResults))
	assert.Equal(t, 1, len(result.TestResults[0].Suites))
	assert.Equal(t, 6, len(result.TestResults[0].Suites[0].Tests))
}

func Test_TestResults_Combine(t *testing.T) {
	suite := newSuite("1", "foo")
	newTest(&suite, "1", "foo.1")
	newTest(&suite, "2", "foo.2")
	newTest(&suite, "3", "foo.3")

	testResult := NewTestResults()
	testResult.Suites = append(testResult.Suites, suite)

	suite = newSuite("1", "foo")
	newTest(&suite, "1", "foo.1")
	newTest(&suite, "2", "foo.2")

	testResultToMerge := NewTestResults()
	testResultToMerge.Suites = append(testResultToMerge.Suites, suite)

	testResult.Combine(testResultToMerge)

	assert.Equal(t, 1, len(testResult.Suites))
	assert.Equal(t, 3, len(testResult.Suites[0].Tests))
}

func Test_Suite_Combine(t *testing.T) {
	suite := newSuite("1", "foo")
	newTest(&suite, "1", "foo.1")
	newTest(&suite, "2", "foo.2")
	newTest(&suite, "3", "foo.3")
	suiteToMerge := newSuite("1", "foo")
	newTest(&suiteToMerge, "1", "foo.1")
	newTest(&suiteToMerge, "2", "foo.2")

	suite.Combine(suiteToMerge)

	assert.Equal(t, 3, len(suite.Tests), "Should combine tests properly given test name and test file")

	test := NewTest()
	test.ID = "4"
	test.Name = "4.foo.4"
	test.File = "foo.4"
	test.State = StateSkipped

	suiteToMerge = newSuite("1", "foo")
	suiteToMerge.AppendTest(test)
	suite.Combine(suiteToMerge)

	assert.Equal(t, 4, len(suite.Tests))
	assert.Equal(t, true, suite.Tests[len(suite.Tests)-1].State == StateSkipped)

	test.State = StatePassed
	suiteToMerge.AppendTest(test)
	suite.Combine(suiteToMerge)
	assert.Equal(t, 4, len(suite.Tests))
	assert.Equal(t, true, suite.Tests[len(suite.Tests)-1].State == StatePassed, "If tests are the same, passed state should take priority over skipped state")

	test.State = StateFailed
	suiteToMerge.AppendTest(test)
	suite.Combine(suiteToMerge)
	assert.Equal(t, 4, len(suite.Tests))
	assert.Equal(t, true, suite.Tests[len(suite.Tests)-1].State == StateFailed, "If tests are the same, failed state should take priority over passed state")
}

func Test_NewTest_Results(t *testing.T) {
	testResults := NewTestResults()

	assert.IsType(t, TestResults{}, testResults)
	assert.Equal(t, StatusSuccess, testResults.Status)
	assert.Equal(t, "", testResults.StatusMessage)
}

func Test_TestResults_Aggregate(t *testing.T) {
	testResults := TestResults{}

	testResults.Aggregate()
	assert.Equal(t, testResults.Summary, Summary{})

	suite := NewSuite()
	suite.Summary.Total = 6
	suite.Summary.Passed = 1
	suite.Summary.Skipped = 2
	suite.Summary.Error = 2
	suite.Summary.Failed = 1
	suite.Summary.Disabled = 1
	suite.Summary.Duration = time.Duration(1)

	testResults.Suites = append(testResults.Suites, suite)

	testResults.Aggregate()
	assert.Equal(t, testResults.Summary, Summary{6, 1, 2, 2, 1, 1, 1})

	suite = NewSuite()
	suite.Summary.Total = 12
	suite.Summary.Passed = 2
	suite.Summary.Skipped = 4
	suite.Summary.Error = 2
	suite.Summary.Failed = 2
	suite.Summary.Disabled = 2
	suite.Summary.Duration = time.Duration(10)

	testResults.Suites = append(testResults.Suites, suite)
	testResults.Aggregate()

	assert.Equal(t, testResults.Summary, Summary{18, 3, 6, 4, 3, 3, 11})
}

func Test_TestResults_ArrangeSuitesByTestFile(t *testing.T) {
	testResults := NewTestResults()
	testResults.ID = "1"

	suite := newSuite("1", "test_suite_name/with_special_chars.go")
	newTest(&suite, "1", "foo/foo.go")
	newTest(&suite, "2", "foo/bar.go")
	testResults.Suites = append(testResults.Suites, suite)

	suite = newSuite("2", "golang")
	newTest(&suite, "3", "golang/foo.go")
	newTest(&suite, "4", "golang/bar.go")
	testResults.Suites = append(testResults.Suites, suite)

	suite = newSuite("3", "foo/foo.go")
	newTest(&suite, "5", "foo/foo.go")
	newTest(&suite, "6", "foo/foo.go")
	testResults.Suites = append(testResults.Suites, suite)

	assert.Equal(t, 3, len(testResults.Suites), "test results should have correct number of suites before arrangement")
	for _, suite := range testResults.Suites {
		assert.Equal(t, 2, len(suite.Tests), "suites should have correct number of tests before arrangement")
	}

	testResults.Aggregate()

	testResults.ArrangeSuitesByTestFile()

	assert.Equal(t, 4, len(testResults.Suites), "test results should have correct number of suites after arrangement")

	suite = testResults.Suites[0]
	assert.Equal(t, "foo/foo.go", suite.Name, "suite name should match")
	assert.Equal(t, 3, len(suite.Tests), "should contain correct number of tests")
	assert.Equal(t, "foo/foo.go#1", suite.Tests[0].Name, "test name should match")
	assert.Equal(t, "foo/foo.go#5", suite.Tests[1].Name, "test name should match")
	assert.Equal(t, "foo/foo.go#6", suite.Tests[2].Name, "test name should match")

	suite = testResults.Suites[1]
	assert.Equal(t, "foo/bar.go", suite.Name, "suite name should match")
	assert.Equal(t, 1, len(suite.Tests), "should remove tests from old suite")
	assert.Equal(t, "foo/bar.go#2", suite.Tests[0].Name, "test name should match")

	suite = testResults.Suites[2]
	assert.Equal(t, "golang/foo.go", suite.Name, "suite name should match")
	assert.Equal(t, 1, len(suite.Tests), "should remove tests from old suite")
	assert.Equal(t, "golang/foo.go#3", suite.Tests[0].Name, "test name should match")

	suite = testResults.Suites[3]
	assert.Equal(t, "golang/bar.go", suite.Name, "suite name should match")
	assert.Equal(t, 1, len(suite.Tests), "should remove tests from old suite")
	assert.Equal(t, "golang/bar.go#4", suite.Tests[0].Name, "test name should match")
}

func Test_TestResults_ArrangeSuitesByTestFile_SingleSuite(t *testing.T) {
	testResults := NewTestResults()
	testResults.ID = "1"

	suite := newSuite("1", "test_suite_name/with_special_chars.go")
	newTest(&suite, "1", "foo/foo.go")
	newTest(&suite, "2", "foo/bar.go")
	test := NewTest()
	test.ID = "3"
	test.Name = "Foo"
	suite.Tests = append(suite.Tests, test)
	test.ID = "4"
	test.Name = "Bar"
	suite.Tests = append(suite.Tests, test)

	testResults.Suites = append(testResults.Suites, suite)

	assert.Equal(t, 1, len(testResults.Suites), "test results should have correct number of suites before arrangement")
	for _, suite := range testResults.Suites {
		assert.Equal(t, 4, len(suite.Tests), "suites should have correct number of tests before arrangement")
	}

	testResults.ArrangeSuitesByTestFile()

	assert.Equal(t, 3, len(testResults.Suites), "test results should have correct number of suites after arrangement")

	suite = testResults.Suites[0]
	assert.Equal(t, "foo/foo.go", suite.Name, "suite name should match")
	assert.Equal(t, 1, len(suite.Tests), "should contain correct number of tests")
	assert.Equal(t, "foo/foo.go#1", suite.Tests[0].Name, "test name should match")

	suite = testResults.Suites[1]
	assert.Equal(t, "foo/bar.go", suite.Name, "suite name should match")
	assert.Equal(t, 1, len(suite.Tests), "should remove tests from old suite")
	assert.Equal(t, "foo/bar.go#2", suite.Tests[0].Name, "test name should match")

	suite = testResults.Suites[2]
	assert.Equal(t, "test_suite_name/with_special_chars.go", suite.Name, "suite name should match")
	assert.Equal(t, 2, len(suite.Tests), "should remove tests from old suite")
	assert.Equal(t, "Foo", suite.Tests[0].Name, "test name should match")
	assert.Equal(t, "Bar", suite.Tests[1].Name, "test name should match")

}

func Test_NewSuite(t *testing.T) {
	suite := NewSuite()

	assert.IsType(t, suite, Suite{})
}

func Test_Suite_Aggregate(t *testing.T) {
	suite := NewSuite()

	suite.Aggregate()
	assert.Equal(t, Summary{}, suite.Summary)

	test := NewTest()
	suite.Tests = append(suite.Tests, test)
	suite.Aggregate()

	assert.Equal(t, Summary{Total: 1, Passed: 1}, suite.Summary)

	test = NewTest()
	test.State = StateFailed
	suite.Tests = append(suite.Tests, test)
	suite.Aggregate()

	assert.Equal(t, Summary{Total: 2, Passed: 1, Failed: 1}, suite.Summary)

	test = NewTest()
	test.State = StateSkipped
	test.Duration = time.Duration(10)
	suite.Tests = append(suite.Tests, test)
	suite.Aggregate()

	assert.Equal(t, Summary{Total: 3, Passed: 1, Failed: 1, Skipped: 1, Duration: 10}, suite.Summary)

	test = NewTest()
	test.State = StateError
	test.Duration = time.Duration(50)
	suite.Tests = append(suite.Tests, test)
	suite.Aggregate()

	assert.Equal(t, Summary{Total: 4, Passed: 1, Failed: 1, Skipped: 1, Error: 1, Duration: 60}, suite.Summary)

	test = NewTest()
	test.State = StateDisabled
	test.Duration = time.Duration(50)
	suite.Tests = append(suite.Tests, test)
	suite.Aggregate()

	assert.Equal(t, Summary{Total: 5, Passed: 1, Failed: 1, Skipped: 1, Error: 1, Disabled: 1, Duration: 110}, suite.Summary, "should not sum up tests duration if a suite duration is present")

	suite.Summary.Duration = 0
	suite.Aggregate()
	assert.Equal(t, Summary{Total: 5, Passed: 1, Failed: 1, Skipped: 1, Error: 1, Disabled: 1, Duration: 110}, suite.Summary, "should sum up tests duration when there is no suite duration present")

	suite.Summary.Duration = -100
	suite.Aggregate()
	assert.Equal(t, Summary{Total: 5, Passed: 1, Failed: 1, Skipped: 1, Error: 1, Disabled: 1, Duration: 110}, suite.Summary, "should sum up tests duration when there duration is invalid")
}

func Test_Summary_Merge(t *testing.T) {
	summary1 := Summary{Total: 10, Passed: 6, Failed: 1, Skipped: 1, Error: 1, Disabled: 1, Duration: 10}
	summary2 := Summary{Total: 20, Passed: 1, Failed: 16, Skipped: 1, Error: 1, Disabled: 1, Duration: 100}
	summary3 := Summary{Total: 15, Passed: 10, Failed: 2, Skipped: 1, Error: 1, Disabled: 1, Duration: 10}
	summary4 := Summary{Total: 25, Passed: 2, Failed: 1, Skipped: 20, Error: 1, Disabled: 1, Duration: 105}

	result := Summary{}
	for _, s := range []Summary{summary1, summary2, summary3, summary4} {
		result.Merge(&s)
	}

	assert.Equal(t, Summary{
		Total:    70,
		Passed:   19,
		Skipped:  23,
		Error:    4,
		Failed:   20,
		Disabled: 4,
		Duration: 225,
	}, result)

	result = Summary{}
	result.Merge(&Summary{})
	assert.Equal(t, Summary{}, result, "empty summaries should be zeroed")

}

func Test_NewTest(t *testing.T) {
	t.Setenv("IP", "192.168.0.1")
	t.Setenv("SEMAPHORE_PIPELINE_ID", "1")
	t.Setenv("SEMAPHORE_WORKFLOW_ID", "2")
	t.Setenv("SEMAPHORE_JOB_NAME", "Test job")
	t.Setenv("SEMAPHORE_JOB_ID", "3")
	t.Setenv("SEMAPHORE_PROJECT_ID", "123")
	t.Setenv("SEMAPHORE_AGENT_MACHINE_TYPE", "t1-awsm")
	t.Setenv("SEMAPHORE_AGENT_MACHINE_OS_IMAGE", "w95")
	t.Setenv("SEMAPHORE_JOB_CREATION_TIME", "1693481195")

	// For branch
	t.Setenv("SEMAPHORE_GIT_REF_TYPE", "branch")
	t.Setenv("SEMAPHORE_GIT_BRANCH", "awsm-w95")
	t.Setenv("SEMAPHORE_GIT_SHA", "1234567890abcdef")
	branchTest := NewTest()

	// For PR
	t.Setenv("SEMAPHORE_GIT_REF_TYPE", "pull-request")
	t.Setenv("SEMAPHORE_GIT_PR_BRANCH", "pr-awsm-w95")
	t.Setenv("SEMAPHORE_GIT_PR_SHA", "fedcba0987654321")
	prTest := NewTest()

	// For tag
	t.Setenv("SEMAPHORE_GIT_REF_TYPE", "tag")
	t.Setenv("SEMAPHORE_GIT_TAG", "v1.0.0")
	tagTest := NewTest()

	t.Run("T", func(t *testing.T) {
		// branch
		assert.Equal(t, branchTest.State, StatePassed, "is in passed state by default")
		assert.Equal(t, branchTest.SemEnv, SemEnv{
			PipelineID:   "1",
			ProjectID:    "123",
			WorkflowID:   "2",
			JobID:        "3",
			JobStartedAt: "1693481195",
			JobName:      "Test job",
			AgentType:    "t1-awsm",
			AgentOsImage: "w95",
			GitRefType:   "branch",
			GitRefName:   "awsm-w95",
			GitRefSha:    "1234567890abcdef",
		})

		// pr
		assert.Equal(t, prTest.State, StatePassed, "is in passed state by default")
		assert.Equal(t, prTest.SemEnv, SemEnv{
			PipelineID:   "1",
			ProjectID:    "123",
			WorkflowID:   "2",
			JobID:        "3",
			JobStartedAt: "1693481195",
			JobName:      "Test job",
			AgentType:    "t1-awsm",
			AgentOsImage: "w95",
			GitRefType:   "pull-request",
			GitRefName:   "pr-awsm-w95",
			GitRefSha:    "fedcba0987654321",
		})

		// tag
		assert.Equal(t, tagTest.State, StatePassed, "is in passed state by default")
		assert.Equal(t, tagTest.SemEnv, SemEnv{
			PipelineID:   "1",
			ProjectID:    "123",
			WorkflowID:   "2",
			JobID:        "3",
			JobStartedAt: "1693481195",
			JobName:      "Test job",
			AgentType:    "t1-awsm",
			AgentOsImage: "w95",
			GitRefType:   "tag",
			GitRefName:   "awsm-w95",
			GitRefSha:    "1234567890abcdef",
		})
	})

}

func Test_NewError(t *testing.T) {
	obj := NewError()

	assert.IsType(t, obj, Error{})
}

func Test_NewFailure(t *testing.T) {
	obj := NewFailure()

	assert.IsType(t, obj, Failure{})
}

func Test_EnsureID(t *testing.T) {
	suite := NewSuite()
	suite.ID = uuid.NewString()

	test := NewTest()
	test.Name = "foo"
	test.EnsureID(suite)

	testToCompare := NewTest()
	testToCompare.Name = "foo"
	testToCompare.EnsureID(suite)

	assert.Equal(t, true, test.ID == testToCompare.ID, "should have the same ID in the same suite")

	testToCompare.Classname = "bar"
	testToCompare.EnsureID(suite)
	assert.Equal(t, false, test.ID == testToCompare.ID, "should have different ID in the same suite when classname differs")
}

func newTest(suite *Suite, id string, file string) {
	test := NewTest()
	test.ID = id
	test.File = file
	test.Name = fmt.Sprintf("%s#%s", file, id)
	suite.AppendTest(test)
}

func newSuite(id string, name string) Suite {
	suite := NewSuite()
	suite.ID = id
	suite.Name = name
	return suite
}

func TestTrimTextTo(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		limit    int
		expected string
	}{
		{
			name:     "text shorter than limit",
			input:    "short text",
			limit:    100,
			expected: "short text",
		},
		{
			name:     "text exactly at limit",
			input:    "exactly10!",
			limit:    10,
			expected: "exactly10!",
		},
		{
			name:     "text longer than limit",
			input:    "this is a very long text that needs trimming",
			limit:    10,
			expected: "...[truncated]...\ns trimming",
		},
		{
			name:     "empty string",
			input:    "",
			limit:    10,
			expected: "",
		},
		{
			name:     "unicode text trimming",
			input:    "Hello 世界 this is a test message",
			limit:    10,
			expected: "...[truncated]...\nst message",
		},
		{
			name:     "multiline text trimming",
			input:    "line1\nline2\nline3\nline4\nline5",
			limit:    15,
			expected: "...[truncated]...\nne3\nline4\nline5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TrimTextTo(tt.input, tt.limit)
			assert.Equal(t, tt.expected, result)
		})
	}
}
