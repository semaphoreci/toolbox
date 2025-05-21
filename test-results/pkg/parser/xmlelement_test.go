package parser

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewXMLElement(t *testing.T) {
	el := NewXMLElement()

	assert.Equal(t, el, XMLElement{})
}

func TestXMLElement_Attr(t *testing.T) {

	reader := bytes.NewReader([]byte(`<test foo="1" bar="2"></test>`))

	xmlelement := NewXMLElement()
	xmlelement.Parse(reader)

	assert.Equal(t, "1", xmlelement.Attr("foo"))
	assert.Equal(t, "2", xmlelement.Attr("bar"))
	assert.Equal(t, "", xmlelement.Attr("baz"))
}

func TestXMLElement_Tag(t *testing.T) {

	reader := bytes.NewReader([]byte(`<test foo="1" bar="2"></test>`))

	xmlelement := NewXMLElement()
	xmlelement.Parse(reader)

	assert.Equal(t, "test", xmlelement.Tag())
}

func TestXMLElement_Parse(t *testing.T) {
	malformedData := [...]string{
		"test",
		"<test",
		"<test>",
		"<test><te</test>",
		"<test name=\"1 < 2\"></test>",
		"<test name=\"1 & 2\"></test>",
	}

	for _, data := range malformedData {
		reader := bytes.NewReader([]byte(data))

		xmlelement := NewXMLElement()
		err := xmlelement.Parse(reader)
		assert.Error(t, err, "should error on malformed xml")
	}
}
