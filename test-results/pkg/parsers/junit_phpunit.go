package parsers

import (
	"fmt"
	"strings"

	"github.com/semaphoreci/toolbox/test-results/pkg/logger"
	"github.com/semaphoreci/toolbox/test-results/pkg/parser"
)

// JUnitPHPUnit ...
type JUnitPHPUnit struct {
}

// NewJUnitPHPUnit ...
func NewJUnitPHPUnit() JUnitPHPUnit {
	return JUnitPHPUnit{}
}

// GetName ...
func (me JUnitPHPUnit) GetName() string {
	return "phpunit"
}

// GetDescription ...
func (me JUnitPHPUnit) GetDescription() string {
	return "PHP PHPUnit test output (JUnit format)"
}

// GetSupportedExtensions ...
func (me JUnitPHPUnit) GetSupportedExtensions() []string {
	return []string{".xml"}
}

// IsApplicable ...
func (me JUnitPHPUnit) IsApplicable(path string) bool {
	return false
}

// Parse ...
func (me JUnitPHPUnit) Parse(path string) parser.TestResults {
	results := parser.NewTestResults()
	results.Name = "PHPUnit Suite"
	results.Framework = me.GetName()
	results.EnsureID()

	xmlElement, err := LoadXML(path)

	if err != nil {
		logger.Error("Loading XML failed: %v", err)
		results.Status = parser.StatusError
		results.StatusMessage = err.Error()
		return results
	}

	flattenTestSuites(xmlElement)

	switch xmlElement.Tag() {
	case "testsuites":
		logger.Debug("Root <testsuites> element found")
		results = me.newTestResults(*xmlElement)
	case "testsuite":
		logger.Debug("No root <testsuites> element found")
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

func flattenTestSuites(xmlElement *parser.XMLElement) {
	newSuites := []parser.XMLElement{}

	for _, node := range xmlElement.Children {
		switch node.Tag() {
		case "testsuite":
			flattenTestSuite(node, node.Attr("name"), &newSuites)
		}
	}

	xmlElement.Children = newSuites
}

func flattenTestSuite(xmlNode parser.XMLElement, suiteNamePrefix string, newSuites *[]parser.XMLElement) {
loop:
	for _, childNode := range xmlNode.Children {
		switch childNode.Tag() {
		case "testsuite":
			newSuitePrefix := prefixSuiteName(childNode.Attr("name"), suiteNamePrefix)
			flattenTestSuite(childNode, newSuitePrefix, newSuites)

		case "testcase":
			xmlNode.Attributes["name"] = suiteNamePrefix
			*newSuites = append(*newSuites, xmlNode)
			break loop
		}
	}
}

func prefixSuiteName(suiteName string, prefix string) string {
	if prefix == "" {
		return suiteName
	}

	if suiteName == "" {
		return prefix
	}

	return prefix + "\\" + suiteName
}

func (me JUnitPHPUnit) newTestResults(xml parser.XMLElement) parser.TestResults {
	testResults := parser.NewTestResults()
	testResults.Name = "PHPUnit Suite"

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

	testResults.Framework = me.GetName()
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

func (me JUnitPHPUnit) newSuite(xml parser.XMLElement, testResults parser.TestResults) parser.Suite {
	suite := parser.NewSuite()

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

func (me JUnitPHPUnit) newTest(xml parser.XMLElement, suite parser.Suite) parser.Test {
	test := parser.NewTest()
	for attr, value := range xml.Attributes {
		switch attr {
		case "name":
			test.Name = value
		case "time":
			test.Duration = parser.ParseTime(value)
		case "classname":
			test.Classname = value
		case "id":
			test.ID = value
		case "file":
			test.File = strings.TrimPrefix(value, "./")
		case "package":
			test.Package = value
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
