package parser

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/google/uuid"
)

// Properties maps additional parameters for test suites
type Properties map[string]string

// State indicates state of specific test
type State string

const (
	// StatePassed indicates that test was successful
	StatePassed State = "passed"
	// StateError indicates that test errored due to unexpected behaviour when running test i.e. exception
	StateError State = "error"
	// StateFailed indicates that test failed due to invalid test result
	StateFailed State = "failed"
	// StateSkipped indicates that test was skipped
	StateSkipped State = "skipped"
	// StateDisabled indicates that test was disabled
	StateDisabled State = "disabled"
)

// Status stores information about parsing results
type Status string

const (
	// StatusSuccess indicates that parsing was successful
	StatusSuccess Status = "success"

	// StatusError indicates that parsing failed due to error
	StatusError Status = "error"
)

type Result struct {
	TestResults []TestResults `json:"testResults"`
}

func NewResult() Result {
	return Result{
		TestResults: []TestResults{},
	}
}

// Combine test results that are part of result
func (me *Result) Combine(other Result) {
	for i := range other.TestResults {
		testResult := other.TestResults[i]
		testResult.Flatten()
		foundTestResultsIdx, found := me.hasTestResults(testResult)
		if found {
			me.TestResults[foundTestResultsIdx].Combine(testResult)
			me.TestResults[foundTestResultsIdx].Aggregate()
		} else {
			me.TestResults = append(me.TestResults, testResult)
		}
	}

	sort.SliceStable(me.TestResults, func(i, j int) bool { return me.TestResults[i].ID < me.TestResults[j].ID })

	for i := range me.TestResults {
		me.TestResults[i].Aggregate()
	}
}

// Flatten makes sure we don't have duplicated suites in test results
func (me *TestResults) Flatten() {
	testResults := NewTestResults()

	for i := range me.Suites {
		foundSuiteIdx, found := testResults.hasSuite(me.Suites[i])
		if found {
			testResults.Suites[foundSuiteIdx].Combine(me.Suites[i])
			testResults.Suites[foundSuiteIdx].Aggregate()
		} else {
			testResults.Suites = append(testResults.Suites, me.Suites[i])
		}
	}
	me.Suites = testResults.Suites

	sort.SliceStable(me.Suites, func(i, j int) bool {
		return me.Suites[i].ID < me.Suites[j].ID
	})
}

func (me *Result) hasTestResults(testResults TestResults) (int, bool) {
	for i := range me.TestResults {
		if me.TestResults[i].ID == testResults.ID {
			return i, true
		}
	}
	return -1, false
}

type TestResults struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Framework     string  `json:"framework"`
	IsDisabled    bool    `json:"isDisabled"`
	Summary       Summary `json:"summary"`
	Status        Status  `json:"status"`
	StatusMessage string  `json:"statusMessage"`
	Suites        []Suite `json:"suites"`
}

func NewTestResults() TestResults {
	return TestResults{
		Suites:        []Suite{},
		Status:        StatusSuccess,
		StatusMessage: "",
	}
}

func (me *TestResults) Combine(other TestResults) {
	if me.ID == other.ID {
		for i := range other.Suites {
			foundSuiteIdx, found := me.hasSuite(other.Suites[i])
			if found {
				me.Suites[foundSuiteIdx].Combine(other.Suites[i])
				me.Suites[foundSuiteIdx].Aggregate()
			} else {
				me.Suites = append(me.Suites, other.Suites[i])
			}
		}

		sort.SliceStable(me.Suites, func(i, j int) bool {
			return me.Suites[i].ID < me.Suites[j].ID
		})
	}
}

func (me *TestResults) hasSuite(suite Suite) (int, bool) {
	for i := range me.Suites {
		if me.Suites[i].ID == suite.ID {
			return i, true
		}
	}
	return -1, false
}

