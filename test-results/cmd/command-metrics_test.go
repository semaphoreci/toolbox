package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSanitizeDirective(t *testing.T) {
	t.Parallel()

	longInput := strings.Repeat("a", 81)
	longExpected := strings.Repeat("a", 77) + "..."

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "trims whitespace and collapses",
			input:    "\t  echo   hello \nworld  ",
			expected: "echo hello world",
		},
		{
			name:     "replaces brackets with Mermaid entity codes",
			input:    "if [ \"$VAR\" == 1 ] { echo ok }",
			expected: "if #91; \"$VAR\" == 1 #93; #123; echo ok #125;",
		},
		{
			name:     "drops control characters",
			input:    "echo \x00hi",
			expected: "echo hi",
		},
		{
			name:     "truncates long labels",
			input:    longInput,
			expected: longExpected,
		},
		{
			name:     "provides fallback for empty values",
			input:    "   ",
			expected: "(unnamed command)",
		},
		{
			name:     "escapes angle brackets from heredoc with Mermaid entity codes",
			input:    "json_data=$(cat <<EOF\n{\n  \"key\": \"value\"\n}\nEOF\n)",
			expected: "json_data=$(cat #60;#60;EOF #123; \"key\"#58; \"value\" #125; EOF )",
		},
		{
			name:     "escapes commas and colons in JSON",
			input:    `json=$(cat <<EOF {"a": "b", "c": "d"} EOF)`,
			expected: `json=$(cat #60;#60;EOF #123;"a"#58; "b"#44; "c"#58; "d"#125; EOF)`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, sanitizeDirective(tc.input))
		})
	}
}

func TestCommandMetricsGeneratesMermaidTimeline(t *testing.T) {
	tempDir := t.TempDir()

	jobLogPath := filepath.Join(tempDir, "job_log_sample.json")
	var entries []map[string]any

	entries = append(entries, map[string]any{
		"event":       "cmd_finished",
		"directive":   "if [ \"$A\" == \"main\" ]; then\necho hi\nfi",
		"started_at":  int64(10),
		"finished_at": int64(12),
	})

	entries = append(entries, map[string]any{
		"event":       "cmd_finished",
		"directive":   "",
		"started_at":  int64(12),
		"finished_at": int64(12),
	})

	entries = append(entries, map[string]any{
		"event":       "cmd_finished",
		"directive":   strings.Repeat("b", 100),
		"started_at":  int64(20),
		"finished_at": int64(25),
	})

	entries = append(entries, map[string]any{
		"event":       "cmd_started",
		"directive":   "ignored",
		"started_at":  int64(30),
		"finished_at": int64(40),
	})

	var builder strings.Builder
	for _, entry := range entries {
		payload, err := json.Marshal(entry)
		require.NoError(t, err)
		builder.Write(payload)
		builder.WriteString("\n")
	}

	err := os.WriteFile(jobLogPath, []byte(builder.String()), 0600)
	require.NoError(t, err)

	outputPath := filepath.Join(tempDir, "report.md")
	err = os.WriteFile(outputPath, []byte("# Existing content\n"), 0600)
	require.NoError(t, err)

	srcFlag := commandMetricsCmd.Flags().Lookup("src")
	require.NotNil(t, srcFlag)

	defer func() {
		_ = commandMetricsCmd.Flags().Set("src", srcFlag.DefValue)
	}()

	err = commandMetricsCmd.Flags().Set("src", jobLogPath)
	require.NoError(t, err)

	err = commandMetricsCmd.RunE(commandMetricsCmd, []string{outputPath})
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	text := string(content)

	require.Contains(t, text, "# Existing content")
	require.Contains(t, text, "## 🧭 Job Timeline")

	require.Contains(t, text, `if #91; "$A" == "main" #93;; then echo hi fi[2s] :step0, 10, 2s`)
	require.Contains(t, text, `(unnamed command)[1s] :step1, 12, 1s`)
	require.Contains(t, text, strings.Repeat("b", 77)+`...[5s] :step2, 20, 5s`)
}
