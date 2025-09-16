package parsers

import (
	"fmt"

	"github.com/semaphoreci/toolbox/test-results/pkg/logger"
	"github.com/semaphoreci/toolbox/test-results/pkg/parser"
)

// JUnitEmbedded ...
type JUnitEmbedded struct {
}

// NewJUnitEmbedded initialize an empty embedded parser
func NewJUnitEmbedded() JUnitEmbedded {
	return JUnitEmbedded{}
}

// GetName returns the parser name
func (e JUnitEmbedded) GetName() string {
	return "embedded"
}

// GetDescription returns the parser description
func (e JUnitEmbedded) GetDescription() string {
	return "JUnit format with nested testsuites"
}

// GetSupportedExtensions ...
func (e JUnitEmbedded) GetSupportedExtensions() []string {
	return []string{".xml"}
}

// IsApplicable checks if the current xml is compatible with embedded parser.
func (e JUnitEmbedded) IsApplicable(string) bool {
	return false
}

// Parse parses the string given using the embedded format
func (e JUnitEmbedded) Parse(path string) parser.TestResults {
	results := parser.NewTestResults()
	results.Name = "Suite"
	results.Framework = e.GetName()
	results.EnsureID()

	xmlElement, err := LoadXML(path)

	if err != nil {
		logger.Error("Loading XML failed: %v", err)
		results.Status = parser.StatusError
		results.StatusMessage = err.Error()
		return results
	}

	flatten(xmlElement)

	switch xmlElement.Tag() {
	case "testsuites":
		logger.Debug("Root <testsuites> element found")
		results = e.newTestResults(*xmlElement)
	case "testsuite":
		tag := xmlElement.Tag()
		logger.Debug("<testsuite> as root element not supported")
		results.Status = parser.StatusError
		results.StatusMessage = fmt.Sprintf("Invalid root element found: <%s>, must be <testsuites>", tag)
	default:
		tag := xmlElement.Tag()
		logger.Debug("Invalid root element found: <%s>", tag)
		results.Status = parser.StatusError
		results.StatusMessage = fmt.Sprintf("Invalid root element found: <%s>,  must be <testsuites>", tag)
	}

	return results
}

func flatten(root *parser.XMLElement) {
	var newSuites []parser.XMLElement

	for _, node := range root.Children {
		switch node.Tag() {
		case "testsuite":
			flattenSingleSuite(node, node.Attr("name"), &newSuites)
		}
	}

	root.Children = newSuites
}

func flattenSingleSuite(xmlNode parser.XMLElement, suiteNamePrefix string, newSuites *[]parser.XMLElement) {
	childNodeAdded := false
	for _, childNode := range xmlNode.Children {
		switch childNode.Tag() {
		case "testsuite":
			newSuitePrefix := prefixSuiteName(childNode.Attr("name"), suiteNamePrefix)
			flattenSingleSuite(childNode, newSuitePrefix, newSuites)
		case "testcase":
			if !childNodeAdded {
				xmlNode.Attributes["name"] = suiteNamePrefix
				*newSuites = append(*newSuites, xmlNode)
				childNodeAdded = true
			}
		}
	}
}

func (e JUnitEmbedded) newTestResults(xmlElement parser.XMLElement) parser.TestResults {
	results := parser.NewTestResults()
	results.Name = "Suite"

	for attr, value := range xmlElement.Attributes {
		switch attr {
		case "name":
			results.Name = value
		case "time":
			results.Summary.Duration = parser.ParseTime(value)
		case "tests":
			results.Summary.Total = parser.ParseInt(value)
		case "failures":
			results.Summary.Failed = parser.ParseInt(value)
		case "errors":
			results.Summary.Error = parser.ParseInt(value)
		case "disabled":
			results.IsDisabled = parser.ParseBool(value)
		}
	}

	results.Framework = e.GetName()
	results.EnsureID()

	for _, node := range xmlElement.Children {
		switch node.Tag() {
		case "testsuite":
			results.Suites = append(results.Suites, e.newSuite(node, results))
		}
	}
	results.Summary.Passed = results.Summary.Total - results.Summary.Error - results.Summary.Failed

	return results
}

func (e JUnitEmbedded) newSuite(xml parser.XMLElement, testResults parser.TestResults) parser.Suite {
	suite := parser.NewSuite()

	logger.Trace("Parsing Suite element with name: %s", xml.Attr("name"))

	for attr, value := range xml.Attributes {
		switch attr {
		case "name":
			suite.Name = value
		case "time":
			suite.Summary.Duration = parser.ParseTime(value)
		case "tests":
			suite.Summary.Total = parser.ParseInt(value)
		case "failures":
			suite.Summary.Failed = parser.ParseInt(value)
		case "errors":
			suite.Summary.Error = parser.ParseInt(value)
		case "disabled":
			suite.Summary.Disabled = parser.ParseInt(value)
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
			suite.Tests = append(suite.Tests, e.newTest(node, suite))
		}
	}

	suite.Aggregate()

	return suite
}

func (e JUnitEmbedded) newTest(xml parser.XMLElement, suite parser.Suite) parser.Test {
	test := parser.NewTest()
	logger.Trace("Parsing Test element with name: %s", xml.Attr("name"))

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