func (me *TestResults) ArrangeSuitesByTestFile() {
	newSuites := []Suite{}

	for _, suite := range me.Suites {
		for _, test := range suite.Tests {
			var (
				idx        int
				foundSuite *Suite
			)
			if test.File != "" {
				idx, foundSuite = EnsureSuiteByName(newSuites, test.File)
			} else {
				idx, foundSuite = EnsureSuiteByName(newSuites, suite.Name)
			}

			foundSuite.Tests = append(foundSuite.Tests, test)
			foundSuite.Aggregate()

			if idx == -1 {
				foundSuite.EnsureID(*me)
				newSuites = append(newSuites, *foundSuite)
			}
		}
	}

	me.Suites = newSuites
	me.Aggregate()
}

func EnsureSuiteByName(suites []Suite, name string) (int, *Suite) {
	for i := range suites {
		if suites[i].Name == name {
			return i, &suites[i]
		}
	}
	suite := NewSuite()
	suite.Name = name

	return -1, &suite
}

func (me *TestResults) EnsureID() {
	if me.ID == "" {
		me.ID = me.Name
	}

	if me.Framework != "" {
		me.ID = fmt.Sprintf("%s%s", me.ID, me.Framework)
	}

	me.ID = UUID(uuid.Nil, me.ID).String()
}

func (me *TestResults) RegenerateID() {
	me.ID = ""
	me.EnsureID()
	for suiteIdx := range me.Suites {
		me.Suites[suiteIdx].ID = ""
		me.Suites[suiteIdx].EnsureID(*me)
		for testIdx := range me.Suites[suiteIdx].Tests {
			me.Suites[suiteIdx].Tests[testIdx].ID = ""
			me.Suites[suiteIdx].Tests[testIdx].EnsureID(me.Suites[suiteIdx])
		}
	}
}

// Aggregate all test suite summaries
func (me *TestResults) Aggregate() {
	summary := Summary{}

	for i := range me.Suites {
		summary.Duration += me.Suites[i].Summary.Duration
		summary.Skipped += me.Suites[i].Summary.Skipped
		summary.Error += me.Suites[i].Summary.Error
		summary.Total += me.Suites[i].Summary.Total
		summary.Failed += me.Suites[i].Summary.Failed
		summary.Passed += me.Suites[i].Summary.Passed
		summary.Disabled += me.Suites[i].Summary.Disabled
	}

	me.Summary = summary
}

type Suite struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	IsSkipped  bool       `json:"isSkipped"`
	IsDisabled bool       `json:"isDisabled"`
	Timestamp  string     `json:"timestamp"`
	Hostname   string     `json:"hostname"`
	Package    string     `json:"package"`
	Properties Properties `json:"properties"`
	Summary    Summary    `json:"summary"`
	SystemOut  string     `json:"systemOut"`
	SystemErr  string     `json:"systemErr"`
	Tests      []Test     `json:"tests"`
}

func NewSuite() Suite {
	return Suite{Tests: []Test{}}
}

func (me *Suite) Combine(other Suite) {
	if me.ID == other.ID {
		for i := range other.Tests {
			if !me.hasTest(other.Tests[i]) {
				me.Tests = append(me.Tests, other.Tests[i])
			}

			shouldReplace, indexToReplace := me.shouldReplaceTest(other.Tests[i])

			if shouldReplace && indexToReplace != -1 {
				me.Tests[indexToReplace] = other.Tests[i]
			}

		}

		sort.SliceStable(me.Tests, func(i, j int) bool {
			return me.Tests[i].ID < me.Tests[j].ID
		})
	}
}
func (me *Suite) shouldReplaceTest(test Test) (shouldReplace bool, foundIndex int) {
	foundIndex = -1
	shouldReplace = false
	for i := range me.Tests {
		if me.Tests[i].ID == test.ID {
			foundIndex = i
			break
		}
	}

	if foundIndex == -1 {
		return
	}

	foundTest := me.Tests[foundIndex]

	if foundTest.State == StateSkipped {
		shouldReplace = true
		return
	}
	if foundTest.State == StatePassed && test.State == StateFailed || test.State == StateError {
		shouldReplace = true
		return
	}

	return
}
func (me *Suite) hasTest(test Test) bool {
	for i := range me.Tests {
		if me.Tests[i].ID == test.ID {
			return true
		}
	}
	return false
}

