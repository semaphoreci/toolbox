package parsers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/semaphoreci/toolbox/test-results/pkg/logger"
	"github.com/semaphoreci/toolbox/test-results/pkg/parser"
)

const (
	semaphoreJSONLSchema    = "semaphore/test-report"
	semaphoreJSONLMaxLine   = 1024 * 1024
	semaphoreJSONLProbeSize = 64 * 1024
)

type semaphoreJSONLIdentity struct {
	Name      string   `json:"name"`
	Classname string   `json:"classname"`
	File      string   `json:"file"`
	Line      int      `json:"line"`
	Package   string   `json:"package"`
	Hierarchy []string `json:"hierarchy"`
}

type semaphoreJSONLFailure struct {
	Type      string `json:"type"`
	Message   string `json:"message"`
	Stack     string `json:"stack"`
	Truncated bool   `json:"truncated"`
}

type semaphoreJSONLAttempt struct {
	Index       int                    `json:"index"`
	State       string                 `json:"state"`
	Kind        string                 `json:"kind"`
	StartedAt   string                 `json:"started_at"`
	DurationMs  *float64               `json:"duration_ms"`
	Failure     *semaphoreJSONLFailure `json:"failure"`
	SkipMessage string                 `json:"skip_message"`
}

type semaphoreJSONLRecord struct {
	Kind       string                  `json:"kind"`
	Schema     string                  `json:"schema"`
	Version    string                  `json:"version"`
	Reporter   string                  `json:"reporter"`
	Framework  string                  `json:"framework"`
	Hostname   string                  `json:"hostname"`
	Seed       *int64                  `json:"seed"`
	StartedAt  string                  `json:"started_at"`
	Identity   semaphoreJSONLIdentity  `json:"identity"`
	State      string                  `json:"state"`
	DurationMs *float64                `json:"duration_ms"`
	Attempts   []semaphoreJSONLAttempt `json:"attempts"`
	Tests      *int                    `json:"tests"`
}

type semaphoreJSONLEnvelope struct {
	Framework string
	Hostname  string
	StartedAt string
	Seed      *int64
}

func (me semaphoreJSONLEnvelope) Properties() parser.Properties {
	if me.Seed == nil {
		return nil
	}

	return parser.Properties{"seed": strconv.FormatInt(*me.Seed, 10)}
}

// SemaphoreJSONL ...
type SemaphoreJSONL struct {
}

// NewSemaphoreJSONL ...
func NewSemaphoreJSONL() SemaphoreJSONL {
	return SemaphoreJSONL{}
}

// GetName ...
func (me SemaphoreJSONL) GetName() string {
	return "semaphore-jsonl"
}

// GetDescription ...
func (me SemaphoreJSONL) GetDescription() string {
	return "Semaphore test report stream (JSON Lines)"
}

// GetSupportedExtensions ...
func (me SemaphoreJSONL) GetSupportedExtensions() []string {
	return []string{".jsonl", ".json"}
}

// IsApplicable ...
func (me SemaphoreJSONL) IsApplicable(path string) bool {
	logger.Debug("Checking applicability of %s parser", me.GetName())

	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		logger.Debug("Opening file failed: %v", err)
		return false
	}
	defer file.Close()

	line, err := bufio.NewReaderSize(file, semaphoreJSONLProbeSize).ReadSlice('\n')
	if err != nil && err != io.EOF {
		logger.Debug("Reading header line failed: %v", err)
		return false
	}

	header := semaphoreJSONLRecord{}
	if err := json.Unmarshal(bytes.TrimSpace(line), &header); err != nil {
		logger.Debug("Header line is not a JSON object: %v", err)
		return false
	}

	return header.Kind == "report" && header.Schema == semaphoreJSONLSchema
}

// Parse ...
func (me SemaphoreJSONL) Parse(path string) parser.TestResults {
	results := parser.NewTestResults()
	results.Name = Title(me.GetName() + " suite")
	results.Framework = me.GetName()
	results.EnsureID()

	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		logger.Error("Opening file failed: %v", err)
		results.Status = parser.StatusError
		results.StatusMessage = err.Error()
		return results
	}
	defer file.Close()

	suite := parser.NewSuite()
	suite.Name = me.GetName()
	suite.EnsureID(results)

	oversized := 0
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, semaphoreJSONLMaxLine), semaphoreJSONLMaxLine)
	scanner.Split(semaphoreJSONLSplit(&oversized))

	envelope := semaphoreJSONLEnvelope{}
	messages := []string{}
	seenHeader := false
	seenSummary := false
	contextTests := 0

	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		record := semaphoreJSONLRecord{}
		if err := json.Unmarshal(line, &record); err != nil {
			logger.Debug("Skipping unparsable line: %v", err)
			continue
		}

		switch record.Kind {
		case "report":
			envelope = semaphoreJSONLEnvelope{
				Framework: record.Framework,
				Hostname:  record.Hostname,
				StartedAt: record.StartedAt,
				Seed:      record.Seed,
			}
			contextTests = 0
			seenSummary = false

			if !seenHeader {
				seenHeader = true
				if record.Framework != "" {
					results.Name = Title(record.Framework + " suite")
					results.Framework = record.Framework
					results.ID = ""
					results.EnsureID()

					suite = parser.NewSuite()
					suite.Name = record.Framework
					suite.EnsureID(results)
				}
			}
		case "test":
			contextTests++
			suite.Tests = append(suite.Tests, me.newTests(record, suite)...)
		case "summary":
			seenSummary = true
			if record.Tests != nil && *record.Tests != contextTests {
				messages = append(messages, fmt.Sprintf("truncated: summary reported %d tests, parsed %d", *record.Tests, contextTests))
			}
		case "output":
			logger.Debug("Skipping output record")
		default:
			logger.Debug("Skipping unknown record kind: %s", record.Kind)
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Error("Reading %s failed: %v", path, err)
		results.Status = parser.StatusError
		results.StatusMessage = err.Error()
		return results
	}

	if !seenSummary {
		messages = append(messages, "truncated: missing summary trailer")
	}

	if oversized > 0 {
		messages = append(messages, fmt.Sprintf("skipped %d lines over %d bytes", oversized, semaphoreJSONLMaxLine))
	}

	results.Suites = append(results.Suites, suite)
	me.arrangeSuites(&results, envelope)
	results.StatusMessage = strings.Join(messages, "; ")

	return results
}

