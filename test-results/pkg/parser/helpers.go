package parser

import (
	"fmt"
	"strconv"
	"time"

	"github.com/semaphoreci/toolbox/test-results/pkg/logger"
)

// ParseProperties maps <properties>
func ParseProperties(xml XMLElement) Properties {
	properties := make(map[string]string)
	for _, node := range xml.Children {
		properties[node.Attr("name")] = node.Attr("value")
	}

	return properties
}

// PropertyExists checks if `propertyName` is in Properties
func PropertyExists(properties Properties, propertyName string) bool {
	if _, ok := properties[propertyName]; ok {
		return true
	}
	return false
}

// ParseFailure parses <failure> element from junit schema
func ParseFailure(xml XMLElement) *Failure {
	failure := NewFailure()

	failure.Body = string(xml.Contents)
	failure.Message = xml.Attr("message")
	failure.Type = xml.Attr("type")

	return &failure
}

// ParseError parses <error> element from junit schema
func ParseError(xml XMLElement) *Error {
	err := NewError()

	err.Body = string(xml.Contents)
	err.Message = xml.Attr("message")
	err.Type = xml.Attr("type")

	return &err
}

// ParseTime parsers time from junit.xml schemas
func ParseTime(s string) time.Duration {

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Warn("Duration parsing failed: %v", err)
		return 0
	}
	// append 's' to end of input to use `time` built in duration parser
	d, err := time.ParseDuration(fmt.Sprintf("%fs", f))
	if err != nil {
		logger.Warn("Duration parsing failed: %v", err)
		return 0
	}

	return d
}

// ParseInt parsers string respresentation of integer to integer value
func ParseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		logger.Warn("Integer parsing failed: %v", err)
		return 0
	}
	return i
}

// ParseBool parses stirng representation of boolean to boolean value
func ParseBool(s string) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		logger.Warn("Boolean parsing failed: %v", err)
		return false
	}
	return b
}