// Aggregate all tests in suite
func (me *Suite) Aggregate() {
	summary := Summary{}

	for _, test := range me.Tests {
		summary.Duration += test.Duration
		summary.Total++
		switch test.State {
		case StateSkipped:
			summary.Skipped++
		case StateFailed:
			summary.Failed++
		case StateError:
			summary.Error++
		case StatePassed:
			summary.Passed++
		case StateDisabled:
			summary.Disabled++
		}
	}

	// If current duration is not zero and current duration is bigger than calculated duration, use it
	if me.Summary.Duration > 0 && me.Summary.Duration > summary.Duration {
		summary.Duration = me.Summary.Duration
	}

	me.Summary = summary
}

func (me *Suite) EnsureID(tr TestResults) {
	if me.ID == "" {
		me.ID = me.Name
	}

	oldID, err := uuid.Parse(tr.ID)
	if err != nil {
		oldID = uuid.Nil
	}

	me.ID = UUID(oldID, me.ID).String()
}

func (me *Suite) AppendTest(test Test) {
	me.Tests = append(me.Tests, test)
	me.Aggregate()
}

type SemEnv struct {
	ProjectID string `json:"projectId"`

	PipelineID string `json:"pipelineId"`
	WorkflowID string `json:"workflowId"`

	JobStartedAt string `json:"pipelineStartedAt"`

	JobName string `json:"jobName"`
	JobID   string `json:"jobId"`

	AgentType    string `json:"agentType"`
	AgentOsImage string `json:"agentOsImage"`

	GitRefType string `json:"gitRefType"`
	GitRefName string `json:"gitRefName"`
	GitRefSha  string `json:"gitRefSha"`
}

func NewSemEnv() SemEnv {
	refName := ""
	refSha := ""
	switch os.Getenv("SEMAPHORE_GIT_REF_TYPE") {
	case "branch":
		refName = os.Getenv("SEMAPHORE_GIT_BRANCH")
		refSha = os.Getenv("SEMAPHORE_GIT_SHA")
	case "tag":
		refName = os.Getenv("SEMAPHORE_GIT_BRANCH")
		refSha = os.Getenv("SEMAPHORE_GIT_SHA")
	case "pull-request":
		refName = os.Getenv("SEMAPHORE_GIT_PR_BRANCH")
		refSha = os.Getenv("SEMAPHORE_GIT_PR_SHA")
	}

	return SemEnv{
		ProjectID:    os.Getenv("SEMAPHORE_PROJECT_ID"),
		PipelineID:   os.Getenv("SEMAPHORE_PIPELINE_ID"),
		JobStartedAt: os.Getenv("SEMAPHORE_JOB_CREATION_TIME"),
		WorkflowID:   os.Getenv("SEMAPHORE_WORKFLOW_ID"),
		JobName:      os.Getenv("SEMAPHORE_JOB_NAME"),
		JobID:        os.Getenv("SEMAPHORE_JOB_ID"),
		AgentType:    os.Getenv("SEMAPHORE_AGENT_MACHINE_TYPE"),
		AgentOsImage: os.Getenv("SEMAPHORE_AGENT_MACHINE_OS_IMAGE"),
		GitRefType:   os.Getenv("SEMAPHORE_GIT_REF_TYPE"),
		GitRefName:   refName,
		GitRefSha:    refSha,
	}
}

type Test struct {
	ID        string        `json:"id"`
	File      string        `json:"file"`
	Classname string        `json:"classname"`
	Package   string        `json:"package"`
	Name      string        `json:"name"`
	Duration  time.Duration `json:"duration"`
	State     State         `json:"state"`
	Failure   *Failure      `json:"failure"`
	Error     *Error        `json:"error"`
	SystemOut string        `json:"systemOut"`
	SystemErr string        `json:"systemErr"`
	SemEnv    SemEnv        `json:"semaphoreEnv"`
}

