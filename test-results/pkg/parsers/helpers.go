package parsers

import (
	"bytes"
	"os"

	"github.com/semaphoreci/toolbox/test-results/pkg/fileloader"
	"github.com/semaphoreci/toolbox/test-results/pkg/parser"
)

// LoadPath ...
func LoadPath(path string) (*bytes.Reader, error) {
	var reader *bytes.Reader
	// Preload path with loader. If nothing is found in file cache - load it up from path.
	reader, found := fileloader.Load(path, &bytes.Reader{})

	if !found {
		file, err := os.ReadFile(path)

		if err != nil {
			return nil, err
		}

		b := bytes.NewReader(file)
		reader, _ = fileloader.Load(path, b)
	}
	return reader, nil
}

// LoadXML ...
func LoadXML(path string) (*parser.XMLElement, error) {
	reader, err := LoadPath(path)
	if err != nil {
		return nil, err
	}

	xmlElement := parser.NewXMLElement()

	err = xmlElement.Parse(reader)
	if err != nil {
		return nil, err
	}

	return &xmlElement, nil
}

// LoadFile loads a file from the given path
func LoadFile(path string) ([]byte, error) {
	// Check file cache first
	reader, found := fileloader.Load(path, &bytes.Reader{})

	if found {
		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(reader)
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Cache for future use
	fileloader.Load(path, bytes.NewReader(data))

	return data, nil
}
