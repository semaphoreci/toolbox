package parsers

import (
	"fmt"

	"github.com/semaphoreci/toolbox/test-results/pkg/logger"
	"github.com/semaphoreci/toolbox/test-results/pkg/parser"
)

var availableParsers = []parser.Parser{
	NewRSpec(),
	NewExUnit(),
	NewMocha(),
	NewGoLang(),
	NewPHPUnit(),
	NewGeneric(),
	NewEmbedded(),
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
		logger.Debug("Parser not found")
	}

	for _, p := range availableParsers {
		isApplicable := p.IsApplicable(path)
		logger.Debug("Looking for applicable parser, checking %s -> %t", p.GetName(), isApplicable)
		if isApplicable {
			logger.Trace("Found applicable parser: %s", p.GetName())
			return p, nil
		}
	}

	return nil, fmt.Errorf("no applicable parsers found")
}
