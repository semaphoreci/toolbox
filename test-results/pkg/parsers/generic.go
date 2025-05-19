package parsers

import (
	"fmt"
	"strings"

	"github.com/semaphoreci/toolbox/test-results/pkg/logger"
	"github.com/semaphoreci/toolbox/test-results/pkg/parser"
)

// Generic ...
type Generic struct {
}

// NewGeneric ...
func NewGeneric() Generic {
	return Generic{}
}

// IsApplicable ...
func (me Generic) IsApplicable(path string) bool {
	logger.Debug("Checking applicability of %s parser", me.GetName())
	return true
}

// GetName ...
func (me Generic) GetName() string {
	return "generic"
}

// Parse ...
func (me Generic) Parse(path string) parser.TestResults {
	results := parser.NewTestResults()

	xmlElement, err := LoadXML(path)

	if err != nil {
		logger.Error("Loading XML failed: %v", err)
		results.Status = parser.StatusError
		results.StatusMessage = err.Error()
		return results
	}

	switch xmlElement.Tag() {
	case "testsuites":
		logger.Debug("Root <testsuites> element found")
		results = me.newTestResults(*xmlElement)
	case "testsuite":
		logger.Debug("No root <testsuites> element found")
		results.Name = strings.Title(me.GetName() + " suite")
		results.EnsureID()
		results.Suites = append(results.Suites, me.newSuite(*xmlElement, results))
	default:
		tag := xmlElement.Tag()
		logger.Debug("Invalid root element found: <%s>", tag)
		results.Status = parser.StatusError
		results.StatusMessage = fmt.Sprintf("Invalid root element found: <%s>, must be one of <testsuites>, <testsuite>", tag)
		return results
	}

	results.Aggregate()
	results.Status = parser.StatusSuccess

	return results
}

func (me Generic) newTestResults(xml parser.XMLElement) parser.TestResults {
	testResults := parser.NewTestResults()
	logger.Trace("Parsing TestResults element with name: %s", xml.Attr("name"))

	for attr, value := range xml.Attributes {
		switch attr {
		case "name":
			testResults.Name = value
		case "time":
			testResults.Summary.Duration = parser.ParseTime(value)
		case "tests":
			testResults.Summary.Total = parser.ParseInt(value)
		case "failures":
			testResults.Summary.Failed = parser.ParseInt(value)
		case "errors":
			testResults.Summary.Error = parser.ParseInt(value)
		case "disabled":
			testResults.Summary.Disabled = parser.ParseInt(value)
		}
	}
	testResults.EnsureID()

	for _, node := range xml.Children {
		switch node.Tag() {
		case "testsuite":
			testResults.Suites = append(testResults.Suites, me.newSuite(node, testResults))
		}
	}
	testResults.Summary.Passed = testResults.Summary.Total - testResults.Summary.Error - testResults.Summary.Failed

	return testResults
}

func (me Generic) newSuite(xml parser.XMLElement, results parser.TestResults) parser.Suite {
	suite := parser.NewSuite()

	logger.Trace("Parsing Suite element with name: %s", xml.Attr("name"))

	for attr, value := range xml.Attributes {
		switch attr {
		case "name":
			suite.Name = value
		case "tests":
			suite.Summary.Total = parser.ParseInt(value)
		case "failures":
			suite.Summary.Failed = parser.ParseInt(value)
		case "errors":
			suite.Summary.Error = parser.ParseInt(value)
		case "time":
			suite.Summary.Duration = parser.ParseTime(value)
		case "disabled":
			suite.Summary.Disabled = parser.ParseInt(value)
		case "skipped":
			suite.Summary.Skipped = parser.ParseInt(value)
		case "timestamp":
			suite.Timestamp = value
		case "hostname":
			suite.Hostname = value
		case "id":
			suite.ID = value
		case "package":
			suite.Package = value
		}
	}

	suite.EnsureID(results)

	for _, node := range xml.Children {
		switch node.Tag() {
		case "properties":
			suite.Properties = parser.ParseProperties(node)
		case "system-out":
			suite.SystemOut = string(node.Contents)
		case "system-err":
			suite.SystemErr = string(node.Contents)
		case "testcase":
			suite.Tests = append(suite.Tests, me.newTest(node, suite))
		}
	}
	suite.Aggregate()

	return suite
}

func (me Generic) newTest(xml parser.XMLElement, suite parser.Suite) parser.Test {
	test := parser.NewTest()
	logger.Trace("Parsing Test element with name: %s", xml.Attr("name"))

	for attr, value := range xml.Attributes {
		switch attr {
		case "name":
			test.Name = value
		case "file":
			test.File = value
		case "time":
			test.Duration = parser.ParseTime(value)
		case "classname":
			test.Classname = value
		case "class":
			test.Classname = value
		}
	}

	for _, node := range xml.Children {
		switch node.Tag() {
		case "failure":
			test.State = parser.StateFailed
			test.Failure = parser.ParseFailure(node)
		case "error":
			test.State = parser.StateError
			test.Error = parser.ParseError(node)
		case "skipped":
			test.State = parser.StateSkipped
		case "system-out":
			test.SystemOut = string(node.Contents)
		case "system-err":
			test.SystemErr = string(node.Contents)
		}
	}

	test.EnsureID(suite)

	return test
}
