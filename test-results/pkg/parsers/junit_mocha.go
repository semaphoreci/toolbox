package parsers

import (
	"fmt"
	"strings"

	"github.com/semaphoreci/toolbox/test-results/pkg/logger"
	"github.com/semaphoreci/toolbox/test-results/pkg/parser"
)

// JUnitMocha ...
type JUnitMocha struct {
}

// NewJUnitMocha ...
func NewJUnitMocha() JUnitMocha {
	return JUnitMocha{}
}

// GetName ...
func (me JUnitMocha) GetName() string {
	return "mocha"
}

// GetDescription ...
func (me JUnitMocha) GetDescription() string {
	return "Mocha test output in JUnit format"
}

// GetSupportedExtensions ...
func (me JUnitMocha) GetSupportedExtensions() []string {
	return []string{".xml"}
}

// IsApplicable ...
func (me JUnitMocha) IsApplicable(path string) bool {
	xmlElement, err := LoadXML(path)
	logger.Debug("Checking applicability of %s parser", me.GetName())

	if err != nil {
		logger.Error("Loading XML failed: %v", err)
		return false
	}

	switch xmlElement.Tag() {
	case "testsuites":
		for attr, value := range xmlElement.Attributes {
			logger.Trace("%s %s", attr, value)
			switch attr {
			case "name":
				if strings.Contains(strings.ToLower(value), "mocha") {
					return true
				}
			}
		}
	}
	return false
}

// Parse ...
func (me JUnitMocha) Parse(path string) parser.TestResults {
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
		results.Framework = me.GetName()
		results.Suites = append(results.Suites, me.newSuite(*xmlElement, results))
	default:
		tag := xmlElement.Tag()
		logger.Debug("Invalid root element found: <%s>", tag)
		results.Status = parser.StatusError
		results.StatusMessage = fmt.Sprintf("Invalid root element found: <%s>, must be one of <testsuites>, <testsuite>", tag)
	}

	results.Aggregate()

	return results
}

func (me JUnitMocha) newTestResults(xml parser.XMLElement) parser.TestResults {
	testResults := parser.NewTestResults()

	testResults.Framework = me.GetName()

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
			testResults.IsDisabled = parser.ParseBool(value)
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

func (me JUnitMocha) newSuite(xml parser.XMLElement, testResults parser.TestResults) parser.Suite {
	suite := parser.NewSuite()

	for attr, value := range xml.Attributes {
		switch attr {
		case "name":
			suite.Name = strings.TrimPrefix(value, "Mocha")
		case "tests":
			suite.Summary.Total = parser.ParseInt(value)
		case "failures":
			suite.Summary.Failed = parser.ParseInt(value)
		case "errors":
			suite.Summary.Error = parser.ParseInt(value)
		case "time":
			suite.Summary.Duration = parser.ParseTime(value)
		case "disabled":
			suite.IsDisabled = parser.ParseBool(value)
		case "skipped":
			suite.IsSkipped = parser.ParseBool(value)
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

	suite.EnsureID(testResults)
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

func (me JUnitMocha) newTest(xml parser.XMLElement, suite parser.Suite) parser.Test {
	test := parser.NewTest()

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
