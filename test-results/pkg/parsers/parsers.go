package parsers

import (
	"fmt"
	"path/filepath"

	"github.com/semaphoreci/toolbox/test-results/pkg/logger"
	"github.com/semaphoreci/toolbox/test-results/pkg/parser"
)

var availableParsers = []parser.Parser{
	// JUnit parsers
	NewJUnitRSpec(),
	NewJUnitExUnit(),
	NewJUnitMocha(),
	NewJUnitGoLang(),
	NewJUnitPHPUnit(),
	NewJUnitEmbedded(),
	NewJUnitGeneric(),
	// Custom parsers
	NewGoStaticcheck(),
	NewGoRevive(),
}

// FindParser ...
func FindParser(name string, path string) (parser.Parser, error) {
	if name != "auto" {
		for _, p := range availableParsers {
			if p.GetName() == name {
				logger.Debug("Found parser: %s", p.GetName())
				return p, nil
			}
		}
		logger.Debug("Parser not found: %s", name)
		return nil, fmt.Errorf("parser not found: %s", name)
	}

	// First filter by file extension
	fileExt := filepath.Ext(path)
	var compatibleParsers []parser.Parser

	for _, p := range availableParsers {
		supportedExts := p.GetSupportedExtensions()
		for _, ext := range supportedExts {
			if ext == fileExt {
				compatibleParsers = append(compatibleParsers, p)
				logger.Debug("Parser %s supports extension %s", p.GetName(), fileExt)
				break
			}
		}
	}

	// Then check IsApplicable only for compatible parsers
	for _, p := range compatibleParsers {
		isApplicable := p.IsApplicable(path)
		logger.Debug("Looking for applicable parser, checking %s -> %t", p.GetName(), isApplicable)
		if isApplicable {
			logger.Trace("Found applicable parser: %s", p.GetName())
			return p, nil
		}
	}

	return nil, fmt.Errorf("no applicable parsers found for %s", path)
}

// GetAvailableParserNames returns a list of all available parser names
func GetAvailableParserNames() []string {
	names := make([]string, len(availableParsers))
	for i, p := range availableParsers {
		names[i] = p.GetName()
	}
	return names
}

// ParserInfo contains parser name and description
type ParserInfo struct {
	Name        string
	Description string
}

// GetAvailableParsers returns a list of all available parsers with descriptions
func GetAvailableParsers() []ParserInfo {
	parsers := make([]ParserInfo, len(availableParsers))
	for i, p := range availableParsers {
		parsers[i] = ParserInfo{
			Name:        p.GetName(),
			Description: p.GetDescription(),
		}
	}
	return parsers
}

// GetSupportedExtensions returns all unique extensions supported by all parsers
func GetSupportedExtensions() []string {
	extMap := make(map[string]bool)
	for _, p := range availableParsers {
		for _, ext := range p.GetSupportedExtensions() {
			extMap[ext] = true
		}
	}

	var extensions []string
	for ext := range extMap {
		extensions = append(extensions, ext)
	}
	return extensions
}
