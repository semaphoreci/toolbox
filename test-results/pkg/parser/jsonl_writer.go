package parser

import (
	"bytes"
	"encoding/json"
	"os"
)

const (
	jsonlSchema  = "semaphore/test-report"
	jsonlVersion = "2.0-draft.2"
)

type jsonlHeader struct {
	Kind      string `json:"kind"`
	Schema    string `json:"schema"`
	Version   string `json:"version"`
	Reporter  string `json:"reporter"`
	Framework string `json:"framework"`
	Hostname  string `json:"hostname,omitempty"`
	StartedAt string `json:"started_at,omitempty"`
}

type jsonlIdentity struct {
	Name      string `json:"name"`
	Classname string `json:"classname,omitempty"`
	File      string `json:"file,omitempty"`
	Package   string `json:"package,omitempty"`
}

type jsonlFailure struct {
	Type    string `json:"type,omitempty"`
	Message string `json:"message"`
	Stack   string `json:"stack,omitempty"`
}

type jsonlAttempt struct {
	Index      int           `json:"index"`
	State      string        `json:"state"`
	Kind       string        `json:"kind,omitempty"`
	DurationMS float64       `json:"duration_ms,omitempty"`
	Failure    *jsonlFailure `json:"failure,omitempty"`
}

type jsonlTest struct {
	Kind       string            `json:"kind"`
	Identity   jsonlIdentity     `json:"identity"`
	State      string            `json:"state"`
	DurationMS float64           `json:"duration_ms"`
	Tags       map[string]string `json:"tags"`
	Attempts   []jsonlAttempt    `json:"attempts"`
}

type jsonlSummary struct {
	Kind       string  `json:"kind"`
	Tests      int     `json:"tests"`
	Complete   bool    `json:"complete"`
	DurationMS float64 `json:"duration_ms"`
}

// WriteJSONLResult serializes every parsed result set as one JSON Lines
// stream — one report context per TestResults, concatenated.
func WriteJSONLResult(result *Result, reporter string) ([]byte, error) {
	var buf bytes.Buffer
	for i := range result.TestResults {
		payload, err := WriteJSONL(&result.TestResults[i], reporter)
		if err != nil {
			return nil, err
		}
		buf.Write(payload)
	}
	return buf.Bytes(), nil
}

// WriteJSONL serializes parsed test results as a semaphore/test-report
// JSON Lines stream. Consecutive tests sharing an ID within a suite are
// folded into one test record with ordered attempts.
func WriteJSONL(results *TestResults, reporter string) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	hostname, _ := os.Hostname()

	framework := results.Framework
	if framework == "" {
		framework = "generic"
	}

	startedAt := ""
	if len(results.Suites) > 0 {
		startedAt = results.Suites[0].Timestamp
	}

	if err := encoder.Encode(jsonlHeader{
		Kind:      "report",
		Schema:    jsonlSchema,
		Version:   jsonlVersion,
		Reporter:  reporter,
		Framework: framework,
		Hostname:  hostname,
		StartedAt: startedAt,
	}); err != nil {
		return nil, err
	}

	testCount := 0
	var totalMS float64

	for _, suite := range results.Suites {
		for start := 0; start < len(suite.Tests); {
			end := start
			for end+1 < len(suite.Tests) && suite.Tests[end+1].ID == suite.Tests[start].ID {
				end++
			}

			group := suite.Tests[start : end+1]
			record := jsonlTest{
				Kind: "test",
				Identity: jsonlIdentity{
					Name:      group[0].Name,
					Classname: group[0].Classname,
					File:      group[0].File,
					Package:   group[0].Package,
				},
				Tags: map[string]string{},
			}

			for idx := range group {
				attempt := jsonlAttempt{
					Index:      idx,
					State:      jsonlState(group[idx].State),
					DurationMS: float64(group[idx].Duration.Microseconds()) / 1000,
				}
				if len(group) > 1 {
					attempt.Kind = "report_repeat"
				}
				if group[idx].Failure != nil {
					attempt.Failure = &jsonlFailure{
						Type:    group[idx].Failure.Type,
						Message: group[idx].Failure.Message,
						Stack:   group[idx].Failure.Body,
					}
				} else if group[idx].Error != nil {
					attempt.Failure = &jsonlFailure{
						Type:    group[idx].Error.Type,
						Message: group[idx].Error.Message,
						Stack:   group[idx].Error.Body,
					}
				}
				record.Attempts = append(record.Attempts, attempt)
				totalMS += attempt.DurationMS
			}

			final := record.Attempts[len(record.Attempts)-1]
			record.State = final.State
			record.DurationMS = final.DurationMS

			if err := encoder.Encode(record); err != nil {
				return nil, err
			}
			testCount++
			start = end + 1
		}
	}

	if err := encoder.Encode(jsonlSummary{
		Kind:       "summary",
		Tests:      testCount,
		Complete:   true,
		DurationMS: totalMS,
	}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func jsonlState(state State) string {
	switch state {
	case StatePassed:
		return "passed"
	case StateFailed:
		return "failed"
	case StateError:
		return "error"
	case StateSkipped, StateDisabled:
		return "skipped"
	default:
		return "error"
	}
}
