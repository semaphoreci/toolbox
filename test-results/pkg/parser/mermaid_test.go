package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEscapeGanttLabel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no special characters",
			input:    "echo hello world",
			expected: "echo hello world",
		},
		{
			name:     "brackets and braces",
			input:    "if [ \"$VAR\" == 1 ] { echo ok }",
			expected: "if #91; \"$VAR\" == 1 #93; #123; echo ok #125;",
		},
		{
			name:     "angle brackets (heredoc)",
			input:    "cat <<EOF",
			expected: "cat #60;#60;EOF",
		},
		{
			name:     "commas and colons",
			input:    `{"key": "value", "num": 123}`,
			expected: `#123;"key"#58; "value"#44; "num"#58; 123#125;`,
		},
		{
			name:     "newlines and tabs become spaces",
			input:    "line1\nline2\tline3",
			expected: "line1 line2 line3",
		},
		{
			name:     "control characters removed",
			input:    "text\x00with\x01control",
			expected: "textwithcontrol",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, EscapeGanttLabel(tc.input))
		})
	}
}
