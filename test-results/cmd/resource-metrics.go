package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var resourceMetricsCmd = &cobra.Command{
	Use:   "resource-metrics",
	Short: "Generates a resource utilization summary markdown report from agent metrics",
	Long:  `Generates a resource utilization summary markdown report from agent metrics`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		type ResourceMetric struct {
			Timestamp  string
			CPU        float64
			Memory     float64
			SystemDisk float64
			DockerDisk float64
		}

		srcFile, err := cmd.Flags().GetString("src")
		if err != nil {
			return fmt.Errorf("src cannot be parsed: %w", err)
		}

		file, err := os.Open(filepath.Clean(srcFile))
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		defer file.Close()

		metricLineRegex := regexp.MustCompile(`^(.*?) \|  cpu:(.*)%,  mem:\s*(.*)%,  system_disk:\s*(.*)%,  docker_disk:\s*(.*)%,(.*)$`)

		var metrics []ResourceMetric
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			matches := metricLineRegex.FindStringSubmatch(line)
			if len(matches) != 7 {
				continue
			}
			var m ResourceMetric
			m.Timestamp = matches[1]
			_, err := fmt.Sscanf(matches[2], "%f", &m.CPU)
			if err != nil {
				continue
			}
			_, err = fmt.Sscanf(matches[3], "%f", &m.Memory)
			if err != nil {
				continue
			}
			_, err = fmt.Sscanf(matches[4], "%f", &m.SystemDisk)
			if err != nil {
				continue
			}
			_, err = fmt.Sscanf(matches[5], "%f", &m.DockerDisk)
			if err != nil {
				continue
			}
			metrics = append(metrics, m)
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("error reading file: %w", err)
		}

		if len(metrics) == 0 {
			return fmt.Errorf("no valid data found")
		}

		step := 1
		if len(metrics) > 100 {
			step = len(metrics) / 100
		}

		var (
			xLabels          []string
			cpuSeries        []string
			memSeries        []string
			sysDiskSeries    []string
			dockerDiskSeries []string
		)

		layout := "Mon 02 Jan 2006 03:04:05 PM MST"
		startTime, err := time.Parse(layout, metrics[0].Timestamp)
		if err != nil {
			return fmt.Errorf("failed to parse start time: %w", err)
		}

		min := func(f1, f2 float64) float64 {
			if f1 < f2 {
				return f1
			}
			return f2
		}

		max := func(f1, f2 float64) float64 {
			if f1 > f2 {
				return f1
			}
			return f2
		}

		cpuMin, cpuMax := metrics[0].CPU, metrics[0].CPU
		memMin, memMax := metrics[0].Memory, metrics[0].Memory
		diskMin, diskMax := metrics[0].SystemDisk, metrics[0].SystemDisk
		dockerMin, dockerMax := metrics[0].DockerDisk, metrics[0].DockerDisk

		for i := 0; i < len(metrics); i += step {
			m := metrics[i]
			cpuMin = min(cpuMin, m.CPU)
			cpuMax = max(cpuMax, m.CPU)
			memMin = min(memMin, m.Memory)
			memMax = max(memMax, m.Memory)
			diskMin = min(diskMin, m.SystemDisk)
			diskMax = max(diskMax, m.SystemDisk)
			dockerMin = min(dockerMin, m.DockerDisk)
			dockerMax = max(dockerMax, m.DockerDisk)

			t, err := time.Parse(layout, m.Timestamp)
			if err != nil {
				xLabels = append(xLabels, "\"??:??\"")
			} else {
				duration := t.Sub(startTime)
				seconds := int(duration.Seconds())
				xLabels = append(xLabels, fmt.Sprintf("\"%02d:%02d\"", seconds/60, seconds%60))
			}
			cpuSeries = append(cpuSeries, fmt.Sprintf("%.2f", m.CPU))
			memSeries = append(memSeries, fmt.Sprintf("%.2f", m.Memory))
			sysDiskSeries = append(sysDiskSeries, fmt.Sprintf("%.2f", m.SystemDisk))
			dockerDiskSeries = append(dockerDiskSeries, fmt.Sprintf("%.2f", m.DockerDisk))
		}

		out := "## 🎯 System Metrics Summary\n\n"
		out += fmt.Sprintf("**Total datapoints:** `%d`  \n", len(metrics))
		out += fmt.Sprintf("**🕒 Time Range:** `%s` → `%s`  \n\n", metrics[0].Timestamp, metrics[len(metrics)-1].Timestamp)
		out += fmt.Sprintf("- **🔥 CPU:** `min: %.2f%%`, `max: %.2f%%`  \n", cpuMin, cpuMax)
		out += fmt.Sprintf("- **🧠 Memory:** `min: %.2f%%`, `max: %.2f%%`  \n", memMin, memMax)
		out += fmt.Sprintf("- **💽 System Disk:** `min: %.2f%%`, `max: %.2f%%`  \n", diskMin, diskMax)
		out += fmt.Sprintf("- **🐳 Docker Disk:** `min: %.2f%%`, `max: %.2f%%`\n\n", dockerMin, dockerMax)
		out += "---\n\n"

		out += "```mermaid\n"
		out += "xychart-beta\n"
		out += "title \"CPU Usage\"\n"
		out += fmt.Sprintf("x-axis [%s]\n", strings.Join(xLabels, ", "))
		out += "y-axis \"Usage (%)\"\n"
		out += fmt.Sprintf("line [%s]\n", strings.Join(cpuSeries, ", "))
		out += fmt.Sprintf("bar [%s]\n", strings.Join(cpuSeries, ", "))
		out += "```\n\n"

		out += "```mermaid\n"
		out += "xychart-beta\n"
		out += "title \"Memory Usage\"\n"
		out += fmt.Sprintf("x-axis [%s]\n", strings.Join(xLabels, ", "))
		out += "y-axis \"Usage (%)\"\n"
		out += fmt.Sprintf("line [%s]\n", strings.Join(memSeries, ", "))
		out += fmt.Sprintf("bar [%s]\n", strings.Join(memSeries, ", "))
		out += "```\n\n"

		out += "```mermaid\n"
		out += "xychart-beta\n"
		out += "title \"System Disk Usage\"\n"
		out += fmt.Sprintf("x-axis [%s]\n", strings.Join(xLabels, ", "))
		out += "y-axis \"Disk Usage (%)\"\n"
		out += fmt.Sprintf("line [%s]\n", strings.Join(sysDiskSeries, ", "))
		out += fmt.Sprintf("bar [%s]\n", strings.Join(sysDiskSeries, ", "))
		out += "```\n"

		out += "```mermaid\n"
		out += "xychart-beta\n"
		out += "title \"Docker Disk Usage\"\n"
		out += fmt.Sprintf("x-axis [%s]\n", strings.Join(xLabels, ", "))
		out += "y-axis \"Disk Usage (%)\"\n"
		out += fmt.Sprintf("line [%s]\n", strings.Join(dockerDiskSeries, ", "))
		out += fmt.Sprintf("bar [%s]\n", strings.Join(dockerDiskSeries, ", "))
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
	resourceMetricsCmd.Flags().String("src", "/tmp/system-metrics", "source file to read system metrics from")
	rootCmd.AddCommand(resourceMetricsCmd)
}
