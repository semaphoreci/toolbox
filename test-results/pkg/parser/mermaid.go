package parser

import (
	"fmt"
	"strings"
	"unicode"
)

// EscapeGanttLabel sanitizes text for use in Mermaid gantt chart labels.
// It escapes special characters that have meaning in Mermaid syntax using
// entity codes as documented at:
// https://mermaid.js.org/syntax/sequenceDiagram.html#entity-codes-to-escape-characters
func EscapeGanttLabel(text string) string {
	var result strings.Builder
	result.Grow(len(text))

	for _, r := range text {
		result.WriteString(escapeChar(r))
	}

	return result.String()
}

func escapeChar(r rune) string {
	switch r {
	case '\r', '\n', '\t':
		return " "
	case '[', ']', '{', '}', '<', '>', ',', ':':
		// Use Mermaid's entity code format: #[decimal];
		return fmt.Sprintf("#%d;", r)
	default:
		if unicode.IsControl(r) {
			return ""
		}
		return string(r)
	}
}