func (me SemaphoreJSONL) newTests(record semaphoreJSONLRecord, suite parser.Suite) []parser.Test {
	tests := []parser.Test{}

	attempts := record.Attempts
	if len(attempts) == 0 {
		attempts = []semaphoreJSONLAttempt{{State: record.State, DurationMs: record.DurationMs}}
	}

	for idx, attempt := range attempts {
		test := parser.NewTest()
		test.Name = record.Identity.Name
		test.Classname = record.Identity.Classname
		test.Package = record.Identity.Package
		test.File = strings.TrimPrefix(record.Identity.File, "./")
		test.State = semaphoreJSONLState(attempt.State)

		if attempt.DurationMs != nil {
			test.Duration = semaphoreJSONLDuration(attempt.DurationMs)
		} else if idx == len(attempts)-1 {
			test.Duration = semaphoreJSONLDuration(record.DurationMs)
		}

		if attempt.Failure != nil {
			switch test.State {
			case parser.StateError:
				test.Error = &parser.Error{
					Type:    attempt.Failure.Type,
					Message: attempt.Failure.Message,
					Body:    attempt.Failure.Stack,
				}
			case parser.StateFailed:
				test.Failure = &parser.Failure{
					Type:    attempt.Failure.Type,
					Message: attempt.Failure.Message,
					Body:    attempt.Failure.Stack,
				}
			}
		}

		test.EnsureID(suite)
		tests = append(tests, test)
	}

	return tests
}

func (me SemaphoreJSONL) arrangeSuites(results *parser.TestResults, envelope semaphoreJSONLEnvelope) {
	newSuites := []parser.Suite{}

	for _, suite := range results.Suites {
		for _, test := range suite.Tests {
			name := test.File
			if name == "" {
				name = test.Classname
			}
			if name == "" {
				name = suite.Name
			}

			idx, foundSuite := parser.EnsureSuiteByName(newSuites, name)

			foundSuite.Tests = append(foundSuite.Tests, test)
			foundSuite.Aggregate()

			if idx == -1 {
				foundSuite.Package = test.Package
				foundSuite.Hostname = envelope.Hostname
				foundSuite.Timestamp = envelope.StartedAt
				foundSuite.Properties = envelope.Properties()
				foundSuite.EnsureID(*results)
				newSuites = append(newSuites, *foundSuite)
			}
		}
	}

	results.Suites = newSuites
	results.Aggregate()
}

func semaphoreJSONLState(state string) parser.State {
	switch state {
	case "passed":
		return parser.StatePassed
	case "failed":
		return parser.StateFailed
	case "error":
		return parser.StateError
	case "skipped":
		return parser.StateSkipped
	}

	logger.Debug("Unknown test state: %s", state)
	return parser.StatePassed
}

func semaphoreJSONLDuration(ms *float64) time.Duration {
	if ms == nil {
		return 0
	}

	return time.Duration(*ms * float64(time.Millisecond))
}

// semaphoreJSONLSplit scans lines like bufio.ScanLines, but drops lines longer
// than the format's line cap instead of failing the whole scan.
func semaphoreJSONLSplit(oversized *int) bufio.SplitFunc {
	dropping := false

	return func(data []byte, atEOF bool) (int, []byte, error) {
		if dropping {
			if i := bytes.IndexByte(data, '\n'); i >= 0 {
				dropping = false
				return i + 1, nil, nil
			}
			return len(data), nil, nil
		}

		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			return i + 1, dropCR(data[:i]), nil
		}

		if atEOF {
			if len(data) == 0 {
				return 0, nil, nil
			}
			if len(data) >= semaphoreJSONLMaxLine {
				*oversized++
				return len(data), nil, nil
			}
			return len(data), dropCR(data), nil
		}

		if len(data) >= semaphoreJSONLMaxLine {
			*oversized++
			dropping = true
			return len(data), nil, nil
		}

		return 0, nil, nil
	}
}

func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[:len(data)-1]
	}

	return data
}
