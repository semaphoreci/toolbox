package parsers

import (
	"bufio"
	"bytes"
	"os"

	"github.com/semaphoreci/toolbox/test-results/pkg/fileloader"
	"github.com/semaphoreci/toolbox/test-results/pkg/parser"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// LoadPath ...
func LoadPath(path string) (*bytes.Reader, error) {
	var reader *bytes.Reader
	// Preload path with loader. If nothing is found in file cache - load it up from path.
	reader, found := fileloader.Load(path, &bytes.Reader{})

	if !found {
		file, err := os.ReadFile(path) // #nosec

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

	data, err := os.ReadFile(path) // #nosec
	if err != nil {
		return nil, err
	}

	// Cache for future use
	fileloader.Load(path, bytes.NewReader(data))

	return data, nil
}

func Title(s string) string {
	caser := cases.Title(language.English)
	return caser.String(s)
}

// jsonlSplit scans lines like bufio.ScanLines, but drops lines longer than
// maxLine instead of failing the whole scan.
func jsonlSplit(maxLine int, oversized *int) bufio.SplitFunc {
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
			if len(data) >= maxLine {
				*oversized++
				return len(data), nil, nil
			}
			return len(data), dropCR(data), nil
		}

		if len(data) >= maxLine {
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
