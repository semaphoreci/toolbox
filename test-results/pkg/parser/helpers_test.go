package parser_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/semaphoreci/toolbox/test-results/pkg/parser"
	"github.com/stretchr/testify/assert"
)

func TestParseProperties(t *testing.T) {
	data := `
		<properties>
			<property name="foo" value="bar"></property>
			<property name="bar" value="123"></property>
		</properties>
	`
	reader := bytes.NewReader([]byte(data))
	xmlelement := parser.NewXMLElement()
	err := xmlelement.Parse(reader)
	assert.Nil(t, err, "should not error")

	properties := parser.ParseProperties(xmlelement)
	assert.Equal(t, parser.Properties{"foo": "bar", "bar": "123"}, properties)
}

func TestPropertyExists(t *testing.T) {
	properties := parser.Properties{"foo": "bar", "bar": "123"}

	assert.Equal(t, true, parser.PropertyExists(properties, "foo"))
	assert.Equal(t, true, parser.PropertyExists(properties, "bar"))
	assert.Equal(t, false, parser.PropertyExists(properties, "baz"))
}

func TestParseFailure(t *testing.T) {
	data := `<failure message="failure message" type="failure type">Some failure</failure>`
	reader := bytes.NewReader([]byte(data))
	xmlelement := parser.NewXMLElement()
	err := xmlelement.Parse(reader)
	assert.Nil(t, err, "should not error")

	assert.Equal(t, &parser.Failure{Message: "failure message", Type: "failure type", Body: "Some failure"}, parser.ParseFailure(xmlelement))
}

func TestParseError(t *testing.T) {
	data := `<error message="error message" type="error type">Some error</error>`
	reader := bytes.NewReader([]byte(data))
	xmlelement := parser.NewXMLElement()
	err := xmlelement.Parse(reader)
	assert.Nil(t, err, "should not error")

	assert.Equal(t, &parser.Error{Message: "error message", Type: "error type", Body: "Some error"}, parser.ParseError(xmlelement))
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		desc  string
		input string
		want  time.Duration
	}{
		{"parses 0.0013 correctly", "0.0013", 1*time.Millisecond + 300*time.Microsecond},
		{"parses 0.01 correctly", "0.01", 10 * time.Millisecond},
		{"parses 0.1 correctly", "0.1", 100 * time.Millisecond},
		{"parses 0 correctly", "0", 0 * time.Second},
		{"parses 1 correctly", "1", 1 * time.Second},
		{"parses 60 correctly", "60", 1 * time.Minute},
		{"parses 61.123 correctly", "61.123", 1*time.Minute + 1*time.Second + 123*time.Millisecond},
		{"parses 8.10623168945312e-05s correctly", "8.10623168945312e-05", 81 * time.Microsecond},
		{"parses invalid number correctly #1", "a60", 0},
		{"parses invalid number correctly #2", "6a0", 0},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {

			got := parser.ParseTime(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		desc  string
		input string
		want  int
	}{
		{"parses 0 correctly", "0", 0},
		{"parses 1 correctly", "1", 1},
		{"parses 60 correctly", "60", 60},
		{"parses 1000000 correctly", "1000000", 1_000_000},
		{"parses invalid number correctly #1", "a60", 0},
		{"parses invalid number correctly #2", "6a0", 0},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {

			got := parser.ParseInt(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		desc  string
		input string
		want  bool
	}{
		{"parses true correctly", "true", true},
		{"parses false correctly", "false", false},
		{"parses 0 correctly", "0", false},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {

			got := parser.ParseBool(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
