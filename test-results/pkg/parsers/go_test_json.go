package parsers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/semaphoreci/toolbox/test-results/pkg/logger"
	"github.com/semaphoreci/toolbox/test-results/pkg/parser"
)

const (
	// goTestJSONFramework mirrors the JUnit golang parser so that a run keeps
	// the same test ids whether it is carried by gotestsum or by this parser.
	goTestJSONFramework = "golang"

	goTestJSONMaxLine   = 1024 * 1024
	goTestJSONProbeSize = 64 * 1024

	goTestJSONPackageFailureName = "PackageFailure"
	goTestJSONPackageFailureType = "package_failure"
	goTestJSONStackLines         = 30
)

var (
	goTestJSONActions = map[string]bool{
		"start":  true,
		"run":    true,
		"output": true,
		"pass":   true,
		"fail":   true,
		"skip":   true,
		"cont":   true,
		"pause":  true,
	}

	goTestJSONPanicPattern       = regexp.MustCompile(`(?m)^(panic:|fatal error:|runtime error:)`)
	goTestJSONFailureSitePattern = regexp.MustCompile(`(?m)^\s+(\S+\.go):(\d+): ?(.*)`)
)

type goTestJSONEvent struct {
	Time    time.Time `json:"Time"`
	Action  string    `json:"Action"`
	Package string    `json:"Package"`
	Test    string    `json:"Test"`
	Elapsed float64   `json:"Elapsed"`
	Output  string    `json:"Output"`
}

// goTestJSONProbe uses pointers so that a missing field is distinguishable from
// an empty one, which is what separates a test2json event from every other
// JSON Lines format the CLI accepts.
type goTestJSONProbe struct {
	Action  *string `json:"Action"`
	Package *string `json:"Package"`
	Time    *string `json:"Time"`
}

type goTestJSONAttempt struct {
	State     string
	StartedAt time.Time
	Elapsed   float64
	Output    []string
}

type goTestJSONRecord struct {
	Name     string
	Attempts []*goTestJSONAttempt
}

func (me *goTestJSONRecord) current() *goTestJSONAttempt {
	if len(me.Attempts) == 0 {
		return nil
	}

	return me.Attempts[len(me.Attempts)-1]
}

type goTestJSONPackage struct {
	Name    string
	Failed  bool
	Started time.Time
	Output  []string
	Records []*goTestJSONRecord
	byName  map[string]*goTestJSONRecord
}

type goTestJSONRun struct {
	Packages []*goTestJSONPackage
	byName   map[string]*goTestJSONPackage
}

func newGoTestJSONRun() *goTestJSONRun {
	return &goTestJSONRun{byName: map[string]*goTestJSONPackage{}}
}

func (me *goTestJSONRun) apply(event goTestJSONEvent) {
	if event.Package == "" {
		logger.Debug("Skipping event without a package")
		return
	}

	pkg := me.byName[event.Package]
	if pkg == nil {
		pkg = &goTestJSONPackage{
			Name:    event.Package,
			Started: event.Time,
			byName:  map[string]*goTestJSONRecord{},
		}
		me.byName[event.Package] = pkg
		me.Packages = append(me.Packages, pkg)
	}

	if event.Test == "" {
		switch event.Action {
		case "output":
			pkg.Output = append(pkg.Output, event.Output)
		case "fail":
			pkg.Failed = true
		}
		return
	}

	record := pkg.byName[event.Test]
	if record == nil {
		record = &goTestJSONRecord{Name: event.Test}
		pkg.byName[event.Test] = record
		pkg.Records = append(pkg.Records, record)
	}

	switch event.Action {
	case "run":
		record.Attempts = append(record.Attempts, &goTestJSONAttempt{StartedAt: event.Time})
	case "output":
		if current := record.current(); current != nil {
			current.Output = append(current.Output, event.Output)
		}
	case "pass", "fail", "skip":
		if current := record.current(); current != nil && current.State == "" {
			current.State = event.Action
			current.Elapsed = event.Elapsed
		}
	}
}

// GoTestJSON ...
type GoTestJSON struct {
}

// NewGoTestJSON ...
func NewGoTestJSON() GoTestJSON {
	return GoTestJSON{}
}

// GetName ...
func (me GoTestJSON) GetName() string {
	return "go-test-json"
}

// GetDescription ...
func (me GoTestJSON) GetDescription() string {
	return "go test -json event stream"
}

// GetSupportedExtensions ...
func (me GoTestJSON) GetSupportedExtensions() []string {
	return []string{".jsonl", ".json"}
}

// IsApplicable ...
func (me GoTestJSON) IsApplicable(path string) bool {
	logger.Debug("Checking applicability of %s parser", me.GetName())

	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		logger.Debug("Opening file failed: %v", err)
		return false
	}
	defer file.Close()

	reader := bufio.NewReaderSize(file, goTestJSONProbeSize)

	for {
		line, err := reader.ReadSlice('\n')
		if err != nil && err != io.EOF {
			logger.Debug("Reading first line failed: %v", err)
			return false
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			if err == io.EOF {
				return false
			}
			continue
		}

		probe := goTestJSONProbe{}
		if err := json.Unmarshal(line, &probe); err != nil {
			logger.Debug("First line is not a JSON object: %v", err)
			return false
		}

		if probe.Action == nil || !goTestJSONActions[*probe.Action] {
			return false
		}

		return probe.Package != nil || probe.Time != nil
	}
}