func NewTest() Test {
	return Test{
		State:  StatePassed,
		SemEnv: NewSemEnv(),
	}
}

func (me *Test) EnsureID(s Suite) {
	// Determine the ID based on the various test details
	testIdentity := fmt.Sprintf("%s.%s.%s.%s.%s", me.ID, me.Name, me.Classname, me.Package, me.File)

	me.ID = UUID(uuid.MustParse(s.ID), testIdentity).String()
}

type err struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Body    string `json:"body"`
}

type Failure err

func NewFailure() Failure {
	return Failure{}
}

type Error err

func NewError() Error {
	return Error{}
}

type Summary struct {
	Total    int           `json:"total"`
	Passed   int           `json:"passed"`
	Skipped  int           `json:"skipped"`
	Error    int           `json:"error"`
	Failed   int           `json:"failed"`
	Disabled int           `json:"disabled"`
	Duration time.Duration `json:"duration"`
}

// Merge merges two summaries together summing each field
func (s *Summary) Merge(withSummary *Summary) {
	s.Total += withSummary.Total
	s.Passed += withSummary.Passed
	s.Skipped += withSummary.Skipped
	s.Error += withSummary.Error
	s.Failed += withSummary.Failed
	s.Disabled += withSummary.Disabled
	s.Duration += withSummary.Duration
}

type TestResult struct {
	TestID   string
	GitSha   string
	Duration time.Duration
	JobID    string
	State    State
}

func (t *TestResult) String() []string {
	return []string{t.TestID, t.GitSha, fmt.Sprintf("%d", t.Duration.Milliseconds()), t.JobID, string(t.State)}
}

func UUID(id uuid.UUID, str string) uuid.UUID {
	return uuid.NewMD5(id, []byte(str))
}

// TrimTextTo trims a string to the last N characters, adding a truncation marker if needed
func TrimTextTo(s string, n int) string {
	if len(s) <= n {
		return s
	}
	// Keep the last N characters
	truncated := s[len(s)-int(n):]
	return "...[truncated]...\n" + truncated
}

// FilterFailedTests returns a new Result containing only failed/errored tests
// while preserving the original summary statistics
func (r *Result) FilterFailedTests() Result {
	filtered := Result{
		TestResults: make([]TestResults, 0, len(r.TestResults)),
	}

	for _, tr := range r.TestResults {
		filteredTR := TestResults{
			ID:            tr.ID,
			Name:          tr.Name,
			Framework:     tr.Framework,
			IsDisabled:    tr.IsDisabled,
			Summary:       tr.Summary, // Preserve original summary
			Status:        tr.Status,
			StatusMessage: tr.StatusMessage,
			Suites:        make([]Suite, 0, len(tr.Suites)),
		}

		for _, suite := range tr.Suites {
			filteredSuite := Suite{
				ID:         suite.ID,
				Name:       suite.Name,
				IsSkipped:  suite.IsSkipped,
				IsDisabled: suite.IsDisabled,
				Timestamp:  suite.Timestamp,
				Hostname:   suite.Hostname,
				Package:    suite.Package,
				Properties: suite.Properties,
				Summary:    suite.Summary, // Preserve original summary
				SystemOut:  suite.SystemOut,
				SystemErr:  suite.SystemErr,
				Tests:      make([]Test, 0),
			}

			// Only include failed or errored tests
			for _, test := range suite.Tests {
				if test.State == StateFailed || test.State == StateError {
					filteredSuite.Tests = append(filteredSuite.Tests, test)
				}
			}

			filteredTR.Suites = append(filteredTR.Suites, filteredSuite)
		}

		filtered.TestResults = append(filtered.TestResults, filteredTR)
	}

	return filtered
}
