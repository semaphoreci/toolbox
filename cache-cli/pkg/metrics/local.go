//revive:disable-next-line:var-naming
package metrics

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type LocalMetricsManager struct {
	ToolboxMetricsPath string
}

func NewLocalMetricsBackend() (*LocalMetricsManager, error) {
	basePath := "/tmp"
	if runtime.GOOS == "windows" {
		basePath = os.TempDir()
	}

	return &LocalMetricsManager{
		ToolboxMetricsPath: filepath.Join(basePath, "toolbox_metrics"),
	}, nil
}

func (b *LocalMetricsManager) Enabled() bool {
	return os.Getenv("SEMAPHORE_TOOLBOX_METRICS_ENABLED") == "true"
}

func (b *LocalMetricsManager) LogEvent(event CacheEvent) error {
	if !b.Enabled() {
		return nil
	}

	return publishEventToFile(b.ToolboxMetricsPath, event)
}

func publishEventToFile(file string, event CacheEvent) error {
	server := event.Server
	if server == "" {
		server = CacheServerIP()
	}

	user := event.User
	if user == "" {
		user = CacheUsername()
	}

	command := event.Command
	if command == "" {
		command = CommandRestore
	}

	corruptValue := 0
	if event.Corrupt {
		corruptValue = 1
	}

	durationMs := event.Duration.Milliseconds()
	if durationMs < 0 {
		durationMs = 0
	}

	// #nosec
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	line := fmt.Sprintf(
		"%s,server=%s,user=%s,command=%s,corrupt=%d size=%d,duration=%d\n",
		MeasurementName,
		escapeTagValue(server),
		escapeTagValue(user),
		escapeTagValue(command),
		corruptValue,
		event.SizeBytes,
		durationMs,
	)

	_, err = f.WriteString(line)
	if err != nil {
		_ = f.Close()
		return err
	}

	return f.Close()
}

func escapeTagValue(value string) string {
	if value == "" {
		return "unknown"
	}

	replacer := strings.NewReplacer(",", "\\,", " ", "\\ ", "=", "\\=")
	return replacer.Replace(value)
}