// Parse ...
func (me GoTestJSON) Parse(path string) parser.TestResults {
	results := parser.NewTestResults()
	// The name is left empty and the framework is reported as golang on
	// purpose: gotestsum emits a <testsuites> root without a name, and both
	// feed the same id chain. Changing either breaks test id continuity.
	results.Framework = goTestJSONFramework
	results.EnsureID()

	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		logger.Error("Opening file failed: %v", err)
		results.Status = parser.StatusError
		results.StatusMessage = err.Error()
		return results
	}
	defer file.Close()

	run := newGoTestJSONRun()
	oversized := 0

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, goTestJSONMaxLine), goTestJSONMaxLine)
	scanner.Split(jsonlSplit(goTestJSONMaxLine, &oversized))

	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 || line[0] != '{' {
			continue
		}

		event := goTestJSONEvent{}
		if err := json.Unmarshal(line, &event); err != nil {
			logger.Debug("Skipping unparsable line: %v", err)
			continue
		}

		run.apply(event)
	}

	if err := scanner.Err(); err != nil {
		logger.Error("Reading %s failed: %v", path, err)
		results.Status = parser.StatusError
		results.StatusMessage = err.Error()
		return results
	}

	for _, pkg := range run.Packages {
		suite := me.newSuite(pkg, results)
		if len(suite.Tests) == 0 {
			continue
		}

		results.Suites = append(results.Suites, suite)
	}

	if oversized > 0 {
		results.StatusMessage = fmt.Sprintf("skipped %d lines over %d bytes", oversized, goTestJSONMaxLine)
	}

	results.Aggregate()

	return results
}

func (me GoTestJSON) newSuite(pkg *goTestJSONPackage, results parser.TestResults) parser.Suite {
	suite := parser.NewSuite()
	suite.Name = pkg.Name
	if !pkg.Started.IsZero() {
		suite.Timestamp = pkg.Started.Format(time.RFC3339Nano)
	}
	suite.EnsureID(results)

	for _, record := range pkg.Records {
		for _, attempt := range record.Attempts {
			suite.Tests = append(suite.Tests, me.newTest(record.Name, pkg.Name, attempt, suite))
		}
	}

	if pkg.Failed && len(pkg.Records) == 0 {
		suite.Tests = append(suite.Tests, me.newPackageFailure(pkg, suite))
	}

	suite.Aggregate()

	return suite
}

// newTest turns one attempt into one test. The CLI model has no attempt
// dimension, so repeated runs of the same test become repeated tests that
// share an id.
func (me GoTestJSON) newTest(name string, pkg string, attempt *goTestJSONAttempt, suite parser.Suite) parser.Test {
	test := parser.NewTest()
	test.Name = name
	test.Classname = pkg
	test.State = goTestJSONState(attempt)
	test.Duration = time.Duration(attempt.Elapsed * float64(time.Second))

	switch test.State {
	case parser.StateError:
		failure := goTestJSONFailure(attempt)
		test.Error = &parser.Error{Type: failure.Type, Message: failure.Message, Body: failure.Body}
	case parser.StateFailed:
		test.Failure = goTestJSONFailure(attempt)
	}

	test.EnsureID(suite)

	return test
}

func (me GoTestJSON) newPackageFailure(pkg *goTestJSONPackage, suite parser.Suite) parser.Test {
	output := strings.Join(pkg.Output, "")

	test := parser.NewTest()
	test.Name = goTestJSONPackageFailureName
	test.Classname = pkg.Name
	test.State = parser.StateError
	test.Error = &parser.Error{
		Type:    goTestJSONPackageFailureType,
		Message: goTestJSONFirstLine(output),
		Body:    goTestJSONTrimStack(output),
	}
	test.EnsureID(suite)

	return test
}

func goTestJSONState(attempt *goTestJSONAttempt) parser.State {
	switch attempt.State {
	case "pass":
		return parser.StatePassed
	case "skip":
		return parser.StateSkipped
	case "fail":
		if goTestJSONPanicPattern.MatchString(strings.Join(attempt.Output, "")) {
			return parser.StateError
		}
		return parser.StateFailed
	default:
		return parser.StateError
	}
}

// goTestJSONFailure keeps the failing call site inside the body. Promoting it
// to Test.File would make a test's identity move with its assertions.
func goTestJSONFailure(attempt *goTestJSONAttempt) *parser.Failure {
	output := strings.Join(attempt.Output, "")

	failure := parser.NewFailure()
	failure.Message = goTestJSONFirstFailureLine(output)
	failure.Body = goTestJSONTrimStack(output)

	if goTestJSONPanicPattern.MatchString(output) {
		failure.Type = "panic"
	}

	if site := goTestJSONFailureSitePattern.FindStringSubmatch(output); site != nil {
		failure.Body = site[1] + ":" + site[2] + "\n" + failure.Body
	}

	return &failure
}

func goTestJSONFirstFailureLine(output string) string {
	if site := goTestJSONFailureSitePattern.FindStringSubmatch(output); site != nil && site[3] != "" {
		return strings.TrimSpace(site[3])
	}

	if match := goTestJSONPanicPattern.FindString(output); match != "" {
		for _, line := range strings.Split(output, "\n") {
			if strings.HasPrefix(line, match) {
				return strings.TrimSpace(line)
			}
		}
	}

	return goTestJSONFirstLine(output)
}

func goTestJSONFirstLine(output string) string {
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "=== ") && !strings.HasPrefix(trimmed, "--- ") {
			return trimmed
		}
	}

	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) != "" {
			return strings.TrimSpace(line)
		}
	}

	return ""
}

func goTestJSONTrimStack(output string) string {
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	if len(lines) > goTestJSONStackLines {
		lines = lines[len(lines)-goTestJSONStackLines:]
	}

	return strings.Join(lines, "\n")
}
