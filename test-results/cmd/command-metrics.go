package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

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
			out += fmt.Sprintf("    %s[%ds] :step%d, %d, %ds\n", node.Directive, duration, i, node.StartedAt, duration)
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
	commandMetricsCmd.Flags().String("src", "/tmp/job_log_*.json", "source file to read system metrics from")
	rootCmd.AddCommand(commandMetricsCmd)
}
