package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
)

const maxDirectiveLabelLength = 80

func sanitizeDirective(raw string) string {
	replacer := strings.NewReplacer(
		"\r\n", " ",
		"\n", " ",
		"\r", " ",
		"\t", " ",
		"[", "(",
		"]", ")",
		"{", "(",
		"}", ")",
	)

	clean := replacer.Replace(raw)
	clean = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, clean)

	// Collapse repeated whitespace to a single space.
	clean = strings.Join(strings.Fields(clean), " ")
	clean = strings.TrimSpace(clean)

	if clean == "" {
		return "(unnamed command)"
	}

	runes := []rune(clean)
	if len(runes) > maxDirectiveLabelLength {
		truncationLimit := maxDirectiveLabelLength
		if truncationLimit > 3 {
			truncationLimit = truncationLimit - 3
		}
		clean = string(runes[:truncationLimit]) + "..."
	}

	return clean
}

var commandMetricsCmd = &cobra.Command{
	Use:   "command-metrics",
	Short: "Generates a command summary markdown report from agent metrics",
	Long:  `Generates a command summary markdown report from agent metrics`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := ""

		srcFile, err := cmd.Flags().GetString("src")
		if err != nil {
			return fmt.Errorf("src cannot be parsed: %w", err)
		}

		matches, err := filepath.Glob(srcFile)
		if err != nil || len(matches) == 0 {
			return fmt.Errorf("failed to find job log file: %w", err)
		}

		type CmdFinished struct {
			Event      string `json:"event"`
			Directive  string `json:"directive"`
			StartedAt  int64  `json:"started_at"`
			FinishedAt int64  `json:"finished_at"`
		}

		lines, err := os.ReadFile(matches[0])
		if err != nil {
			return fmt.Errorf("could not read job log: %w", err)
		}

		var flowNodes []CmdFinished
		for _, raw := range strings.Split(string(lines), "\n") {
			if strings.TrimSpace(raw) == "" {
				continue
			}
			var entry CmdFinished
			if err := json.Unmarshal([]byte(raw), &entry); err != nil {
				continue
			}
			if entry.Event == "cmd_finished" {
				flowNodes = append(flowNodes, entry)
			}
		}

		out += "## 🧭 Job Timeline\n\n```mermaid\ngantt\n    title Job Command Timeline\n    dateFormat X\n    axisFormat %X\n"
		for i, node := range flowNodes {
			duration := node.FinishedAt - node.StartedAt
			if duration < 1 {
				duration = 1
			}
			label := sanitizeDirective(node.Directive)
			out += fmt.Sprintf("    %s[%ds] :step%d, %d, %ds\n", label, duration, i, node.StartedAt, duration)
		}
		out += "```\n"

		f, err := os.OpenFile(args[0], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("failed to open output file: %w", err)
		}
		defer f.Close()

		if _, err := f.WriteString(out); err != nil {
			return fmt.Errorf("failed to append to output file: %w", err)
		}
		return nil
	},
}

func init() {
	tmpDir := os.TempDir()
	defaultSrc := filepath.Join(tmpDir, "job_log_*.json")

	commandMetricsCmd.Flags().String("src", defaultSrc, "source file to read system metrics from")
	rootCmd.AddCommand(commandMetricsCmd)
}
