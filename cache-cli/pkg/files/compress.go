package files

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func Compress(key, path string) (string, error) {
	epochNanos := time.Now().Nanosecond()
	tempFileName := filepath.Join(os.TempDir(), fmt.Sprintf("%s-%d", key, epochNanos))

	cmd := compressionCommand(path, tempFileName)
	_, err := cmd.Output()
	if err != nil {
		return tempFileName, err
	}

	return tempFileName, nil
}

func compressionCommand(path, tempFileName string) *exec.Cmd {
	if filepath.IsAbs(path) {
		return exec.Command("tar", "czPf", tempFileName, path)
	}

	return exec.Command("tar", "czf", tempFileName, path)
}
