package parser

// Parser ...
type Parser interface {
	Parse(string) TestResults
	IsApplicable(string) bool
	GetName() string
}
